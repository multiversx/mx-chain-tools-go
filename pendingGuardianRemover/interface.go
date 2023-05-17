package pendingGuardianRemover

import (
	"context"

	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
)

// TxBuilder defines the component able to build and sign a guarded transaction
type TxBuilder interface {
	ApplyUserSignatureAndGenerateTx(cryptoHolder core.CryptoComponentsHolder, arg data.ArgCreateTransaction) (*data.Transaction, error)
	ApplyGuardianSignature(cryptoHolderGuardian core.CryptoComponentsHolder, tx *data.Transaction) error
	IsInterfaceNil() bool
}

// TxsSaver defines the component able to save transactions to a file
type TxsSaver interface {
	Save(txs []*data.Transaction) error
	IsInterfaceNil() bool
}

// HttpClient defines the behavior of http client able to make http requests
type HttpClient interface {
	GetHTTP(ctx context.Context, endpoint string) ([]byte, int, error)
	PostHTTP(ctx context.Context, endpoint string, data []byte) ([]byte, int, error)
	IsInterfaceNil() bool
}

// HttpClientWrapper defines the behavior of wrapper over HttpClient
type HttpClientWrapper interface {
	GetAccount(ctx context.Context, address string) (*data.Account, error)
	GetGuardianData(ctx context.Context, address string) (*api.GuardianData, error)
	SendTransaction(ctx context.Context, txBuff []byte) (string, error)
	IsInterfaceNil() bool
}
