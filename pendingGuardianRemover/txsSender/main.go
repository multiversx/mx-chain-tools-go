package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-tools-go/pendingGuardianRemover/httpClientWrapper"
	"github.com/multiversx/mx-chain-tools-go/pendingGuardianRemover/txsFileHandler"
	"github.com/multiversx/mx-chain-tools-go/pendingGuardianRemover/txsSender/config"
	"github.com/multiversx/mx-chain-tools-go/pendingGuardianRemover/txsSender/txSender"
	"github.com/multiversx/mx-sdk-go/core/http"
	"github.com/urfave/cli"
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
	}
}

func getFlagsConfig(ctx *cli.Context) config.ContextFlagsConfig {
	flagsConfig := config.ContextFlagsConfig{}

	flagsConfig.LogLevel = ctx.GlobalString(logLevel.Name)
	flagsConfig.ConfigurationFile = ctx.GlobalString(configurationFile.Name)

	return flagsConfig
}
