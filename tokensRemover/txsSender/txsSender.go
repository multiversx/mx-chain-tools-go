package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
)

type txsSender struct {
	proxy proxyProvider
}

func (ts *txsSender) send(txs []*data.Transaction) error {
	cfg, err := ts.proxy.GetNetworkConfig(context.Background())
	if err != nil {
		return err
	}

	roundDuration := cfg.RoundDuration
	log.Info("found", "round duration", roundDuration, "num of txs to send", len(txs))
	for idx, tx := range txs {
		err = ts.waitForNonceIncremental(tx.SndAddr, tx.Nonce, 60)
		if err != nil {
			log.Error("failed to send tx", "tx index", idx, "error", err)
			return err
		}

		hash, err := ts.proxy.SendTransaction(context.Background(), tx)
		if err != nil {
			log.Error("failed to send tx", "tx index", idx, "error", err)
			return err
		}

		time.Sleep(time.Millisecond * time.Duration(roundDuration))
		log.Info("sent transaction", "tx hash", hash, "current tx index:", idx, "remaining num of txs", len(txs)-idx-1)
	}

	return nil
}

func (ts *txsSender) waitForNonceIncremental(address string, expectedNonce uint64, waitTime uint64) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for numRetrials := uint64(0); numRetrials < waitTime; <-ticker.C {
		accountNonce, errNonce := ts.getNonce(address)
		log.LogIfError(errNonce)

		if accountNonce == expectedNonce {
			return nil
		}

		numRetrials++
	}

	return fmt.Errorf("max retrials exceeded limit of %d seconds", waitTime)
}

func (ts *txsSender) getNonce(address string) (uint64, error) {
	addr, err := data.NewAddressFromBech32String(address)
	if err != nil {
		return 0, err
	}

	account, err := ts.proxy.GetAccount(context.Background(), addr)
	if err != nil {
		return 0, err
	}

	return account.Nonce, nil
}
