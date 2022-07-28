package main

import (
	"github.com/ElrondNetwork/elrond-tools-go/trieTools/trieToolsCommon"
	"github.com/urfave/cli"
)

var (
	tokensDirectory = cli.StringFlag{
		Name:  "tokens-dir",
		Usage: "This flag specifies the `directory` where the application will find all exported tokens from all shards",
		Value: "in",
	}
)

type ContextFlagsZeroBalanceSysAcc struct {
	trieToolsCommon.ContextFlagsConfig
	tokensDirectory string
}

func getFlags() []cli.Flag {
	return []cli.Flag{
		trieToolsCommon.WorkingDirectory,
		trieToolsCommon.LogLevel,
		trieToolsCommon.DisableAnsiColor,
		trieToolsCommon.LogSaveFile,
		trieToolsCommon.LogWithLoggerName,
		trieToolsCommon.ProfileMode,
		tokensDirectory,
	}
}

func getFlagsConfig(ctx *cli.Context) ContextFlagsZeroBalanceSysAcc {
	flagsConfig := ContextFlagsZeroBalanceSysAcc{}

	flagsConfig.WorkingDir = ctx.GlobalString(trieToolsCommon.WorkingDirectory.Name)
	flagsConfig.LogLevel = ctx.GlobalString(trieToolsCommon.LogLevel.Name)
	flagsConfig.SaveLogFile = ctx.GlobalBool(trieToolsCommon.LogSaveFile.Name)
	flagsConfig.EnableLogName = ctx.GlobalBool(trieToolsCommon.LogWithLoggerName.Name)
	flagsConfig.EnablePprof = ctx.GlobalBool(trieToolsCommon.ProfileMode.Name)
	flagsConfig.tokensDirectory = ctx.GlobalString(tokensDirectory.Name)

	return flagsConfig
}
