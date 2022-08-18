package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/blockchain"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/builders"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/examples"
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
	intervalsPerTx = 2
)

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
	log.Info("read from input", "file", flagsConfig.Tokens, "num of tokens", len(tokens))

	tokensSorted := sortTokensIDByNonce(tokens)

	tokensIntervals := groupTokensByIntervals(tokensSorted)

	txs, _ := createTxs(tokensIntervals)

	_ = txs

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

	return tokens, nil
}

func sortTokensIDByNonce(tokens map[string]struct{}) map[string][]uint64 {
	ret := make(map[string][]uint64)
	for token := range tokens {
		splits := strings.Split(token, "-")
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

	return ret
}

type interval struct {
	start uint64
	end   uint64
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

func createTxs(tokens map[string][]*interval) ([]transaction.Transaction, error) {
	args := blockchain.ArgsElrondProxy{
		ProxyURL:            "https://gateway.elrond.com",
		Client:              nil,
		SameScState:         false,
		ShouldBeSynced:      false,
		FinalityCheck:       false,
		CacheExpirationTime: time.Minute,
		EntityType:          core.Proxy,
	}

	proxy, _ := blockchain.NewElrondProxy(args)

	w := interactors.NewWallet()

	privateKey, err := w.LoadPrivateKeyFromPemData([]byte(examples.AlicePemContents))
	if err != nil {
		return nil, fmt.Errorf("unable to load alice.pem; error: %w", err)
	}
	// Generate address from private key
	address, err := w.GetAddressFromPrivateKey(privateKey)
	if err != nil {
		return nil, err

	}

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

	txBuilder, err := builders.NewTxBuilder(blockchain.NewTxSigner())
	if err != nil {
		return nil, err
	}

	ti, err := interactors.NewTransactionInteractor(proxy, txBuilder)
	if err != nil {
		return nil, err
	}

	txsData := createTxData(tokens)

	ctTxs := 0
	totalSentTxs := 0
	for _, txData := range txsData {
		transactionArguments.Data = []byte(txData)
		tx, err := ti.ApplySignatureAndGenerateTx(privateKey, transactionArguments)
		if err != nil {
			return nil, err
		}
		ti.AddTransaction(tx)

		transactionArguments.Nonce++
		ctTxs++
		if ctTxs < 100 {
			continue
		}

		hashes, err := ti.SendTransactionsAsBunch(context.Background(), 100)
		if err != nil {
			return nil, err
		}

		totalSentTxs += 100
		log.Info("transactions sent", "hashes", hashes)
	}

	if totalSentTxs < len(txsData) {
		hashes, err := ti.SendTransactionsAsBunch(context.Background(), len(txsData)-totalSentTxs)
		if err != nil {
			return nil, err
		}

		log.Info("transactions sent", "hashes", hashes)
	}

	return nil, nil
}

func createTxData(tokens map[string][]*interval) []string {
	txsData := make([]string, 0)
	for token, intervals := range tokens {
		if len(intervals) > intervalsPerTx {
			txsData = append(txsData, splitIntervals(token, intervals)...)
			continue
		}

		tokensOnData, _ := tokensIntervalsAsOnData(token, intervals)
		txsData = append(txsData, tokensOnData)
		continue

	}

	return txsData
}

func splitIntervals(token string, intervals []*interval) []string {
	bulks := len(intervals) / intervalsPerTx
	allData := make([]string, 0)

	start := 0
	end := intervalsPerTx
	for i := 0; i < bulks; i++ {
		d, _ := tokensIntervalsAsOnData(token, intervals[start:end])

		start += intervalsPerTx
		end += intervalsPerTx
		allData = append(allData, d)
	}

	remaining := len(intervals) % intervalsPerTx
	if remaining == 0 {
		return allData
	}

	d, _ := tokensIntervalsAsOnData(token, intervals[start:start+remaining])
	allData = append(allData, d)

	return allData
}

func tokensIntervalsAsOnData(token string, intervals []*interval) (string, error) {
	builder := builders.NewTxDataBuilder().
		Function("ESDTDeleteMetadata").
		ArgHexString(hex.EncodeToString([]byte(token))).
		ArgBigInt(big.NewInt(int64(len(intervals))))

	for _, interval := range intervals {
		builder.
			ArgBigInt(big.NewInt(int64(interval.start))).
			ArgBigInt(big.NewInt(int64(interval.end)))
	}

	return builder.ToDataString()
}
