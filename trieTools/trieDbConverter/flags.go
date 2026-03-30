package main

import (
	"github.com/multiversx/mx-chain-tools-go/trieTools/trieToolsCommon"
	"github.com/urfave/cli"
)

const (
	staticType                   = "static"
	pruningDbType                = "pruning"
	defaultMaxTrieCopyGoroutines = 100
)

var (
	originDbType = cli.StringFlag{
		Name:  "origin-db-type",
		Usage: "This flag specifies the type of the origin db.",
		Value: staticType,
	}
	targetDbType = cli.StringFlag{
		Name:  "target-db-type",
		Usage: "This flag specifies the type of the target db.",
		Value: staticType,
	}
	maxGoroutines = cli.IntFlag{
		Name:  "max-goroutines",
		Usage: "Maximum number of goroutines used to copy the main trie and the data tries.",
		Value: defaultMaxTrieCopyGoroutines,
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
		originDbType,
		maxGoroutines,
		targetDbType,
	}
}

func getFlagsConfig(ctx *cli.Context) ContextFlagsDbConverter {
	flagsConfig := ContextFlagsDbConverter{}

	flagsConfig.WorkingDir = ctx.GlobalString(trieToolsCommon.WorkingDirectory.Name)
	flagsConfig.DbDir = ctx.GlobalString(trieToolsCommon.DbDirectory.Name)
	flagsConfig.LogLevel = ctx.GlobalString(trieToolsCommon.LogLevel.Name)
	flagsConfig.SaveLogFile = ctx.GlobalBool(trieToolsCommon.LogSaveFile.Name)
	flagsConfig.EnableLogName = ctx.GlobalBool(trieToolsCommon.LogWithLoggerName.Name)
	flagsConfig.EnablePprof = ctx.GlobalBool(trieToolsCommon.ProfileMode.Name)
	flagsConfig.HexRootHash = ctx.GlobalString(trieToolsCommon.HexRootHash.Name)
	flagsConfig.OriginDbType = ctx.GlobalString(originDbType.Name)
	flagsConfig.MaxGoroutines = ctx.GlobalInt(maxGoroutines.Name)
	flagsConfig.TargetDbType = ctx.GlobalString(targetDbType.Name)

	return flagsConfig
}

// ContextFlagsDbConverter is the flags config for db converter
type ContextFlagsDbConverter struct {
	trieToolsCommon.ContextFlagsConfig
	TargetDb      string
	OriginDbType  string
	MaxGoroutines int
	TargetDbType  string
}
