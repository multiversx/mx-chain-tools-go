package txGenerator

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-tools-go/pendingGuardianRemover"
	"github.com/multiversx/mx-chain-tools-go/pendingGuardianRemover/txsCreator/config"
	"github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
)

type txGenerator struct {
	txBuilder            pendingGuardianRemover.TxBuilder
	userCryptoHolder     core.CryptoComponentsHolder
	guardianCryptoHolder core.CryptoComponentsHolder
	httpClient           pendingGuardianRemover.HttpClientWrapper
	config               config.TxGeneratorConfig
}

// NewTxGenerator creates a new tx generator
func NewTxGenerator(
	txBuilder pendingGuardianRemover.TxBuilder,
	userCryptoHolder core.CryptoComponentsHolder,
	guardianCryptoHolder core.CryptoComponentsHolder,
	httpClient pendingGuardianRemover.HttpClientWrapper,
	config config.TxGeneratorConfig,
) (*txGenerator, error) {

	if check.IfNil(txBuilder) {
		return nil, pendingGuardianRemover.ErrNilTXBuilder
	}
	if check.IfNil(userCryptoHolder) {
		return nil, fmt.Errorf("%w for user", pendingGuardianRemover.ErrNilCryptoHolder)
	}
	if check.IfNil(guardianCryptoHolder) {
		return nil, fmt.Errorf("%w for guardian", pendingGuardianRemover.ErrNilCryptoHolder)
	}
	if check.IfNil(httpClient) {
		return nil, pendingGuardianRemover.ErrNilHttpClient
	}

	return &txGenerator{
		txBuilder:            txBuilder,
		userCryptoHolder:     userCryptoHolder,
		guardianCryptoHolder: guardianCryptoHolder,
		httpClient:           httpClient,
		config:               config,
	}, nil
}

// GenerateTxs generates transactions and saves them into a json file
func (txGen *txGenerator) GenerateTxs() ([]*data.Transaction, error) {
	startingNonce, err := txGen.getAccountNonce(context.Background())
	if err != nil {
		return nil, err
	}

	txs := make([]*data.Transaction, 0, txGen.config.NumOfTransactions)
	for i := uint64(0); i < txGen.config.NumOfTransactions; i++ {
		newTx, errCreate := txGen.createOneTx(startingNonce + i)
		if errCreate != nil {
			return nil, errCreate
		}

		txs = append(txs, newTx)
	}

	return txs, nil
}

func (txGen *txGenerator) createOneTx(nonce uint64) (*data.Transaction, error) {
	txArgs := data.ArgCreateTransaction{
		Nonce:        nonce,
		Value:        "0",
		RcvAddr:      txGen.userCryptoHolder.GetBech32(),
		SndAddr:      txGen.userCryptoHolder.GetBech32(),
		GasPrice:     txGen.config.GasPrice,
		GasLimit:     txGen.config.GasLimit,
		Data:         txGen.getTxData(),
		ChainID:      txGen.config.ChainID,
		Version:      2,
		Options:      2,
		GuardianAddr: txGen.guardianCryptoHolder.GetBech32(),
	}

	tx, err := txGen.txBuilder.ApplyUserSignatureAndGenerateTx(txGen.userCryptoHolder, txArgs)
	if err != nil {
		return nil, err
	}

	err = txGen.txBuilder.ApplyGuardianSignature(txGen.guardianCryptoHolder, tx)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (txGen *txGenerator) getTxData() []byte {
	guardianPK := txGen.guardianCryptoHolder.GetPublicKey()
	guardianAddr, _ := guardianPK.ToByteArray()
	return []byte("SetGuardian@" + hex.EncodeToString(guardianAddr) + "@" + hex.EncodeToString([]byte(txGen.config.ServiceUID)))
}

func (txGen *txGenerator) getAccountNonce(ctx context.Context) (uint64, error) {
	account, err := txGen.httpClient.GetAccount(ctx, txGen.userCryptoHolder.GetBech32())
	if err != nil {
		return 0, err
	}

	return account.Nonce, nil
}
