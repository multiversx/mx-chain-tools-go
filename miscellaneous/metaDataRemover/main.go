package main

import (
	"encoding/json"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/blockchain"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/builders"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/interactors"
	"github.com/ElrondNetwork/elrond-tools-go/miscellaneous/metaDataRemover/config"
	"github.com/ElrondNetwork/elrond-tools-go/trieTools/trieToolsCommon"
	"github.com/pelletier/go-toml"
	"github.com/urfave/cli"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

const (
	ESDTDeleteMetadataPrefix = "ESDTDeleteMetadata"
	logFilePrefix            = "meta-data-remover"
	tomlFile                 = "./config.toml"
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

	tokens, err := readInput(flagsConfig.Tokens)
	if err != nil {
		return err
	}

	tokensSorted, err := sortTokensIDByNonce(tokens)
	if err != nil {
		return err
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	tokensIntervals := groupTokensByIntervals(tokensSorted)
	txsData, err := createTxsData(tokensIntervals, cfg.TokensToDeletePerTransaction)
	if err != nil {
		return err
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
