package main

import (
	"runtime"

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

	cliFlagActualShard = cli.Uint64Flag{
		Name:  "actual-shard",
		Usage: "Generate keys in a given actual shard.",
	}

	cliFlagProjectedShard = cli.Uint64Flag{
		Name:  "projected-shard",
		Usage: "Generate keys in a given projected shard.",
	}

	cliFlagNumShards = cli.UintFlag{
		Name:  "num-shards",
		Usage: "Specifies the total number of actual network shards (with the exception of the metachain). Must be 3 for mainnet.",
		Value: 3,
	}

	cliFlagNumKeys = cli.UintFlag{
		Name:     "num-keys",
		Usage:    "Specifies the total number of keys to generate.",
		Required: true,
	}

	cliFlagStartIndex = cli.UintFlag{
		Name:  "start-index",
		Usage: "Specifies the start `index` (`address index` or `account index`).",
		Value: 0,
	}

	cliFlagUseAccountIndex = cli.BoolFlag{
		Name:  "use-account-index",
		Usage: "Use the `account index` parameter for deriving keys, instead of using the `address index` (default).",
	}

	cliFlagNumTasks = cli.IntFlag{
		Name:  "num-tasks",
		Usage: "How many tasks to use (parallelization level).",
		Value: runtime.NumCPU(),
	}
)

func getAllCliFlags() []cli.Flag {
	return []cli.Flag{
		cliFlagNumShards,
		cliFlagActualShard,
		cliFlagProjectedShard,
		cliFlagNumKeys,
		cliFlagStartIndex,
		cliFlagUseAccountIndex,
		cliFlagNumTasks,
	}
}

type parsedCliFlags struct {
	numShards       uint32
	actualShard     optionalUint32
	projectedShard  optionalUint32
	numKeys         uint
	startIndex      int
	useAccountIndex bool
	numTasks        int
}

func getParsedCliFlags(ctx *cli.Context) parsedCliFlags {
	return parsedCliFlags{
		numShards: uint32(ctx.GlobalUint(cliFlagNumShards.Name)),
		actualShard: optionalUint32{
			Value:    uint32(ctx.GlobalUint64(cliFlagActualShard.Name)),
			HasValue: ctx.GlobalIsSet(cliFlagActualShard.Name),
		},
		projectedShard: optionalUint32{
			Value:    uint32(ctx.GlobalUint64(cliFlagProjectedShard.Name)),
			HasValue: ctx.GlobalIsSet(cliFlagProjectedShard.Name),
		},
		numKeys:         ctx.GlobalUint(cliFlagNumKeys.Name),
		startIndex:      ctx.GlobalInt(cliFlagStartIndex.Name),
		useAccountIndex: ctx.GlobalBool(cliFlagUseAccountIndex.Name),
		numTasks:        ctx.GlobalInt(cliFlagNumTasks.Name),
	}
}
