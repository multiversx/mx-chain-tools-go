package main

import (
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"

	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-tools-go/tgbot/config"
	"github.com/multiversx/mx-chain-tools-go/tgbot/process"
	"github.com/pelletier/go-toml"
	"github.com/urfave/cli"
)

const tomlFile = "./config.toml"

var (
	log = logger.GetOrCreate("main")
)

const helpTemplate = `NAME:
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

func main() {
	app := cli.NewApp()
	cli.AppHelpTemplate = helpTemplate
	app.Name = "Bot balance notifier"
	app.Version = "v1.0.0"
	app.Usage = "This is the entry point for balance notifier tool"
	app.Authors = []cli.Author{
		{
			Name:  "The Multiversx Team",
			Email: "contact@multiversx.com",
		},
	}

	app.Action = startTelegramBot

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func startTelegramBot(_ *cli.Context) error {
	cfg, err := loadConfig()
	if err != nil {
		log.Error("cannot load configuration", "error", err)
		return nil
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	var notifiers []io.Closer
	for _, botCfg := range cfg.BotConfigs {
		notifier, errC := process.NewBalanceNotifier(botCfg)
		if errC != nil {
			log.Error("cannot start balance notifier", "error", errC)
			return nil
		}
		notifiers = append(notifiers, notifier)

		go notifier.StartNotifier()
	}

	<-interrupt
	log.Info("closing app at user's signal")
	for _, notifier := range notifiers {
		_ = notifier.Close()
	}

	time.Sleep(1 * time.Millisecond)
	return nil
}

func loadConfig() (*config.GeneralConfig, error) {
	tomlBytes, err := loadBytesFromFile(tomlFile)
	if err != nil {
		return nil, err
	}

	var gc config.GeneralConfig
	err = toml.Unmarshal(tomlBytes, &gc)
	if err != nil {
		return nil, err
	}

	return &gc, nil
}

func loadBytesFromFile(file string) ([]byte, error) {
	return ioutil.ReadFile(file)
}
