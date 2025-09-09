package main

import (
	"github.com/multiversx/mx-chain-tools-go/trieTools/trieToolsCommon"
	"github.com/urfave/cli"
)

func getFlags() []cli.Flag {
	flags := trieToolsCommon.GetFlags()
	newFlags := []cli.Flag{
		StateChangesDBPath,
	}
	return append(flags, newFlags...)
}

type contextFlagsConfig struct {
	trieToolsCommon.ContextFlagsConfig
	StateChangesDBPath string
}

func getFlagsConfig(ctx *cli.Context) contextFlagsConfig {
	flagsConfig := contextFlagsConfig{
		ContextFlagsConfig: trieToolsCommon.GetFlagsConfig(ctx),
		StateChangesDBPath: ctx.GlobalString(StateChangesDBPath.Name),
	}

	return flagsConfig
}
