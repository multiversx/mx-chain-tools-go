package main

import (
	"github.com/ElrondNetwork/elrond-tools-go/trieTools/tokensExporter/config"
	"github.com/ElrondNetwork/elrond-tools-go/trieTools/trieToolsCommon"
	"github.com/urfave/cli"
)

var (
	outfile = cli.StringFlag{
		Name:  "outfile",
		Usage: "This flag specifies where the output will be stored. It consists of a map<address, tokens>",
		Value: "output.json",
	}

	flagContracts = cli.BoolFlag{
		Name: "contracts",
	}

	flagTokens = cli.StringSliceFlag{
		Name:  "tokens",
		Usage: "Specifies the symbols of enabled custom currencies (i.e. ESDT identifiers).",
		Value: &cli.StringSlice{},
	}
)

func getFlags() []cli.Flag {
	return []cli.Flag{
		trieToolsCommon.WorkingDirectory,
		trieToolsCommon.DbDirectory,
		trieToolsCommon.LogLevel,
		trieToolsCommon.DisableAnsiColor,
		trieToolsCommon.LogSaveFile,
		trieToolsCommon.LogWithLoggerName,
		trieToolsCommon.ProfileMode,
		trieToolsCommon.HexRootHash,
		outfile,
		flagContracts,
		flagTokens,
	}
}

func getFlagsConfig(ctx *cli.Context) config.ContextFlagsTokensExporter {
	flagsConfig := config.ContextFlagsTokensExporter{}

	flagsConfig.WorkingDir = ctx.GlobalString(trieToolsCommon.WorkingDirectory.Name)
	flagsConfig.DbDir = ctx.GlobalString(trieToolsCommon.DbDirectory.Name)
	flagsConfig.LogLevel = ctx.GlobalString(trieToolsCommon.LogLevel.Name)
	flagsConfig.SaveLogFile = ctx.GlobalBool(trieToolsCommon.LogSaveFile.Name)
	flagsConfig.EnableLogName = ctx.GlobalBool(trieToolsCommon.LogWithLoggerName.Name)
	flagsConfig.EnablePprof = ctx.GlobalBool(trieToolsCommon.ProfileMode.Name)
	flagsConfig.HexRootHash = ctx.GlobalString(trieToolsCommon.HexRootHash.Name)
	flagsConfig.Outfile = ctx.GlobalString(outfile.Name)
	flagsConfig.ExportContracts = ctx.GlobalBool(flagContracts.Name)
	flagsConfig.Tokens = ctx.GlobalStringSlice(flagTokens.Name)

	return flagsConfig
}
