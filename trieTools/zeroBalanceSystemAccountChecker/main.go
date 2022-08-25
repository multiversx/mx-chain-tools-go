package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ElrondNetwork/elrond-go-core/core/pubkeyConverter"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-tools-go/elasticreindexer/config"
	"github.com/ElrondNetwork/elrond-tools-go/elasticreindexer/elastic"
	sysAccConfig "github.com/ElrondNetwork/elrond-tools-go/trieTools/zeroBalanceSystemAccountChecker/config"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/pelletier/go-toml"
	"github.com/urfave/cli"
	"strconv"

	"github.com/ElrondNetwork/elrond-tools-go/trieTools/trieToolsCommon"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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

	log.Info("sanity checks...")

	err := logger.SetLogLevel(flagsConfig.LogLevel)
	if err != nil {
		return err
	}

	log.Info("starting processing trie", "pid", os.Getpid())

	globalAddressTokensMap, shardAddressTokensMap, err := readInputs(flagsConfig.TokensDirectory)
	if err != nil {
		return err
	}

	globalExtraTokens, extraTokensPerShard, err := exportSystemAccZeroTokensBalances(globalAddressTokensMap, shardAddressTokensMap)
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

func readInputs(tokensDir string) (map[string]map[string]struct{}, map[uint32]map[string]map[string]struct{}, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return nil, nil, err
	}

	fullPath := filepath.Join(workingDir, tokensDir)
	contents, err := ioutil.ReadDir(fullPath)
	if err != nil {
		return nil, nil, err
	}

	globalAddressTokensMap := make(map[string]map[string]struct{})
	shardAddressTokensMap := make(map[uint32]map[string]map[string]struct{})
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

		shardAddressTokensMap[shardID] = copyMap(addressTokensMapInCurrFile)
		merge(globalAddressTokensMap, addressTokensMapInCurrFile)

		log.Info("read data from",
			"file", file.Name(),
			"shard", shardID,
			"num tokens in shard", getNumTokens(shardAddressTokensMap[shardID]),
			"num addresses in current file", len(shardAddressTokensMap[shardID]),
			"total num addresses in all shards", len(globalAddressTokensMap))
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

func getFileContent(file string) (map[string]map[string]struct{}, error) {
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

	ret := make(map[string]map[string]struct{})
	for address, tokens := range addressTokensMapInCurrFile {
		ret[address] = make(map[string]struct{})
		for token := range tokens {
			addTokenInMapIfHasNonce(token, ret[address])
		}
	}

	return ret, nil
}

func addTokenInMapIfHasNonce(token string, tokens map[string]struct{}) {
	if hasNonce(token) {
		tokens[token] = struct{}{}
	}
}

func hasNonce(token string) bool {
	return strings.Count(token, "-") == 2
}

func copyMap(addressTokensMap map[string]map[string]struct{}) map[string]map[string]struct{} {
	addressTokensMapCopy := make(map[string]map[string]struct{})

	for address, tokens := range addressTokensMap {
		addressTokensMapCopy[address] = make(map[string]struct{})
		for token := range tokens {
			addressTokensMapCopy[address][token] = struct{}{}
		}
	}

	return addressTokensMapCopy
}

func getNumTokens(addressTokensMap map[string]map[string]struct{}) int {
	numTokensInShard := 0
	for _, tokens := range addressTokensMap {
		for range tokens {
			numTokensInShard++
		}
	}

	return numTokensInShard
}

func merge(dest, src map[string]map[string]struct{}) {
	for addressSrc, tokensSrc := range src {
		_, existsInDest := dest[addressSrc]
		if !existsInDest {
			dest[addressSrc] = tokensSrc
		} else {
			log.Debug("same address found in multiple files", "address", addressSrc)
			addTokensInDestAddress(tokensSrc, dest, addressSrc)
		}
	}
}

func addTokensInDestAddress(tokens map[string]struct{}, dest map[string]map[string]struct{}, address string) {
	for token := range tokens {
		dest[address][token] = struct{}{}
	}
}

func exportSystemAccZeroTokensBalances(
	globalAddressTokensMap map[string]map[string]struct{},
	shardAddressTokenMap map[uint32]map[string]map[string]struct{},
) (map[string]struct{}, map[uint32]map[string]struct{}, error) {
	addressConverter, err := pubkeyConverter.NewBech32PubkeyConverter(addressLength, log)
	if err != nil {
		return nil, nil, err
	}

	systemSCAddress := addressConverter.Encode(vmcommon.SystemAccountAddress)
	globalExtraTokens, err := getGlobalExtraTokens(globalAddressTokensMap, systemSCAddress)
	if err != nil {
		return nil, nil, err
	}
	log.Info("found", "global num of extra tokens", len(globalExtraTokens))

	shardExtraTokens := make(map[uint32]map[string]struct{})
	for shardID, addressTokensMap := range shardAddressTokenMap {
		shardTokensInSystemAccAddress, found := addressTokensMap[systemSCAddress]
		if !found {
			return nil, nil, fmt.Errorf("no system account address(%s) found in shard = %v", systemSCAddress, shardID)
		}

		extraTokensInShard := intersection(globalExtraTokens, shardTokensInSystemAccAddress)
		log.Info("found", "shard", shardID, "num tokens in system account", len(shardTokensInSystemAccAddress), "num extra tokens", len(extraTokensInShard))
		shardExtraTokens[shardID] = extraTokensInShard
	}
	if !sanityCheckExtraTokens(shardExtraTokens, globalExtraTokens) {
		return nil, nil, errors.New("sanity check for exported tokens failed")
	}

	return globalExtraTokens, shardExtraTokens, nil
}

func getGlobalExtraTokens(allAddressesTokensMap map[string]map[string]struct{}, systemSCAddress string) (map[string]struct{}, error) {
	allTokensInSystemSCAddress, foundSystemSCAddress := allAddressesTokensMap[systemSCAddress]
	if !foundSystemSCAddress {
		return nil, fmt.Errorf("no system account address(%s) found", systemSCAddress)
	}

	allTokens := getAllTokensWithoutSystemAccount(allAddressesTokensMap, systemSCAddress)
	log.Info("found",
		"global num of tokens in all addresses without system account", len(allTokens),
		"global num of tokens in system sc address", len(allTokensInSystemSCAddress))

	return getExtraTokens(allTokens, allTokensInSystemSCAddress), nil
}

func getAllTokensWithoutSystemAccount(allAddressesTokensMap map[string]map[string]struct{}, systemSCAddress string) map[string]struct{} {
	allAddressTokensMapCopy := copyMap(allAddressesTokensMap)
	delete(allAddressTokensMapCopy, systemSCAddress)

	allTokens := make(map[string]struct{})
	for _, tokens := range allAddressTokensMapCopy {
		for token := range tokens {
			allTokens[token] = struct{}{}
		}
	}

	return allTokens
}

func getExtraTokens(allTokens, allTokensInSystemSCAddress map[string]struct{}) map[string]struct{} {
	extraTokens := make(map[string]struct{})
	for tokenInSystemSC := range allTokensInSystemSCAddress {
		_, exists := allTokens[tokenInSystemSC]
		if !exists {
			extraTokens[tokenInSystemSC] = struct{}{}
		}
	}

	log.Info("found", "num of sfts/nfts/metaesdts metadata only found in system sc address", len(extraTokens))
	return extraTokens
}

func intersection(globalTokens, shardTokens map[string]struct{}) map[string]struct{} {
	ret := make(map[string]struct{})
	for token := range shardTokens {
		_, found := globalTokens[token]
		if found {
			ret[token] = struct{}{}
		}
	}

	return ret
}

func sanityCheckExtraTokens(shardExtraTokensMap map[uint32]map[string]struct{}, globalExtraTokens map[string]struct{}) bool {
	allMergedExtraTokens := make(map[string]struct{})
	for _, extraTokensInShard := range shardExtraTokensMap {
		for extraToken := range extraTokensInShard {
			allMergedExtraTokens[extraToken] = struct{}{}
		}
	}

	return checkSameMap(allMergedExtraTokens, globalExtraTokens)
}

func checkSameMap(map1, map2 map[string]struct{}) bool {
	if len(map1) != len(map2) {
		return false
	}

	for elemInMap1 := range map1 {
		_, foundInMap2 := map2[elemInMap1]
		if !foundInMap2 {
			return false
		}
	}

	return true
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

	nftGetter := newNFTBalanceGetter(cfg.Config.Gateway.URL)
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
