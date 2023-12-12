package main

import (
	"github.com/multiversx/mx-chain-tools-go/netComparator/config"
	"github.com/urfave/cli"
)

var (
	primaryURL = cli.StringFlag{
		Name:  "primary-url",
		Usage: "This flag specifies the primary network's URL.",
	}

	secondaryURL = cli.StringFlag{
		Name:  "secondary-url",
		Usage: "This flag specifies the secondary network's URL.",
	}

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
		Usage: "This flag specifies where the output will be stored. It consists of an html report with the differences.",
		Value: "index.html",
	}
)

func getFlags() []cli.Flag {
	return []cli.Flag{
		primaryURL,
		secondaryURL,
		timestamp,
		outfile,
		number,
	}
}

func getFlagsConfig(ctx *cli.Context) config.ContextFlagsNetComparator {
	flagsConfig := config.ContextFlagsNetComparator{}

	flagsConfig.PrimaryURL = ctx.GlobalString(primaryURL.Name)
	flagsConfig.SecondaryURL = ctx.GlobalString(secondaryURL.Name)
	flagsConfig.Timestamp = ctx.GlobalString(timestamp.Name)
	flagsConfig.Outfile = ctx.GlobalString(outfile.Name)
	flagsConfig.Number = ctx.GlobalInt(number.Name)

	return flagsConfig
}
