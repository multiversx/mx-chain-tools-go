package main

import (
	"context"

	"github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
)

type proxyProvider interface {
	GetNetworkConfig(ctx context.Context) (*data.NetworkConfig, error)
	GetDefaultTransactionArguments(
		ctx context.Context,
		address core.AddressHandler,
		networkConfigs *data.NetworkConfig,
	) (data.ArgCreateTransaction, error)
}

type transactionInteractor interface {
	ApplySignatureAndGenerateTx(skBytes []byte, arg data.ArgCreateTransaction) (*data.Transaction, error)
}

type pemProvider interface {
	getPrivateKeyAndAddress(pemFile string) (*pkAddress, error)
}
