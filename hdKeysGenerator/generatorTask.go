package main

import (
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/interactors"
	"github.com/ElrondNetwork/elrond-tools-go/hdKeysGenerator/common"
)

type generatorTask struct {
	useAccountIndex bool
	firstIndex      int
	lastIndex       int
}

func createTasks(numTasks int, startingIndex int, useAccountIndex bool) ([]generatorTask, int) {
	tasks := make([]generatorTask, 0, numTasks)
	slidingIndex := startingIndex

	for taskIndex := 0; taskIndex < numTasks; taskIndex++ {
		task := generatorTask{
			useAccountIndex: useAccountIndex,
			firstIndex:      slidingIndex,
			lastIndex:       slidingIndex + defaultTaskSize,
		}

		tasks = append(tasks, task)

		slidingIndex += defaultTaskSize
	}

	return tasks, slidingIndex
}

func (task *generatorTask) doGenerateKeys(mnemonic data.Mnemonic, constraints *constraints) ([]common.GeneratedKey, error) {
	wallet := interactors.NewWallet()
	seed := wallet.CreateSeedFromMnemonic(mnemonic)
	goodKeys := make([]common.GeneratedKey, 0)

	accountIndex := 0
	addressIndex := 0
	var changingIndex *int

	if task.useAccountIndex {
		changingIndex = &accountIndex
	} else {
		changingIndex = &addressIndex
	}

	for i := task.firstIndex; i < task.lastIndex; i++ {
		*changingIndex = i

		secretKey := wallet.GetPrivateKeyFromSeed(seed, uint32(accountIndex), uint32(addressIndex))
		addressHandler, err := wallet.GetAddressFromPrivateKey(secretKey)
		if err != nil {
			return nil, err
		}

		isGoodKey := constraints.areSatisfiedByPublicKey(addressHandler.AddressBytes())
		if isGoodKey {
			goodKeys = append(goodKeys, common.GeneratedKey{
				AccountIndex: accountIndex,
				AddressIndex: addressIndex,
				SecretKey:    secretKey,
				PublicKey:    addressHandler.AddressBytes(),
				Address:      addressHandler.AddressAsBech32String(),
			})
		}
	}

	return goodKeys, nil
}
