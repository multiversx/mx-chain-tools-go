package mocks

import (
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
)

type TransactionInteractorStub struct {
	ApplySignatureAndGenerateTxCalled func(skBytes []byte, arg data.ArgCreateTransaction) (*data.Transaction, error)
}

func (tis *TransactionInteractorStub) ApplySignatureAndGenerateTx(skBytes []byte, arg data.ArgCreateTransaction) (*data.Transaction, error) {
	if tis.ApplySignatureAndGenerateTxCalled != nil {
		return tis.ApplySignatureAndGenerateTxCalled(skBytes, arg)
	}

	return nil, nil
}
