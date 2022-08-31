package main

import (
	"encoding/json"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"

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

	fh := newOSFileHandler()
	inputReader := newAddressTokensMapFileReader(fh)

	globalAddressTokensMap, shardAddressTokensMap, err := inputReader.readInputs(flagsConfig.TokensDirectory)
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

	nftGetter := newTokenBalanceGetter(cfg.Config.Gateway.URL, http.Get)
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
