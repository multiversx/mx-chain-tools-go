package main

import (
	"io/ioutil"
	"os"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-tools-go/sccounter/config"
	"github.com/ElrondNetwork/elrond-tools-go/sccounter/process"
	"github.com/pelletier/go-toml"
	"github.com/urfave/cli"
)

const tomlFile = "./config.toml"

var (
	log = logger.GetOrCreate("main")

	// wasmFile defines a string flag with the path to the contract wasm file
	wasmFile = cli.StringFlag{
		Name:  "wasm-file",
		Usage: "Path to the contract wasm file",
	}
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
	app.Name = "SC counter CLI App"
	app.Version = "v1.0.0"
	app.Usage = "This is the entry point for the Smart Contract counter tool"
	app.Flags = []cli.Flag{
		wasmFile,
	}
	app.Authors = []cli.Author{
		{
			Name:  "The Elrond Team",
			Email: "contact@elrond.com",
		},
	}

	app.Action = startSCCounter

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func startSCCounter(ctx *cli.Context) {
	cfg, err := loadConfig()
	if err != nil {
		log.Error("cannot load configuration", "error", err)
		return
	}

	scCounter, err := process.CreateSCCounter(cfg)
	if err != nil {
		log.Error("cannot create sccounter", "error", err)
		return
	}

	err = scCounter.ProcessSCWasm(ctx.String(wasmFile.Name))
	if err != nil {
		log.Error("cannot get number of sc deploys", "error", err)
		return
	}
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
