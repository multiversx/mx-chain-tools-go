package main

import (
	logger "github.com/ElrondNetwork/elrond-go-logger"
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

	cliFlagWorkingDirectory = cli.StringFlag{
		Name:  "working-directory",
		Usage: "This flag specifies the `directory` where the application will use the databases and logs.",
		Value: "",
	}

	cliFlagDbPath = cli.StringFlag{
		Name:     "db-path",
		Usage:    "This flag specifies the `path` where the application will find the trie storage.",
		Required: true,
	}

	cliFlagShard = cli.Uint64Flag{
		Name:     "shard",
		Usage:    "This flag specifies the `shard` ID.",
		Required: true,
	}

	cliFlagNumShards = cli.UintFlag{
		Name:  "num-shards",
		Usage: "Specifies the total number of actual network shards (with the exception of the metachain). Must be 3 for mainnet.",
		Value: 3,
	}

	cliFlagEpoch = cli.Uint64Flag{
		Name:     "epoch",
		Usage:    "This flag specifies the `epoch`.",
		Required: true,
	}

	cliFlagLogLevel = cli.StringFlag{
		Name: "log-level",
		Usage: "This flag specifies the logger `level(s)`. It can contain multiple comma-separated value. For example" +
			", if set to *:INFO the logs for all packages will have the INFO level. However, if set to *:INFO,api:DEBUG" +
			" the logs for all packages will have the INFO level, excepting the api package which will receive a DEBUG" +
			" log level.",
		Value: "*:" + logger.LogDebug.String(),
	}

	cliFlagLogSaveFile = cli.BoolFlag{
		Name:  "log-save",
		Usage: "Boolean option for enabling log saving. If set, it will automatically save all the logs into a file.",
	}

	cliFlagCurrency = cli.StringFlag{
		Name:  "currency",
		Usage: "What balances to export.",
		Value: "EGLD",
	}

	cliFlagCurrencyDecimals = cli.UintFlag{
		Name:  "currency-decimals",
		Usage: "Number of decimals for chosen currency.",
		Value: 18,
	}

	cliFlagExportFormat = cli.StringFlag{
		Name:  "format",
		Usage: "Export format",
		Value: "plain-text",
	}
)

func getAllCliFlags() []cli.Flag {
	return []cli.Flag{
		cliFlagWorkingDirectory,
		cliFlagDbPath,
		cliFlagShard,
		cliFlagNumShards,
		cliFlagEpoch,
		cliFlagLogLevel,
		cliFlagLogSaveFile,
		cliFlagCurrency,
		cliFlagCurrencyDecimals,
		cliFlagExportFormat,
	}
}

type parsedCliFlags struct {
	workingDir       string
	dbPath           string
	shard            uint32
	numShards        uint32
	epoch            uint32
	logLevel         string
	saveLogFile      bool
	currency         string
	currencyDecimals uint32
	exportFormat     string
}

func getParsedCliFlags(ctx *cli.Context) parsedCliFlags {
	return parsedCliFlags{
		workingDir:       ctx.GlobalString(cliFlagWorkingDirectory.Name),
		dbPath:           ctx.GlobalString(cliFlagDbPath.Name),
		shard:            uint32(ctx.GlobalUint64(cliFlagShard.Name)),
		numShards:        uint32(ctx.GlobalUint(cliFlagNumShards.Name)),
		epoch:            uint32(ctx.GlobalUint64(cliFlagEpoch.Name)),
		logLevel:         ctx.GlobalString(cliFlagLogLevel.Name),
		saveLogFile:      ctx.GlobalBool(cliFlagLogSaveFile.Name),
		currency:         ctx.GlobalString(cliFlagCurrency.Name),
		currencyDecimals: uint32(ctx.GlobalUint(cliFlagCurrencyDecimals.Name)),
		exportFormat:     ctx.GlobalString(cliFlagExportFormat.Name),
	}
}
