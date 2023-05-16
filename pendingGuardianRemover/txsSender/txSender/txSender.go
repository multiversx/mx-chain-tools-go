package txSender

import (
	"context"
	"encoding/json"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-tools-go/pendingGuardianRemover"
	"github.com/multiversx/mx-sdk-go/data"
)

var log = logger.GetOrCreate("sender")

type txSender struct {
	httpClient        pendingGuardianRemover.HttpClientWrapper
	delayBetweenSends time.Duration
	txsMap            map[uint64]*data.Transaction
	senderAddress     string
	cancel            func()
}

// NewTxSender returns a new txSender and starts the internal loop
func NewTxSender(
	httpClient pendingGuardianRemover.HttpClientWrapper,
	delayBetweenSends time.Duration,
	txsMap map[uint64]*data.Transaction,
) (*txSender, error) {

	// todo add nil checks

	senderAddr := ""
	for _, tx := range txsMap {
		senderAddr = tx.SndAddr
		break
	}

	sender := &txSender{
		httpClient:        httpClient,
		delayBetweenSends: delayBetweenSends,
		txsMap:            txsMap,
		senderAddress:     senderAddr,
	}

	var ctx context.Context
	ctx, sender.cancel = context.WithCancel(context.Background())
	go sender.sendTransactions(ctx)

	return sender, nil
}

func (sender *txSender) sendTransactions(ctx context.Context) {
	timer := time.NewTimer(sender.delayBetweenSends)
	defer timer.Stop()

	for {
		timer.Reset(sender.delayBetweenSends)

		select {
		case <-timer.C:
			sender.sendNextTransaction(ctx)
		case <-ctx.Done():
			log.Debug("closing txs sender main loop...")
			return
		}
	}
}

func (sender *txSender) sendNextTransaction(ctx context.Context) {
	if !sender.hasPendingGuardian(ctx) {
		log.Debug("account does not have pending guardian, skipping SetGuardian transaction...")
		return
	}

	currentNonce, err := sender.getAccountNonce(ctx)
	if err != nil {
		log.Warn("failed to get account", "address", sender.senderAddress, "error", err.Error())
		return
	}

	tx, found := sender.txsMap[currentNonce]
	if !found {
		log.Warn("account nonce is missing from generated txs", "nonce", currentNonce)
		return
	}

	if tx.SndAddr != sender.senderAddress {
		log.Warn("tx sender does not match the initial sender",
			"nonce", currentNonce,
			"initial sender", sender.senderAddress,
			"tx sender", tx.SndAddr)
		return
	}

	txBuff, err := json.Marshal(tx)
	if err != nil {
		log.Warn("failed to marshal tx", "nonce", currentNonce, "error", err.Error())
		return
	}

	hash, err := sender.httpClient.SendTransaction(ctx, txBuff)
	if err != nil {
		log.Warn("failed to send tx", "nonce", currentNonce, "error", err.Error())
		return
	}

	log.Debug("transaction sent", "hash", hash)
}

func (sender *txSender) hasPendingGuardian(ctx context.Context) bool {
	guardianData, err := sender.httpClient.GetGuardianData(ctx, sender.senderAddress)
	if err != nil {
		return true // force tx in case of error
	}

	if check.IfNilReflect(guardianData) {
		log.Warn("received nil guardian data")
		return false
	}

	if check.IfNilReflect(guardianData.PendingGuardian) {
		return false
	}

	return true
}

func (sender *txSender) getAccountNonce(ctx context.Context) (uint64, error) {
	account, err := sender.httpClient.GetAccount(ctx, sender.senderAddress)
	if err != nil {
		return 0, err
	}

	return account.Nonce, nil
}

// Close closes the internal loop
func (sender *txSender) Close() error {
	sender.cancel()

	return nil
}
