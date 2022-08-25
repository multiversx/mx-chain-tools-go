package main

import (
	"encoding/json"
	"fmt"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/blockchain"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/builders"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/interactors"
	"github.com/ElrondNetwork/elrond-tools-go/miscellaneous/metaDataRemover/config"
	"github.com/ElrondNetwork/elrond-tools-go/trieTools/trieToolsCommon"
	"github.com/pelletier/go-toml"
	"github.com/urfave/cli"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	ESDTDeleteMetadataPrefix = "ESDTDeleteMetadata"
	logFilePrefix            = "meta-data-remover"
	tomlFile                 = "./config.toml"
	outputFilePerms          = 0644
	txsBulkSize              = 100
)

type interval struct {
	start uint64
	end   uint64
}

type tokenData struct {
	tokenID   string
	intervals []*interval
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

	shardTokensMap, err := readInput(flagsConfig.Tokens)
	if err != nil {
		return err
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	shardTxsDataMap := make(map[uint32][][]byte)
	for shardID, tokens := range shardTokensMap {
		log.Info("creating txs data", "shardID", shardID, "num tokens", len(tokens))
		tokensSorted, err := sortTokensIDByNonce(tokens)
		if err != nil {
			return err
		}

		tokensIntervals := groupTokensByIntervals(tokensSorted)
		//_ = sortTokensByMaxConsecutiveNonces(tokensIntervals)

		txsData, err := createTxsData(tokensIntervals, cfg.TokensToDeletePerTransaction)
		if err != nil {
			return err
		}

		log.Info("created", "num of txs", len(txsData), "shardID", shardID, "num of nonces per tx", cfg.TokensToDeletePerTransaction)
		shardTxsDataMap[shardID] = txsData
	}

	args := blockchain.ArgsElrondProxy{
		ProxyURL:            cfg.ProxyUrl,
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

	shardPemsDataMap, err := readPemsData(flagsConfig.Pems)
	if err != nil {
		return err
	}

	if len(shardPemsDataMap) != len(shardTxsDataMap) {
		return fmt.Errorf("provided invalid input; expected number of pem files = number of shards in tokens input; got len(<shard,tokens>) = %d, len(<shard, pem>) = %d",
			len(shardPemsDataMap), len(shardTxsDataMap))
	}

	shardTxsMap := make(map[uint32][]*data.Transaction)
	for shardID, txsData := range shardTxsDataMap {
		pemData, found := shardPemsDataMap[shardID]
		if !found {
			return fmt.Errorf("no pem data provided for shard = %d", shardID)
		}

		log.Info("starting to create txs", "shardID", shardID, "num of txs", len(txsData))
		txsInShard, err := createTxs(pemData, proxy, ti, txsData, cfg.GasLimit)
		if err != nil {
			return err
		}

		shardTxsMap[shardID] = txsInShard

		file := "txsShard" + strconv.Itoa(int(shardID)) + ".json"
		log.Info("saving txs", "shardID", shardID, "file", file)
		err = saveResult(txsInShard, file)
	}

	//log.Info("starting to create txs", "num of txs", len(txsData))
	//err = sendTxs(flagsConfig.Pem, proxy, ti, txsData, txsBulkSize)
	//if err != nil {
	//	return err
	//}
	_ = ti
	return nil
}

func saveResult(txs []*data.Transaction, outfile string) error {
	jsonBytes, err := json.MarshalIndent(txs, "", " ")
	if err != nil {
		return err
	}

	log.Info("writing result in", "file", outfile)
	err = ioutil.WriteFile(outfile, jsonBytes, fs.FileMode(outputFilePerms))
	if err != nil {
		return err
	}
	return nil
}

func readInput(tokensFile string) (map[uint32]map[string]struct{}, error) {
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

	shardTokensMap := make(map[uint32]map[string]struct{})
	err = json.Unmarshal(bytesFromJson, &shardTokensMap)
	if err != nil {
		return nil, err
	}

	log.Info("read from input", "file", tokensFile, "num of shards", len(shardTokensMap), getNumTokens(shardTokensMap))
	return shardTokensMap, nil
}

func readPemsData(pemsFile string) (map[uint32]*pkAddress, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	fullPath := filepath.Join(workingDir, pemsFile)
	contents, err := ioutil.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	shardPemDataMap := make(map[uint32]*pkAddress)
	for _, file := range contents {
		if file.IsDir() {
			continue
		}

		shardID, err := getShardID(file.Name())
		if err != nil {
			return nil, err
		}

		pemData, err := getPrivateKeyAndAddress(filepath.Join(fullPath, file.Name()))
		if err != nil {
			return nil, err
		}

		shardPemDataMap[shardID] = pemData
	}

	return shardPemDataMap, nil
}

func getShardID(file string) (uint32, error) {
	shardIDStr := strings.TrimPrefix(file, "shard")
	shardIDStr = strings.TrimSuffix(shardIDStr, ".pem")
	shardID, err := strconv.Atoi(shardIDStr)
	if err != nil {
		return 0, fmt.Errorf("invalid file input name = %s; expected pem file name to be <shardX.json>, where X = number(e.g. shard0.json)", file)
	}

	return uint32(shardID), nil
}

func getNumTokens(shardTokensMap map[uint32]map[string]struct{}) int {
	numTokensInShard := 0
	for _, tokens := range shardTokensMap {
		for range tokens {
			numTokensInShard++
		}
	}

	return numTokensInShard
}

func loadConfig() (*config.Config, error) {
	tomlBytes, err := ioutil.ReadFile(tomlFile)
	if err != nil {
		return nil, err
	}

	var cfg config.Config
	err = toml.Unmarshal(tomlBytes, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
