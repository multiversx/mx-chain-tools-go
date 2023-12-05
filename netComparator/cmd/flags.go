package main

import (
	"github.com/multiversx/mx-chain-tools-go/netComparator/config"
	"github.com/urfave/cli"
)

var (
	timestamp = cli.StringFlag{
		Name:  "timestamp",
		Usage: "This flag specifies the timestamp after the transactions should be fetched",
	}

	number = cli.IntFlag{
		Name:  "number",
		Usage: "This flag specifies the number of transactions that should be fetched",
		Value: 100,
	}

	outfile = cli.StringFlag{
		Name:  "outfile",
		Usage: "This flag specifies where the output will be stored. It consists of a map<tokens>",
		Value: "output.json",
	}
)

func getFlags() []cli.Flag {
	return []cli.Flag{
		timestamp,
		outfile,
		number,
	}
}

func getFlagsConfig(ctx *cli.Context) config.ContextFlagsNetComparator {
	flagsConfig := config.ContextFlagsNetComparator{}

	flagsConfig.Timestamp = ctx.GlobalString(timestamp.Name)
	flagsConfig.Outfile = ctx.GlobalString(outfile.Name)
	flagsConfig.Number = ctx.GlobalInt(number.Name)

	return flagsConfig
}
