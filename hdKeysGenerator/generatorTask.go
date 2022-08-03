package main

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/interactors"
	"github.com/ElrondNetwork/elrond-tools-go/hdKeysGenerator/common"
)

type generatorTask struct {
	numTasks        int
	taskIndex       int
	useAccountIndex bool
	startIndex      int
	numKeys         int
}

func createTasks(cliFlags parsedCliFlags) []generatorTask {
	tasks := make([]generatorTask, 0, cliFlags.numTasks)

	for taskIndex := 0; taskIndex < cliFlags.numTasks; taskIndex++ {
		task := generatorTask{
			numTasks:        cliFlags.numTasks,
			taskIndex:       taskIndex,
			useAccountIndex: cliFlags.useAccountIndex,
			startIndex:      cliFlags.startIndex,
			numKeys:         int(cliFlags.numKeys)/cliFlags.numTasks + 1,
		}

		tasks = append(tasks, task)
	}

	return tasks
}

func (task *generatorTask) doGenerateKeys(mnemonic data.Mnemonic, constraints *constraints) ([]common.GeneratedKey, error) {
	wallet := interactors.NewWallet()
	seed := wallet.CreateSeedFromMnemonic(mnemonic)
	goodKeys := make([]common.GeneratedKey, 0, task.numKeys)

	accountIndex := 0
	addressIndex := 0
	var changingIndex *int

	if task.useAccountIndex {
		changingIndex = &accountIndex
	} else {
		changingIndex = &addressIndex
	}

	for len(goodKeys) < task.numKeys {
		if *changingIndex%task.numTasks != task.taskIndex {
			*changingIndex++
			continue
		}

		secretKey := wallet.GetPrivateKeyFromSeed(seed, uint32(accountIndex), uint32(addressIndex))
		addressHandler, err := wallet.GetAddressFromPrivateKey(secretKey)
		if err != nil {
			return nil, err
		}

		isGoodKey := constraints.areSatisfiedByPublicKey(addressHandler.AddressBytes())
		if isGoodKey {
			goodKeys = append(goodKeys, common.GeneratedKey{
				Index:     *changingIndex,
				SecretKey: secretKey,
				PublicKey: addressHandler.AddressBytes(),
				Address:   addressHandler.AddressAsBech32String(),
			})

			task.logProgress(len(goodKeys))
		}

		*changingIndex++
	}

	return goodKeys, nil
}

func (task *generatorTask) logProgress(numGenerated int) {
	if numGenerated%100 == 0 && numGenerated > 0 {
		progress := int(float64(numGenerated) / float64(task.numKeys) * 100)
		log.Info("generating keys...", "task", task.taskIndex, "progress", fmt.Sprintf("%d %%", progress))
	}
}
