package mocks

import (
	"github.com/multiversx/mx-sdk-go/data"
)

// TransactionInteractorStub -
type TransactionInteractorStub struct {
	ApplySignatureAndGenerateTxCalled func(skBytes []byte, arg data.ArgCreateTransaction) (*data.Transaction, error)
}

// ApplySignatureAndGenerateTx -
func (tis *TransactionInteractorStub) ApplySignatureAndGenerateTx(skBytes []byte, arg data.ArgCreateTransaction) (*data.Transaction, error) {
	if tis.ApplySignatureAndGenerateTxCalled != nil {
		return tis.ApplySignatureAndGenerateTxCalled(skBytes, arg)
	}

	return nil, nil
}
