package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ElrondNetwork/elrond-go-core/core/pubkeyConverter"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-tools-go/elasticreindexer/config"
	"github.com/ElrondNetwork/elrond-tools-go/elasticreindexer/elastic"
	"github.com/ElrondNetwork/elrond-tools-go/trieTools/trieToolsCommon"
	sysAccConfig "github.com/ElrondNetwork/elrond-tools-go/trieTools/zeroBalanceSystemAccountChecker/config"
	"github.com/pelletier/go-toml"
	"github.com/urfave/cli"
)

const (
	logFilePrefix   = "system-account-zero-tokens-balance-checker"
	addressLength   = 32
	outputFilePerms = 0644
	tomlFile        = "./config.toml"
)

type crossTokenChecker interface {
	crossCheckExtraTokens(tokens map[string]struct{}) ([]string, error)
}

func main() {
	app := cli.NewApp()
	app.Name = "Tokens exporter CLI app"
	app.Usage = "This is the entry point for the tool that checks which tokens are not used anymore(only stored in system account)"
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

	err := logger.SetLogLevel(flagsConfig.LogLevel)
	if err != nil {
		return err
	}

	globalAddressTokensMap, shardAddressTokensMap, err := readInputs(flagsConfig.TokensDirectory)
	if err != nil {
		return err
	}

	addressConverter, err := pubkeyConverter.NewBech32PubkeyConverter(addressLength, log)
	if err != nil {
		return err
	}
	exporter, err := newZeroTokensBalancesExporter(addressConverter)
	if err != nil {
		return err
	}

	globalExtraTokens, extraTokensPerShard, err := exporter.getExtraTokens(globalAddressTokensMap, shardAddressTokensMap)
	if err != nil {
		return err
	}

	if flagsConfig.CrossCheck {
		err = crossCheckExtraTokens(globalExtraTokens, extraTokensPerShard)
		if err != nil {
			return err
		}
	}

	err = saveResult(extraTokensPerShard, flagsConfig.Outfile)
	if err != nil {
		return err
	}

	return nil
}

func readInputs(tokensDir string) (trieToolsCommon.AddressTokensMap, map[uint32]trieToolsCommon.AddressTokensMap, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return nil, nil, err
	}

	fullPath := filepath.Join(workingDir, tokensDir)
	contents, err := ioutil.ReadDir(fullPath)
	if err != nil {
		return nil, nil, err
	}

	globalAddressTokensMap := trieToolsCommon.NewAddressTokensMap()
	shardAddressTokensMap := make(map[uint32]trieToolsCommon.AddressTokensMap)
	for _, file := range contents {
		if file.IsDir() {
			continue
		}

		shardID, err := getShardID(file.Name())
		if err != nil {
			return nil, nil, err
		}

		addressTokensMapInCurrFile, err := getFileContent(filepath.Join(fullPath, file.Name()))
		if err != nil {
			return nil, nil, err
		}

		shardAddressTokensMap[shardID] = addressTokensMapInCurrFile.ShallowClone()
		merge(globalAddressTokensMap, addressTokensMapInCurrFile)

		log.Info("read data from",
			"file", file.Name(),
			"shard", shardID,
			"num tokens in shard", shardAddressTokensMap[shardID].NumTokens(),
			"num addresses in shard", shardAddressTokensMap[shardID].NumAddresses(),
			"total num addresses in all shards", globalAddressTokensMap.NumAddresses())
	}

	return globalAddressTokensMap, shardAddressTokensMap, nil
}

func getShardID(file string) (uint32, error) {
	shardIDStr := strings.TrimPrefix(file, "shard")
	shardIDStr = strings.TrimSuffix(shardIDStr, ".json")
	shardID, err := strconv.Atoi(shardIDStr)
	if err != nil {
		return 0, fmt.Errorf("invalid file input name; expected tokens shard file name to be <shardX.json>, where X = number(e.g. shard0.json)")
	}

	return uint32(shardID), nil
}

func getFileContent(file string) (trieToolsCommon.AddressTokensMap, error) {
	jsonFile, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	bytesFromJson, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	addressTokensMapInCurrFile := make(map[string]map[string]struct{})
	err = json.Unmarshal(bytesFromJson, &addressTokensMapInCurrFile)
	if err != nil {
		return nil, err
	}

	ret := trieToolsCommon.NewAddressTokensMap()
	for address, tokens := range addressTokensMapInCurrFile {
		tokensWithNonce := getTokensWithNonce(tokens)
		ret.Add(address, tokensWithNonce)
	}

	return ret, nil
}

func getTokensWithNonce(tokens map[string]struct{}) map[string]struct{} {
	ret := make(map[string]struct{})

	for token := range tokens {
		addTokenInMapIfHasNonce(token, ret)
	}

	return ret
}

func addTokenInMapIfHasNonce(token string, tokens map[string]struct{}) {
	if hasNonce(token) {
		tokens[token] = struct{}{}
	}
}

func hasNonce(token string) bool {
	return strings.Count(token, "-") == 2
}

func merge(dest, src trieToolsCommon.AddressTokensMap) {
	for addressSrc, tokensSrc := range src.GetMapCopy() {
		if dest.HasAddress(addressSrc) {
			log.Debug("same address found in multiple files", "address", addressSrc)
		}

		dest.Add(addressSrc, tokensSrc)
	}
}

func saveResult(tokens map[uint32]map[string]struct{}, outfile string) error {
	jsonBytes, err := json.MarshalIndent(tokens, "", " ")
	if err != nil {
		return err
	}

	log.Info("writing result in", "file", outfile)
	err = ioutil.WriteFile(outfile, jsonBytes, fs.FileMode(outputFilePerms))
	if err != nil {
		return err
	}

	log.Info("finished exporting zero balance tokens map")
	return nil
}

func crossCheckExtraTokens(globalExtraTokens map[string]struct{}, extraTokensPerShard map[uint32]map[string]struct{}) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	nftGetter := newTokenBalanceGetter(cfg.Config.Gateway.URL)
	elasticClient, err := elastic.NewElasticClient(config.ElasticInstanceConfig{
		URL:      cfg.Config.ElasticIndexerConfig.URL,
		Username: cfg.Config.ElasticIndexerConfig.Username,
		Password: cfg.Config.ElasticIndexerConfig.Password,
	})
	if err != nil {
		return err
	}

	tokensChecker, err := newExtraTokensCrossChecker(elasticClient, nftGetter)
	if err != nil {
		return err
	}

	tokensThatStillExist, err := tokensChecker.crossCheckExtraTokens(globalExtraTokens)
	if err != nil {
		return err
	}

	if len(tokensThatStillExist) == 0 {
		log.Info("all cross-checks were successful; exported tokens are only stored in system account")
		return nil
	}

	log.Error("found tokens with balances that still exist in other accounts; probably found in pending mbs during snapshot; will remove them from exported tokens",
		"tokens", tokensThatStillExist)
	for _, extraTokensInShard := range extraTokensPerShard {
		removeTokensThatStillExist(tokensThatStillExist, extraTokensInShard)
	}

	return nil
}

func loadConfig() (*sysAccConfig.GeneralConfig, error) {
	tomlBytes, err := ioutil.ReadFile(tomlFile)
	if err != nil {
		return nil, err
	}

	var tc sysAccConfig.GeneralConfig
	err = toml.Unmarshal(tomlBytes, &tc)
	if err != nil {
		return nil, err
	}

	return &tc, nil
}

func removeTokensThatStillExist(tokensThatStillExist []string, tokens map[string]struct{}) {
	for _, token := range tokensThatStillExist {
		delete(tokens, token)
	}
}
