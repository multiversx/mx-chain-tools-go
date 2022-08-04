package main

import (
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/interactors"
	"github.com/ElrondNetwork/elrond-tools-go/hdKeysGenerator/common"
)

const (
	fixedTaskSize = 8192
)

type generatorTask struct {
	useAccountIndex bool
	firstIndex      int
	lastIndex       int
	constraints     constraints
}

func createTasks(args argsCreateTasks) ([]generatorTask, int) {
	tasks := make([]generatorTask, 0, args.numTasks)
	slidingIndex := args.startIndex

	for taskIndex := 0; taskIndex < args.numTasks; taskIndex++ {
		task := generatorTask{
			useAccountIndex: args.useAccountIndex,
			firstIndex:      slidingIndex,
			lastIndex:       slidingIndex + fixedTaskSize,
			constraints:     args.constraints,
		}

		tasks = append(tasks, task)
		slidingIndex += fixedTaskSize
	}

	return tasks, slidingIndex
}

func (task *generatorTask) doGenerateKeys(mnemonic data.Mnemonic) ([]common.GeneratedKey, error) {
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

		isGoodKey := task.constraints.areSatisfiedByPublicKey(addressHandler.AddressBytes())
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
