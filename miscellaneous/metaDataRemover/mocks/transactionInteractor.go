package mocks

import (
	"context"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
)

type TransactionInteractorStub struct {
	ApplySignatureAndGenerateTxCalled func(skBytes []byte, arg data.ArgCreateTransaction) (*data.Transaction, error)
	AddTransactionCalled              func(tx *data.Transaction)
	SendTransactionsAsBunchCalled     func(ctx context.Context, bunchSize int) ([]string, error)
}

func (tis *TransactionInteractorStub) ApplySignatureAndGenerateTx(skBytes []byte, arg data.ArgCreateTransaction) (*data.Transaction, error) {
	if tis.ApplySignatureAndGenerateTxCalled != nil {
		return tis.ApplySignatureAndGenerateTxCalled(skBytes, arg)
	}

	return nil, nil
}

func (tis *TransactionInteractorStub) AddTransaction(tx *data.Transaction) {
	if tis.AddTransactionCalled != nil {
		tis.AddTransactionCalled(tx)
	}
}

func (tis *TransactionInteractorStub) SendTransactionsAsBunch(ctx context.Context, bunchSize int) ([]string, error) {
	if tis.SendTransactionsAsBunchCalled != nil {
		return tis.SendTransactionsAsBunchCalled(ctx, bunchSize)
	}
	return nil, nil
}
