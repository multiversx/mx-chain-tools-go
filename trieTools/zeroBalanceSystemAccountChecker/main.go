package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/pubkeyConverter"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	elrondFactory "github.com/ElrondNetwork/elrond-go/cmd/node/factory"
	"github.com/ElrondNetwork/elrond-go/common/logging"
	"github.com/ElrondNetwork/elrond-tools-go/elasticreindexer/config"
	"github.com/ElrondNetwork/elrond-tools-go/elasticreindexer/elastic"
	"github.com/tidwall/gjson"
	"github.com/urfave/cli"

	"github.com/ElrondNetwork/elrond-tools-go/trieTools/trieToolsCommon"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultLogsPath    = "logs"
	logFilePrefix      = "accounts-tokens-exporter"
	addressLength      = 32
	outputFilePerms    = 0644
	maxRequestsRetrial = 10
	multipleSearchBulk = 10000
)

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

	_, errLogger := attachFileLogger(log, flagsConfig.ContextFlagsConfig)
	if errLogger != nil {
		return errLogger
	}

	log.Info("sanity checks...")

	err := logger.SetLogLevel(flagsConfig.LogLevel)
	if err != nil {
		return err
	}

	log.Info("starting processing trie", "pid", os.Getpid())

	addressTokensMap, err := readInputs(flagsConfig.TokensDirectory)
	if err != nil {
		return err
	}

	extraTokens, err := exportSystemAccZeroTokensBalances(addressTokensMap)
	if err != nil {
		return err
	}

	err = saveResult(extraTokens, flagsConfig.Outfile)
	if err != nil {
		return err
	}

	if flagsConfig.CrossCheck {
		return crossCheckExtraTokens(extraTokens)
	}

	return nil
}

func attachFileLogger(log logger.Logger, flagsConfig trieToolsCommon.ContextFlagsConfig) (elrondFactory.FileLoggingHandler, error) {
	var fileLogging elrondFactory.FileLoggingHandler
	var err error
	if flagsConfig.SaveLogFile {
		fileLogging, err = logging.NewFileLogging(logging.ArgsFileLogging{
			WorkingDir:      flagsConfig.WorkingDir,
			DefaultLogsPath: defaultLogsPath,
			LogFilePrefix:   logFilePrefix,
		})
		if err != nil {
			return nil, fmt.Errorf("%w creating a log file", err)
		}
	}

	err = logger.SetDisplayByteSlice(logger.ToHex)
	log.LogIfError(err)
	logger.ToggleLoggerName(flagsConfig.EnableLogName)
	logLevelFlagValue := flagsConfig.LogLevel
	err = logger.SetLogLevel(logLevelFlagValue)
	if err != nil {
		return nil, err
	}

	if flagsConfig.DisableAnsiColor {
		err = logger.RemoveLogObserver(os.Stdout)
		if err != nil {
			return nil, err
		}

		err = logger.AddLogObserver(os.Stdout, &logger.PlainFormatter{})
		if err != nil {
			return nil, err
		}
	}
	log.Trace("logger updated", "level", logLevelFlagValue, "disable ANSI color", flagsConfig.DisableAnsiColor)

	return fileLogging, nil
}

func readInputs(tokensDir string) (map[string]map[string]struct{}, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	fullPath := filepath.Join(workingDir, tokensDir)
	contents, err := ioutil.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	allAddressesTokensMap := make(map[string]map[string]struct{})
	for _, file := range contents {
		if file.IsDir() {
			continue
		}

		addressTokensMapInCurrFile, err := getFileContent(filepath.Join(fullPath, file.Name()))
		if err != nil {
			return nil, err
		}

		merge(allAddressesTokensMap, addressTokensMapInCurrFile)
		log.Info("read data from",
			"file", file.Name(),
			"num addresses in current file", len(addressTokensMapInCurrFile),
			"num addresses in total, after merge", len(allAddressesTokensMap))
	}

	return allAddressesTokensMap, nil
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

	return addressTokensMapInCurrFile, nil
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

func exportSystemAccZeroTokensBalances(allAddressesTokensMap map[string]map[string]struct{}) (map[string]struct{}, error) {
	addressConverter, err := pubkeyConverter.NewBech32PubkeyConverter(addressLength, log)
	if err != nil {
		return nil, err
	}

	systemSCAddress := addressConverter.Encode(vmcommon.SystemAccountAddress)
	allTokensInSystemSCAddress, foundSystemSCAddress := allAddressesTokensMap[systemSCAddress]
	if !foundSystemSCAddress {
		return nil, fmt.Errorf("no system account address(%s) found", systemSCAddress)
	}

	allTokens := getAllTokensWithoutSystemAccount(allAddressesTokensMap, systemSCAddress)
	log.Info("found",
		"total num of tokens in all addresses", len(allTokens),
		"total num of tokens in system sc address", len(allTokensInSystemSCAddress))

	return getExtraTokens(allTokens, allTokensInSystemSCAddress), nil
}

func getAllTokensWithoutSystemAccount(allAddressesTokensMap map[string]map[string]struct{}, systemSCAddress string) map[string]struct{} {
	delete(allAddressesTokensMap, systemSCAddress)

	allTokens := make(map[string]struct{})
	for _, tokens := range allAddressesTokensMap {
		for token := range tokens {
			allTokens[token] = struct{}{}
		}
	}

	return allTokens
}

func getExtraTokens(allTokens, allTokensInSystemSCAddress map[string]struct{}) map[string]struct{} {
	ctTokensOnlyInSystemAcc := 0
	extraTokens := make(map[string]struct{})
	for tokenInSystemSC := range allTokensInSystemSCAddress {
		_, exists := allTokens[tokenInSystemSC]
		if !exists {
			ctTokensOnlyInSystemAcc++
			addTokenInMapIfHasNonce(tokenInSystemSC, extraTokens)
		}
	}

	log.Info("found",
		"num tokens in system account, but not in any other address", ctTokensOnlyInSystemAcc,
		"num of sfts/nfts/metaesdts metadata only found in system sc address", len(extraTokens))

	return extraTokens
}

func addTokenInMapIfHasNonce(token string, tokens map[string]struct{}) {
	if hasNonce(token) {
		tokens[token] = struct{}{}
	}
}

func hasNonce(token string) bool {
	return strings.Count(token, "-") == 2
}

func saveResult(tokens map[string]struct{}, outfile string) error {
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

func crossCheckExtraTokens(tokens map[string]struct{}) error {
	numTokens := len(tokens)
	log.Info("starting to cross-check", "num of tokens", numTokens)

	elasticClient, err := elastic.NewElasticClient(config.ElasticInstanceConfig{
		URL:      "https://elastic-do-nvme-ams-k8s-gateway.elrond.ro/",
		Username: "read_user",
		Password: "RLnGWsyW8xAEuAWzvAwoCI",
	})
	if err != nil {
		return err
	}

	bulkSize := core.MinInt(multipleSearchBulk, numTokens)
	requests := make([]string, 0, bulkSize)
	currTokenIdx := 0
	ctRequests := 0
	nftGetter := newNFTBalanceGetter("https://gateway.elrond.com")
	for token := range tokens {
		currTokenIdx++
		requests = append(requests, createRequest(token))

		notEnoughRequests := len(requests) < bulkSize
		notLastBulk := currTokenIdx != numTokens
		if notEnoughRequests && notLastBulk {
			continue
		}

		respBytes, err := elasticClient.GetMultiple("accountsesdt", requests)
		if err != nil {
			log.Error("elasticClient.GetMultiple(accountsesdt, requests)",
				"error", err,
				"requests", requests)
			return err
		}

		responses := gjson.Get(string(respBytes), "responses").Array()
		err = checkIndexerResponse(requests, responses, nftGetter)
		if err != nil {
			return err
		}

		go printProgress(numTokens, currTokenIdx)

		ctRequests += len(requests)
		requests = make([]string, 0, bulkSize)
	}

	log.Info("finished cross-checking",
		"total num of tokens", numTokens,
		"total num of tokens cross-checked", currTokenIdx,
		"total num of tokens requests in indexer", ctRequests)

	if numTokens != currTokenIdx || numTokens != ctRequests {
		return errors.New("failed to cross check all tokens, check logs")
	}

	return nil
}

func checkIndexerResponse(requests []string, responses []gjson.Result, nftBalanceGetter *nftBalanceGetter) error {
	idxRequestedToken := 0
	for _, res := range responses {
		hits := res.Get("hits.hits").Array()
		if len(hits) != 0 {
			token := gjson.Get(requests[idxRequestedToken], "query.match.identifier.query").String()
			log.Debug("found token in indexer with hits/accounts",
				"token", token,
				"num hits/accounts", len(hits))

			checkFailed, err := crossCheckToken(hits, token, nftBalanceGetter)
			if err != nil {
				return err
			}

			if checkFailed {
				//TODO: HERREEEE
			}
		}
		idxRequestedToken++
	}

	return nil
}

func crossCheckToken(hits []gjson.Result, warnToken string, nftBalanceGetter *nftBalanceGetter) (bool, error) {
	checkFailed := false
	for _, hit := range hits {
		address := hit.Get("_source.address").String()
		balance, err := nftBalanceGetter.getBalance(address, warnToken)
		if err != nil {
			return false, err
		}

		log.Debug("checking gateway if token still exists in trie",
			"token", warnToken,
			"address", address)

		if balance != "0" {
			checkFailed = true
			log.Error("cross-check failed; found token which is still in other address",
				"token", warnToken,
				"balance", balance,
				"address", address)
			break
		} else {
			log.Warn("possible indexer problem",
				"token", warnToken,
				"hit in address", address,
				"found in trie", false)
		}
	}

	return checkFailed, nil
}

func createRequest(token string) string {
	return `{"query" : {"match" : { "identifier": {"query":"` + token + `","operator":"and"}}}}`
}

func printProgress(numTokens, numTokensCrossChecked int) {
	log.Info("status",
		"num cross checked tokens", numTokensCrossChecked,
		"remaining num of tokens to check", numTokens-numTokensCrossChecked,
		"progress(%)", (100*numTokensCrossChecked)/numTokens) // this should not panic with div by zero, since func is only called if numTokens > 0
}
