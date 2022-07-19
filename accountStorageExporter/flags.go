package main

import (
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-tools-go/accountStorageExporter/config"
	"github.com/urfave/cli"
)

var (
	// workingDirectory defines a flag for the path for the working directory.
	workingDirectory = cli.StringFlag{
		Name:  "working-directory",
		Usage: "This flag specifies the `directory` where the application will use the databases and logs.",
		Value: "",
	}
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
	// logWithLoggerName is used to enable log correlation elements
	logWithLoggerName = cli.BoolFlag{
		Name:  "log-logger-name",
		Usage: "Boolean option for logger name in the logs.",
	}
	// disableAnsiColor defines if the logger subsystem should prevent displaying ANSI colors
	disableAnsiColor = cli.BoolFlag{
		Name:  "disable-ansi-color",
		Usage: "Boolean option for disabling ANSI colors in the logging system.",
	}
	// profileMode defines a flag for profiling the binary
	// If enabled, it will open the pprof routes over the default gin rest webserver.
	// There are several routes that will be available for profiling (profiling can be analyzed with: go tool pprof):
	//  /debug/pprof/ (can be accessed in the browser, will list the available options)
	//  /debug/pprof/goroutine
	//  /debug/pprof/heap
	//  /debug/pprof/threadcreate
	//  /debug/pprof/block
	//  /debug/pprof/mutex
	//  /debug/pprof/profile (CPU profile)
	//  /debug/pprof/trace?seconds=5 (CPU trace) -> being a trace, can be analyzed with: go tool trace
	// Usage: go tool pprof http(s)://ip.of.the.server/debug/pprof/xxxxx
	profileMode = cli.BoolFlag{
		Name: "profile-mode",
		Usage: "Boolean option for enabling the profiling mode. If set, the /debug/pprof routes will be available " +
			"on the node for profiling the application.",
	}
	// dbDirectory defines a flag for the db path inside the working directory.
	dbDirectory = cli.StringFlag{
		Name:  "db-directory",
		Usage: "This flag specifies the `directory` where the application will find the trie storage.",
		Value: "dbsh2",
	}
	// hexRootHash defines a flag for the trie root hash expressed in hex format
	hexRootHash = cli.StringFlag{
		Name:  "hex-roothash",
		Usage: "This flag specifies the roothash to start the checking from",
		Value: "fe4ae44105be738f1d1541d283cabd1765b2246a8bd0b0ad9a2ffe10c6802576",
	}
	// address defines a flag that specifies the bech32 address of the account to fetch the storage for
	address = cli.StringFlag{
		Name:  "address",
		Usage: "This flag specifies the bech32 address to fetch the storage for",
		Value: "erd1qqqqqqqqqqqqqpgqh664xhyxeqhmzxzhn75aa2yfy33hd58fhneql8e9ee",
	}
)

func getFlags() []cli.Flag {
	return []cli.Flag{
		workingDirectory,
		dbDirectory,
		logLevel,
		disableAnsiColor,
		logSaveFile,
		logWithLoggerName,
		profileMode,
		hexRootHash,
		address,
	}
}

func getFlagsConfig(ctx *cli.Context) config.ContextFlagsConfig {
	flagsConfig := config.ContextFlagsConfig{}

	flagsConfig.WorkingDir = ctx.GlobalString(workingDirectory.Name)
	flagsConfig.DbDir = ctx.GlobalString(dbDirectory.Name)
	flagsConfig.LogLevel = ctx.GlobalString(logLevel.Name)
	flagsConfig.DisableAnsiColor = ctx.GlobalBool(disableAnsiColor.Name)
	flagsConfig.SaveLogFile = ctx.GlobalBool(logSaveFile.Name)
	flagsConfig.EnableLogName = ctx.GlobalBool(logWithLoggerName.Name)
	flagsConfig.EnablePprof = ctx.GlobalBool(profileMode.Name)
	flagsConfig.HexRootHash = ctx.GlobalString(hexRootHash.Name)
	flagsConfig.Address = ctx.GlobalString(address.Name)

	return flagsConfig
}
