package main

import (
	"github.com/ElrondNetwork/elrond-tools-go/miscellaneous/metaDataRemover/config"
	"github.com/ElrondNetwork/elrond-tools-go/trieTools/trieToolsCommon"
	"github.com/urfave/cli"
)

var (
	outfile = cli.StringFlag{
		Name:  "outfile",
		Usage: "This flag specifies where the output will be stored. It consists of a map<address, tokens>",
		Value: "output.json",
	}
	tokens = cli.StringFlag{
		Name:  "tokens",
		Usage: "This flag specifies the input file; it expects the input to be a map<tokens>",
		Value: "tokens.json",
	}
	pems = cli.StringFlag{
		Name:  "pem",
		Usage: "This flag specifies pems file which should be used to sign and send txs",
		Value: "pems",
	}
)

func getFlags() []cli.Flag {
	return []cli.Flag{
		trieToolsCommon.LogLevel,
		trieToolsCommon.DisableAnsiColor,
		trieToolsCommon.LogSaveFile,
		trieToolsCommon.LogWithLoggerName,
		trieToolsCommon.ProfileMode,
		outfile,
		tokens,
		pems,
	}
}

func getFlagsConfig(ctx *cli.Context) config.ContextFlagsMetaDataRemover {
	flagsConfig := config.ContextFlagsMetaDataRemover{}

	flagsConfig.LogLevel = ctx.GlobalString(trieToolsCommon.LogLevel.Name)
	flagsConfig.SaveLogFile = ctx.GlobalBool(trieToolsCommon.LogSaveFile.Name)
	flagsConfig.EnableLogName = ctx.GlobalBool(trieToolsCommon.LogWithLoggerName.Name)
	flagsConfig.EnablePprof = ctx.GlobalBool(trieToolsCommon.ProfileMode.Name)
	flagsConfig.Outfile = ctx.GlobalString(outfile.Name)
	flagsConfig.Tokens = ctx.GlobalString(tokens.Name)
	flagsConfig.Pems = ctx.GlobalString(pems.Name)

	return flagsConfig
}
