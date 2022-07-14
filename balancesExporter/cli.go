package main

import (
	"fmt"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-tools-go/balancesExporter/export"
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

	cliFlagDbPath = cli.StringFlag{
		Name:     "db-path",
		Usage:    "The path to a node's database.",
		Required: true,
	}

	cliFlagShard = cli.Uint64Flag{
		Name:     "shard",
		Usage:    "The shard to use for export.",
		Required: true,
	}

	cliFlagProjectedShard = cli.Uint64Flag{
		Name:     "projected-shard",
		Usage:    "The projected shard to use for export.",
		Required: false,
	}

	cliFlagNumShards = cli.UintFlag{
		Name:  "num-shards",
		Usage: "Specifies the total number of actual network shards (with the exception of the metachain). Must be 3 for mainnet.",
		Value: 3,
	}

	cliFlagEpoch = cli.Uint64Flag{
		Name:     "epoch",
		Usage:    "The epoch to use for export.",
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
		Usage: fmt.Sprintf("Export format. One of the following: %s", export.AllFormattersNames),
		Value: export.FormatterNamePlainText,
	}

	cliFlagWithContracts = cli.BoolFlag{
		Name:  "with-contracts",
		Usage: "Whether to include contracts in the export.",
	}

	cliFlagWithZero = cli.BoolFlag{
		Name:  "with-zero",
		Usage: "Whether to include accounts with zero balance in the export.",
	}
)

func getAllCliFlags() []cli.Flag {
	return []cli.Flag{
		cliFlagDbPath,
		cliFlagShard,
		cliFlagProjectedShard,
		cliFlagNumShards,
		cliFlagEpoch,
		cliFlagLogLevel,
		cliFlagLogSaveFile,
		cliFlagCurrency,
		cliFlagCurrencyDecimals,
		cliFlagExportFormat,
		cliFlagWithContracts,
		cliFlagWithZero,
	}
}

type parsedCliFlags struct {
	dbPath              string
	shard               uint32
	projectedShard      uint32
	projectedShardIsSet bool
	numShards           uint32
	epoch               uint32
	logLevel            string
	saveLogFile         bool
	currency            string
	currencyDecimals    uint
	exportFormat        string
	withContracts       bool
	withZero            bool
}

func getParsedCliFlags(ctx *cli.Context) parsedCliFlags {
	return parsedCliFlags{
		dbPath:              ctx.GlobalString(cliFlagDbPath.Name),
		shard:               uint32(ctx.GlobalUint64(cliFlagShard.Name)),
		projectedShard:      uint32(ctx.GlobalUint64(cliFlagProjectedShard.Name)),
		projectedShardIsSet: ctx.GlobalIsSet(cliFlagProjectedShard.Name),
		numShards:           uint32(ctx.GlobalUint(cliFlagNumShards.Name)),
		epoch:               uint32(ctx.GlobalUint64(cliFlagEpoch.Name)),
		logLevel:            ctx.GlobalString(cliFlagLogLevel.Name),
		saveLogFile:         ctx.GlobalBool(cliFlagLogSaveFile.Name),
		currency:            ctx.GlobalString(cliFlagCurrency.Name),
		currencyDecimals:    uint(ctx.GlobalUint(cliFlagCurrencyDecimals.Name)),
		exportFormat:        ctx.GlobalString(cliFlagExportFormat.Name),
		withContracts:       ctx.GlobalBool(cliFlagWithContracts.Name),
		withZero:            ctx.GlobalBool(cliFlagWithZero.Name),
	}
}
