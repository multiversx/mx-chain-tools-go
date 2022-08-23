package main

import (
	"context"
	"fmt"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/interactors"
)

func sendTxs(
	pemFile string,
	proxy proxyProvider,
	txInteractor transactionInteractor,
	txsData [][]byte,
	bulkSize int,
) error {
	privateKey, address, err := getPrivateKeyAndAddress(pemFile)
	if err != nil {
		return err
	}

	transactionArguments, err := getDefaultTxsArgs(proxy, address)
	if err != nil {
		return err
	}

	log.Info("starting to send", "num of txs", len(txsData))
	currBulkSize := 0
	totalSentTxs := 0
	for _, txData := range txsData {
		transactionArguments.Nonce++
		transactionArguments.Data = txData
		tx, err := txInteractor.ApplySignatureAndGenerateTx(privateKey, *transactionArguments)
		if err != nil {
			return err
		}

		txInteractor.AddTransaction(tx)
		currBulkSize++
		if currBulkSize < bulkSize {
			continue
		}

		err = sendMultipleTxs(txInteractor, currBulkSize)
		if err != nil {
			return err
		}

		currBulkSize = 0
		totalSentTxs += bulkSize
	}

	remainingTxs := len(txsData) - totalSentTxs
	if remainingTxs > 0 {
		err = sendMultipleTxs(txInteractor, remainingTxs)
		if err != nil {
			return err
		}

		totalSentTxs += remainingTxs
	}

	if totalSentTxs != len(txsData) {
		return fmt.Errorf("did not send all txs; sent %d, should have sent %d", totalSentTxs, len(txsData))
	}

	log.Info("sent all txs")
	return nil
}

func getPrivateKeyAndAddress(pemFile string) ([]byte, core.AddressHandler, error) {
	w := interactors.NewWallet()
	privateKey, err := w.LoadPrivateKeyFromPemFile(pemFile)
	if err != nil {
		return nil, nil, err
	}

	address, err := w.GetAddressFromPrivateKey(privateKey)
	if err != nil {
		return nil, nil, err

	}

	return privateKey, address, nil
}

func getDefaultTxsArgs(proxy proxyProvider, address core.AddressHandler) (*data.ArgCreateTransaction, error) {
	netConfigs, err := proxy.GetNetworkConfig(context.Background())
	if err != nil {
		return nil, err

	}

	transactionArguments, err := proxy.GetDefaultTransactionArguments(context.Background(), address, netConfigs)
	if err != nil {
		return nil, err
	}

	transactionArguments.RcvAddr = address.AddressAsBech32String() // send to self
	transactionArguments.Value = "0"

	return &transactionArguments, nil
}

func sendMultipleTxs(txInteractor transactionInteractor, numTxs int) error {
	hashes, err := txInteractor.SendTransactionsAsBunch(context.Background(), numTxs)
	if err != nil {
		return err
	}

	if len(hashes) != numTxs {
		return fmt.Errorf("failed to send all txs; sent: %d, should have sent: %d, sent tx hashes:%v ", len(hashes), numTxs, hashes)
	}

	log.Info("sent", "num of txs", numTxs, "hashes", hashes)
	return nil
}
