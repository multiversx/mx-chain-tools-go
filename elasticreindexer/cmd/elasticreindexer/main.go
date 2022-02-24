package main

import (
	"io/ioutil"
	"os"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-tools-go/elasticreindexer/config"
	"github.com/ElrondNetwork/elrond-tools-go/elasticreindexer/process"
	"github.com/pelletier/go-toml"
	"github.com/urfave/cli"
)

const tomlFile = "./config.toml"

var (
	log = logger.GetOrCreate("main")

	// overwrite defines a bool flag for overwriting destination's data and skipping mappings and aliases checks
	overwriteFlag = cli.BoolFlag{
		Name:  "overwrite",
		Usage: "If set, the reindexing tool will skip the creation of the index and mapping and will overwrite any existing data.",
	}
	// skipMappingsFlag defines a bool flag for skipping the copying of the source's mapping into the destination
	skipMappingsFlag = cli.BoolFlag{
		Name:  "skip-mappings",
		Usage: "If set, the reindexing tool will skip the copying of the mappings",
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
	app.Name = "Elasticsearch reindexing CLI App"
	app.Version = "v1.0.0"
	app.Usage = "This is the entry point for Elasticsearch reindexing tool"
	app.Flags = []cli.Flag{
		overwriteFlag,
		skipMappingsFlag,
	}
	app.Authors = []cli.Author{
		{
			Name:  "The Elrond Team",
			Email: "contact@elrond.com",
		},
	}

	app.Action = startReindexing

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func startReindexing(ctx *cli.Context) {
	cfg, err := loadConfig()
	if err != nil {
		log.Error("cannot load configuration", "error", err)
		return
	}

	reindexer, err := process.CreateReindexer(cfg)
	if err != nil {
		log.Error("cannot create reindexer", "error", err)
		return
	}

	skipMappings := ctx.Bool(skipMappingsFlag.Name)
	err = reindexer.Process(ctx.Bool(overwriteFlag.Name), skipMappings)
	if err != nil {
		log.Error(err.Error())
		return
	}
}

func loadConfig() (*config.GeneralConfig, error) {
	tomlBytes, err := loadBytesFromFile(tomlFile)
	if err != nil {
		return nil, err
	}

	var tc config.GeneralConfig
	err = toml.Unmarshal(tomlBytes, &tc)
	if err != nil {
		return nil, err
	}

	return &tc, nil
}

func loadBytesFromFile(file string) ([]byte, error) {
	return ioutil.ReadFile(file)
}
