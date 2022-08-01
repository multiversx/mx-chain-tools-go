package main

import (
	"encoding/hex"
	"fmt"

	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/interactors"
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

func (g *generatorTask) doGenerateKeys() ([]generatedKey, error) {
	wallet := interactors.NewWallet()

	numGeneratedTotal := 0
	numGeneratedWithConstraints := 0
	generatedKeys := make([]generatedKey, 0, g.numKeys)

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

		privateKey := wallet.GetPrivateKeyFromMnemonic(g.mnemonic, uint32(accountIndex), uint32(addressIndex))
		addressHandler, err := wallet.GetAddressFromPrivateKey(privateKey)
		if err != nil {
			return nil, err
		}

		if g.constraints.areSatisfiedByPublicKey(addressHandler.AddressBytes()) {
			generatedKeys = append(generatedKeys, generatedKey{
				Index:     *changingIndex,
				SecretKey: hex.EncodeToString(privateKey),
				PublicKey: hex.EncodeToString(addressHandler.AddressBytes()),
				Address:   addressHandler.AddressAsBech32String(),
			})

			fmt.Println(g.taskIndex, accountIndex, addressIndex, hex.EncodeToString(addressHandler.AddressBytes()), addressHandler.AddressAsBech32String())
			numGeneratedWithConstraints++
		}

		numGeneratedTotal++
	}

	return generatedKeys, nil
}
