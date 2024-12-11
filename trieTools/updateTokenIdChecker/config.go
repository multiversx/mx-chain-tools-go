package main

import (
	"github.com/multiversx/mx-chain-tools-go/trieTools/trieToolsCommon"
	"github.com/urfave/cli"
)

// ShardID is a type for the shard id
type ShardID string

const (
	Shard0 ShardID = "shard0"
	Shard1 ShardID = "shard1"
	Shard2 ShardID = "shard2"
	Meta   ShardID = "meta"
)

func getFlags() []cli.Flag {
	flags := trieToolsCommon.GetFlags()
	newFlags := []cli.Flag{
		HexRootHash0,
		HexRootHash1,
		HexRootHash2,
		HexRootHashMeta,
	}
	return append(flags, newFlags...)
}

type contextFlagsConfig struct {
	trieToolsCommon.ContextFlagsConfig
	HexRootHash0    string
	HexRootHash1    string
	HexRootHash2    string
	HexRootHashMeta string
}

func getFlagsConfig(ctx *cli.Context) contextFlagsConfig {
	flagsConfig := contextFlagsConfig{
		ContextFlagsConfig: trieToolsCommon.GetFlagsConfig(ctx),
		HexRootHash0:       ctx.GlobalString(HexRootHash0.Name),
		HexRootHash1:       ctx.GlobalString(HexRootHash1.Name),
		HexRootHash2:       ctx.GlobalString(HexRootHash2.Name),
		HexRootHashMeta:    ctx.GlobalString(HexRootHashMeta.Name),
	}

	return flagsConfig
}
