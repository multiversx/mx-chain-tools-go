package main

import (
	"encoding/hex"
	"fmt"

	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/interactors"
	"github.com/ElrondNetwork/elrond-tools-go/hdKeysGenerator/common"
)

type generatorTask struct {
	numTasks        int
	taskIndex       int
	mnemonic        data.Mnemonic
	constraints     *constraints
	useAccountIndex bool
	startIndex      int
	numKeys         int
}

func (g *generatorTask) doGenerateKeys() ([]common.GeneratedKey, error) {
	wallet := interactors.NewWallet()
	seed := wallet.CreateSeedFromMnemonic(g.mnemonic)

	numGeneratedTotal := 0
	numGeneratedWithConstraints := 0
	generatedKeys := make([]common.GeneratedKey, 0, g.numKeys)

	accountIndex := 0
	addressIndex := 0
	var changingIndex *int

	if g.useAccountIndex {
		changingIndex = &accountIndex
	} else {
		changingIndex = &addressIndex
	}

	*changingIndex = g.startIndex - 1

	for numGeneratedWithConstraints < g.numKeys {
		*changingIndex++

		if *changingIndex%g.numTasks != g.taskIndex {
			continue
		}

		secretKey := wallet.GetPrivateKeyFromSeed(seed, uint32(accountIndex), uint32(addressIndex))
		addressHandler, err := wallet.GetAddressFromPrivateKey(secretKey)
		if err != nil {
			return nil, err
		}

		if g.constraints.areSatisfiedByPublicKey(addressHandler.AddressBytes()) {
			generatedKeys = append(generatedKeys, common.GeneratedKey{
				Index:     *changingIndex,
				SecretKey: secretKey,
				PublicKey: addressHandler.AddressBytes(),
				Address:   addressHandler.AddressAsBech32String(),
			})

			fmt.Println(g.taskIndex, accountIndex, addressIndex, hex.EncodeToString(addressHandler.AddressBytes()), addressHandler.AddressAsBech32String())
			numGeneratedWithConstraints++
		}

		numGeneratedTotal++
	}

	return generatedKeys, nil
}
