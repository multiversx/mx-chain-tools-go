package main

import (
	"io/ioutil"
	"os"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-tools-go/tgbot/config"
	"github.com/ElrondNetwork/elrond-tools-go/tgbot/process"
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
			Name:  "The Elrond Team",
			Email: "contact@elrond.com",
		},
	}

	app.Action = startTelegramBot

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func startTelegramBot(c *cli.Context) {
	cfg, err := loadConfig()
	if err != nil {
		log.Error("cannot load configuration", "error", err)
		return
	}

	notifier, err := process.NewBalanceNotifier(cfg)
	if err != nil {
		log.Error("cannot start balance notifier", "error", err)
		return
	}

	notifier.StartNotifier()
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
