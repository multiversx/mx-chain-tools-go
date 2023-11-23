package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/cmd/node/factory"
	"github.com/multiversx/mx-chain-go/facade"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-logger-go/file"
	"github.com/multiversx/mx-chain-tools-go/jsonToPrometheus/collector"
	"github.com/multiversx/mx-chain-tools-go/jsonToPrometheus/config"
	"github.com/multiversx/mx-chain-tools-go/jsonToPrometheus/httpClientWrapper"
	"github.com/multiversx/mx-sdk-go/core/http"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/urfave/cli"
)

var (
	helpTemplate = `NAME:
   {{.Name}} - {{.Usage}}
USAGE:
   {{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}
   {{if len .Authors}}
AUTHOR:
   {{range .Authors}}{{ . }}{{end}}
   {{end}}{{if .Commands}}
GLOBAL OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}
VERSION:
   {{.Version}}
   {{end}}
`
	// restApiInterfaceFlag defines a flag for the interface on which the rest API will try to bind with
	restApiInterfaceFlag = cli.StringFlag{
		Name: "rest-api-interface",
		Usage: "The interface `address and port` to which the REST API will attempt to bind. " +
			"To bind to all available interfaces, set this flag to :8080",
		Value: facade.DefaultRestInterface,
	}
	// proxyURL defines a flag for the proxy url that will be used to fetch validator statistics
	proxyURL = cli.StringFlag{
		Name:  "proxy-url",
		Usage: "The proxy url that will be used to fetch validator statistics",
		Value: "https://testnet-api.multiversx.com",
	}
	// chain defines a flag for the chain related to proxy url
	chain = cli.StringFlag{
		Name:  "chain",
		Usage: "The chain proxy runs on",
		Value: "public-testnet",
	}
	// configurationFile defines a flag for the path to the main toml configuration file
	configurationFile = cli.StringFlag{
		Name:  "config",
		Usage: "The path for the main configuration file",
		Value: "./config.toml",
	}
	// logLevel defines the logger level
	logLevel = cli.StringFlag{
		Name: "log-level",
		Usage: "This flag specifies the logger `level(s)`. It can contain multiple comma-separated value. For example" +
			", if set to *:INFO the logs for all packages will have the INFO level. However, if set to *:INFO,api:DEBUG" +
			" the logs for all packages will have the INFO level, excepting the api package which will receive a DEBUG" +
			" log level.",
		Value: "*:" + logger.LogDebug.String(),
	}
	// logFile is used when the log output needs to be logged in a file
	logSaveFile = cli.BoolFlag{
		Name:  "log-save",
		Usage: "Boolean option for enabling log saving. If set, it will automatically save all the logs into a file.",
	}
	log = logger.GetOrCreate("main")
)

func main() {
	app := cli.NewApp()
	cli.AppHelpTemplate = helpTemplate
	app.Name = "Validator statistics to prometheus"
	app.Version = "v1.0.0"
	app.Usage = "Use this tool to fetch validator statistics and serve them in prometheus format"
	app.Flags = []cli.Flag{
		restApiInterfaceFlag,
		proxyURL,
		chain,
		configurationFile,
		logLevel,
		logSaveFile,
	}
	app.Authors = []cli.Author{
		{
			Name:  "The MultiversX Team",
			Email: "contact@multiversx.com",
		},
	}

	app.Action = func(c *cli.Context) error {
		return execute(c)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func execute(ctx *cli.Context) error {
	configPath := ctx.GlobalString(configurationFile.Name)
	cfg, err := loadConfig(configPath)
	if err != nil {
		return err
	}

	logLevelFlagValue := ctx.GlobalString(logLevel.Name)
	err = logger.SetLogLevel(logLevelFlagValue)
	if err != nil {
		return err
	}

	var fileLogging factory.FileLoggingHandler
	withLogFile := ctx.GlobalBool(logSaveFile.Name)
	if withLogFile {
		fileLogging, err = attachFileLogger(log, cfg.Logs)
		if err != nil {
			return err
		}
	}

	proxyUrl := ctx.GlobalString(proxyURL.Name)
	httpClient := http.NewHttpClientWrapper(nil, proxyUrl)
	wrapper, err := httpClientWrapper.NewHttpClientWrapper(httpClient)
	if err != nil {
		return err
	}

	chainField := ctx.GlobalString(chain.Name)
	prometheusCollector, err := collector.NewPrometheusCollector(wrapper, chainField, cfg.Keys)
	if err != nil {
		return err
	}
	registry := prometheus.NewRegistry()
	registry.MustRegister(prometheusCollector)

	ws := gin.Default()
	ws.Use(cors.Default())

	ws.GET("/metrics", gin.WrapH(promhttp.HandlerFor(registry, promhttp.HandlerOpts{})))

	restApiInterface := ctx.GlobalString(restApiInterfaceFlag.Name)
	err = ws.Run(restApiInterface)
	if err != nil {
		return err
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs

	if !check.IfNil(fileLogging) {
		err = fileLogging.Close()
		log.LogIfError(err)
	}

	return nil
}

func loadConfig(filepath string) (*config.Config, error) {
	cfg := &config.Config{}
	err := core.LoadTomlFile(cfg, filepath)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func attachFileLogger(log logger.Logger, logsCfg config.LogsConfig) (factory.FileLoggingHandler, error) {
	workingDir := getWorkingDir(log)
	args := file.ArgsFileLogging{
		WorkingDir:      workingDir,
		DefaultLogsPath: "logs",
		LogFilePrefix:   "multiversx-validator-statistics",
	}
	fileLogging, err := file.NewFileLogging(args)
	if err != nil {
		return nil, fmt.Errorf("%w creating a log file", err)
	}

	timeLogLifeSpan := time.Second * time.Duration(logsCfg.LogFileLifeSpanInSec)
	sizeLogLifeSpanInMB := uint64(logsCfg.LogFileLifeSpanInMB)
	err = fileLogging.ChangeFileLifeSpan(timeLogLifeSpan, sizeLogLifeSpanInMB)
	if err != nil {
		return nil, err
	}

	return fileLogging, nil
}

func getWorkingDir(log logger.Logger) string {
	workingDir, err := os.Getwd()
	if err != nil {
		log.LogIfError(err)
		workingDir = ""
	}

	log.Trace("working directory", "path", workingDir)

	return workingDir
}
