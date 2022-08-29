package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-tools-go/miscellaneous/metaDataRemover/config"
	"github.com/ElrondNetwork/elrond-tools-go/trieTools/trieToolsCommon"
	"github.com/pelletier/go-toml"
	"github.com/urfave/cli"
)

const (
	logFilePrefix   = "meta-data-remover"
	tomlFile        = "./config.toml"
	outputFilePerms = 0644
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

	shardTokensMap, err := readTokensInput(flagsConfig.Tokens)
	if err != nil {
		return err
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	shardTxsDataMap, err := createShardTxsDataMap(shardTokensMap, cfg.TokensToDeletePerTransaction)
	if err != nil {
		return err
	}

	shardPemsDataMap, err := readPemsData(flagsConfig.Pems)
	if err != nil {
		return err
	}

	return createShardTxs(flagsConfig.Outfile, cfg, shardPemsDataMap, shardTxsDataMap)
}

func readTokensInput(tokensFile string) (map[uint32]map[string]struct{}, error) {
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
		return 0, fmt.Errorf("invalid file input name = %s; expected pem file name to be <shardX.pem>, where X = number(e.g. shard0.pem)", file)
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
