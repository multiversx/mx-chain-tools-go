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
	token = cli.StringFlag{
		Name:  "token",
		Usage: "Name of the token to get the balance from",
		Value: "",
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
		token,
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
	flagsConfig.Token = ctx.GlobalString(token.Name)

	return flagsConfig
}
