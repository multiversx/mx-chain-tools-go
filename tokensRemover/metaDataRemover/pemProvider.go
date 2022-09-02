package main

import (
	"github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/interactors"
)

type pkAddress struct {
	privateKey []byte
	address    core.AddressHandler
}

type pemDataProvider struct {
}

func (pdp *pemDataProvider) getPrivateKeyAndAddress(pemFile string) (*pkAddress, error) {
	w := interactors.NewWallet()
	privateKey, err := w.LoadPrivateKeyFromPemFile(pemFile)
	if err != nil {
		return nil, err
	}

	address, err := w.GetAddressFromPrivateKey(privateKey)
	if err != nil {
		return nil, err

	}

	return &pkAddress{
		privateKey: privateKey,
		address:    address,
	}, nil
}
