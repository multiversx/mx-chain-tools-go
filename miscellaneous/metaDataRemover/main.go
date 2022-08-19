package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/blockchain"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/builders"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/interactors"
	"github.com/ElrondNetwork/elrond-tools-go/trieTools/trieToolsCommon"
	"github.com/urfave/cli"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	logFilePrefix  = "meta-data-remover"
	intervalsPerTx = 100
	txsBulkSize    = 100
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
	AddTransaction(tx *data.Transaction)
	SendTransactionsAsBunch(ctx context.Context, bunchSize int) ([]string, error)
}

type interval struct {
	start uint64
	end   uint64
}

func main() {
	app := cli.NewApp()
	app.Name = "Tokens exporter CLI app"
	app.Usage = "This is the entry point for the tool that deletes tokens meta-data"
	app.Flags = getFlags()
	app.Authors = []cli.Author{
		{
			Name:  "The Elrond Team",
			Email: "contact@elrond.com",
		},
	}

	app.Action = func(c *cli.Context) error {
		return startProcess(c)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
		return
	}
}

func startProcess(c *cli.Context) error {
	flagsConfig := getFlagsConfig(c)

	_, errLogger := trieToolsCommon.AttachFileLogger(log, logFilePrefix, flagsConfig.ContextFlagsConfig)
	if errLogger != nil {
		return errLogger
	}

	log.Info("sanity checks...")

	err := logger.SetLogLevel(flagsConfig.LogLevel)
	if err != nil {
		return err
	}

	log.Info("starting processing", "pid", os.Getpid())

	tokens, err := readInput(flagsConfig.Tokens)
	if err != nil {
		return err
	}

	tokensSorted, err := sortTokensIDByNonce(tokens)
	if err != nil {
		return err
	}

	tokensIntervals := groupTokensByIntervals(tokensSorted)
	txsData, err := createTxsData(tokensIntervals, intervalsPerTx)
	if err != nil {
		return err
	}

	args := blockchain.ArgsElrondProxy{
		ProxyURL:            "https://gateway.elrond.com",
		Client:              nil,
		SameScState:         false,
		ShouldBeSynced:      false,
		FinalityCheck:       false,
		CacheExpirationTime: time.Minute,
		EntityType:          core.Proxy,
	}

	proxy, err := blockchain.NewElrondProxy(args)
	if err != nil {
		return err
	}

	txBuilder, err := builders.NewTxBuilder(blockchain.NewTxSigner())
	if err != nil {
		return err
	}

	ti, err := interactors.NewTransactionInteractor(proxy, txBuilder)
	if err != nil {
		return err
	}

	err = sendTxs(flagsConfig.Pem, proxy, ti, txsData, txsBulkSize)
	if err != nil {
		return err
	}

	return nil
}

func readInput(tokensFile string) (map[string]struct{}, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	fullPath := filepath.Join(workingDir, tokensFile)
	jsonFile, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}

	bytesFromJson, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	tokens := make(map[string]struct{})
	err = json.Unmarshal(bytesFromJson, &tokens)
	if err != nil {
		return nil, err
	}

	log.Info("read from input", "file", tokensFile, "num of tokens", len(tokens))
	return tokens, nil
}

func sortTokensIDByNonce(tokens map[string]struct{}) (map[string][]uint64, error) {
	ret := make(map[string][]uint64)
	for token := range tokens {
		splits := strings.Split(token, "-")
		if len(splits) != 3 {
			return nil, fmt.Errorf("found invalid format in token = %s; expected [ticker-randSequence-nonce]", token)
		}

		tokenID := splits[0] + "-" + splits[1] // ticker-randSequence
		nonceBI := big.NewInt(0)
		nonceBI.SetString(splits[2], 16)

		ret[tokenID] = append(ret[tokenID], nonceBI.Uint64())
	}

	log.Info("found", "num of tokensID", len(ret))
	for _, nonces := range ret {
		sort.SliceStable(nonces, func(i, j int) bool {
			return nonces[i] < nonces[j]
		})
	}

	return ret, nil
}

func groupTokensByIntervals(tokens map[string][]uint64) map[string][]*interval {
	ret := make(map[string][]*interval)

	for token, nonces := range tokens {
		numNonces := len(nonces)
		for idx := 0; idx < numNonces; idx++ {
			nonce := nonces[idx]
			if idx+1 >= numNonces {
				ret[token] = append(ret[token], &interval{
					start: nonce,
					end:   nonce,
				})
				break
			}

			currInterval := &interval{start: nonce}
			numConsecutiveNonces := uint64(0)
			for idx < numNonces-1 {
				currNonce := nonces[idx]
				nextNonce := nonces[idx+1]
				if nextNonce-currNonce > 1 {
					break
				}

				numConsecutiveNonces++
				idx++
			}

			currInterval.end = currInterval.start + numConsecutiveNonces
			ret[token] = append(ret[token], currInterval)
		}
	}

	for token, intervals := range ret {
		log.Info("found", "tokenID", token, "num of nonces", len(tokens[token]), "num of intervals", len(intervals))
	}

	return ret
}

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
		//log.Info("sending", "tx data", string(tx.Data))
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

func createTxsData(tokens map[string][]*interval, intervalBulkSize int) ([][]byte, error) {
	txsData := make([][]byte, 0)
	for token, intervals := range tokens {
		dataIntervals, err := splitDataIntervals(token, intervals, intervalBulkSize)
		if err != nil {
			return nil, err
		}

		txsData = append(txsData, dataIntervals...)
	}

	return txsData, nil
}

func splitDataIntervals(token string, intervals []*interval, intervalBulkSize int) ([][]byte, error) {
	bulks := len(intervals) / intervalBulkSize
	allData := make([][]byte, 0)

	start := 0
	end := intervalBulkSize
	for i := 0; i < bulks; i++ {
		d, err := tokensIntervalsAsOnData(token, intervals[start:end])
		if err != nil {
			return nil, err
		}

		start += intervalBulkSize
		end += intervalBulkSize
		allData = append(allData, d)
	}

	remaining := len(intervals) % intervalBulkSize
	if remaining == 0 {
		return allData, nil
	}

	d, err := tokensIntervalsAsOnData(token, intervals[start:start+remaining])
	if err != nil {
		return nil, err
	}
	allData = append(allData, d)

	return allData, nil
}

func tokensIntervalsAsOnData(token string, intervals []*interval) ([]byte, error) {
	builder := builders.NewTxDataBuilder().
		Function("ESDTDeleteMetadata").
		ArgHexString(hex.EncodeToString([]byte(token))).
		ArgInt64(int64(len(intervals)))

	for _, interval := range intervals {
		builder.
			ArgInt64(int64(interval.start)).
			ArgInt64(int64(interval.end))
	}

	return builder.ToDataBytes()
}
