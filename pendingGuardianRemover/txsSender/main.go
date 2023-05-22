package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-logger-go/file"
	"github.com/multiversx/mx-chain-tools-go/pendingGuardianRemover/httpClientWrapper"
	"github.com/multiversx/mx-chain-tools-go/pendingGuardianRemover/txsFileHandler"
	"github.com/multiversx/mx-chain-tools-go/pendingGuardianRemover/txsSender/config"
	"github.com/multiversx/mx-chain-tools-go/pendingGuardianRemover/txsSender/txSender"
	"github.com/multiversx/mx-sdk-go/core/http"
	"github.com/urfave/cli"
)

const (
	defaultLogsPath  = "logs"
	logFilePrefix    = "multi-factor-auth-go-service"
	logMaxSizeInMB   = 1024
	logLifeSpanInSec = 86400
)

var log = logger.GetOrCreate("main")

var (
	logLevel = cli.StringFlag{
		Name: "log-level",
		Usage: "This flag specifies the logger `level(s)`. It can contain multiple comma-separated value. For example" +
			", if set to *:INFO the logs for all packages will have the INFO level. However, if set to *:INFO,api:DEBUG" +
			" the logs for all packages will have the INFO level, excepting the api package which will receive a DEBUG" +
			" log level.",
		Value: "*:" + logger.LogDebug.String(),
	}
	// configurationFile defines a flag for the path to the main toml configuration file
	configurationFile = cli.StringFlag{
		Name: "config",
		Usage: "The path for the main configuration file. This TOML file contain the main " +
			"configurations such as storage setups, epoch duration and so on.",
		Value: "config.toml",
	}
	// workingDirectory defines a flag for the path for the working directory.
	workingDirectory = cli.StringFlag{
		Name:  "working-directory",
		Usage: "This flag specifies the `directory` where the node will store databases and logs.",
		Value: "",
	}
	// logSaveFile is used when the log output needs to be logged in a file
	logSaveFile = cli.BoolFlag{
		Name:  "log-save",
		Usage: "Boolean option for enabling log saving. If set, it will automatically save all the logs into a file.",
	}
)

func main() {
	app := cli.NewApp()
	app.Name = "SetGuardian txs sender CLI app"
	app.Usage = "This is the entry point for the SetGuardian transactions sender"
	app.Flags = getFlags()
	app.Authors = []cli.Author{
		{
			Name:  "The MultiversX Team",
			Email: "contact@multiversx.com",
		},
	}

	app.Action = func(c *cli.Context) error {
		return startProcess(c)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func startProcess(ctx *cli.Context) error {
	flags := getFlagsConfig(ctx)

	err := attachLoggers(flags)
	if err != nil {
		return err
	}

	cfg, err := loadConfig(flags.ConfigurationFile)
	if err != nil {
		return err
	}

	fileHandler, err := txsFileHandler.NewFileHandler(cfg.TxSender.TxsFile)
	if err != nil {
		return err
	}

	httpClient := http.NewHttpClientWrapper(nil, cfg.API.NetworkAddress)
	wrapper, err := httpClientWrapper.NewHttpClientWrapper(httpClient)
	if err != nil {
		return err
	}

	txsMap, err := fileHandler.Load()
	if err != nil {
		return err
	}

	sender, err := txSender.NewTxSender(wrapper, time.Duration(cfg.TxSender.DelayBetweenSendsInSec)*time.Second, txsMap)
	if err != nil {
		return err
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs

	log.Info("application closing, calling Close on sender...")
	return sender.Close()
}

func attachLoggers(flags config.ContextFlagsConfig) error {
	logLevelFlagValue := flags.LogLevel
	err := logger.SetLogLevel(logLevelFlagValue)
	if err != nil {
		return err
	}

	if flags.SaveLogFile {
		args := file.ArgsFileLogging{
			WorkingDir:      flags.WorkingDir,
			DefaultLogsPath: defaultLogsPath,
			LogFilePrefix:   logFilePrefix,
		}
		fileLogging, err := file.NewFileLogging(args)
		if err != nil {
			return fmt.Errorf("%w creating a log file", err)
		}

		err = fileLogging.ChangeFileLifeSpan(time.Second*time.Duration(logLifeSpanInSec), logMaxSizeInMB)
		if err != nil {
			return err
		}
	}

	return nil
}

func loadConfig(filepath string) (config.Configs, error) {
	cfg := config.Configs{}
	err := core.LoadTomlFile(&cfg, filepath)
	if err != nil {
		return config.Configs{}, err
	}

	return cfg, nil
}

func getFlags() []cli.Flag {
	return []cli.Flag{
		logLevel,
		configurationFile,
		workingDirectory,
		logSaveFile,
	}
}

func getFlagsConfig(ctx *cli.Context) config.ContextFlagsConfig {
	flagsConfig := config.ContextFlagsConfig{}

	flagsConfig.LogLevel = ctx.GlobalString(logLevel.Name)
	flagsConfig.ConfigurationFile = ctx.GlobalString(configurationFile.Name)
	flagsConfig.WorkingDir = ctx.GlobalString(workingDirectory.Name)
	flagsConfig.SaveLogFile = ctx.GlobalBool(logSaveFile.Name)

	return flagsConfig
}
