package main

import (
	"encoding/json"
	"fmt"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	elrondFactory "github.com/ElrondNetwork/elrond-go/cmd/node/factory"
	"github.com/ElrondNetwork/elrond-go/common/logging"
	"github.com/ElrondNetwork/elrond-tools-go/trieTools/trieToolsCommon"
	"github.com/urfave/cli"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	defaultLogsPath = "logs"
	logFilePrefix   = "accounts-tokens-exporter"
	outputFileName  = "output.json"
	outputFilePerms = 0644
)

func main() {
	app := cli.NewApp()
	app.Name = "Tokens exporter CLI app"
	app.Usage = "This is the entry point for the tool that exports all tokens for a given root hash"
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

	log.Info("finished exporting address-tokens map")
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

	mydir, err := os.Getwd()
	if err != nil {
		return err
	}
	_, err = readInputs(filepath.Join(mydir, flagsConfig.tokensDirectory))
	if err != nil {
		return err
	}

	return exportZeroTokensBalances()
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

func readInputs(parentDir string) (map[string]map[string]struct{}, error) {
	fmt.Println("HEREEEEE" + parentDir)
	contents, err := ioutil.ReadDir(parentDir)
	if err != nil {
		return nil, err
	}

	allAddressesTokensMap := make(map[string]map[string]struct{})
	for _, c := range contents {
		if c.IsDir() {
			continue
		}

		jsonFile, err := os.Open(filepath.Join(parentDir, c.Name()))
		if err != nil {
			return nil, err
		}

		bytesFromJson, _ := ioutil.ReadAll(jsonFile)
		addressTokensMapInCurrFile := make(map[string]map[string]struct{})
		err = json.Unmarshal(bytesFromJson, &addressTokensMapInCurrFile)
		if err != nil {
			return nil, err
		}

		merge(allAddressesTokensMap, addressTokensMapInCurrFile)
		log.Info("read data from",
			"file", c.Name(),
			"num addresses in current file", len(addressTokensMapInCurrFile),
			"num addresses in total", len(allAddressesTokensMap))
	}

	//for address, tokens := range allAddressesTokensMap {
	//	for token := range tokens {
	//		log.Info("", "address", address, "token", token)
	//	}
	//}

	return allAddressesTokensMap, nil
}

func merge(dest, src map[string]map[string]struct{}) {
	for addressSrc, tokensSrc := range src {
		_, existsInDest := dest[addressSrc]
		if !existsInDest {
			dest[addressSrc] = tokensSrc
		} else {
			for tokenInSrc := range tokensSrc {
				dest[addressSrc][tokenInSrc] = struct{}{}
			}
		}
	}
}

func exportZeroTokensBalances() error {
	return nil
}
