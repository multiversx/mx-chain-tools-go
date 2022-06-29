package main

import (
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-tools-go/balancesExporter/config"
	"github.com/urfave/cli"
)

var (
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
)

func getCliFlags() []cli.Flag {
	return []cli.Flag{
		cliFlagWorkingDirectory,
		cliFlagDbPath,
		cliFlagShard,
		cliFlagEpoch,
		cliFlagLogLevel,
		cliFlagLogSaveFile,
	}
}

func getFlagsConfig(ctx *cli.Context) config.ContextFlagsConfig {
	flagsConfig := config.ContextFlagsConfig{}

	flagsConfig.WorkingDir = ctx.GlobalString(cliFlagWorkingDirectory.Name)
	flagsConfig.DbPath = ctx.GlobalString(cliFlagDbPath.Name)
	flagsConfig.Shard = uint32(ctx.GlobalUint64(cliFlagShard.Name))
	flagsConfig.Epoch = uint32(ctx.GlobalUint64(cliFlagEpoch.Name))
	flagsConfig.LogLevel = ctx.GlobalString(cliFlagLogLevel.Name)
	flagsConfig.SaveLogFile = ctx.GlobalBool(cliFlagLogSaveFile.Name)

	return flagsConfig
}
