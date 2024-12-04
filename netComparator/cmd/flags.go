package main

import (
	"github.com/urfave/cli"

	"github.com/multiversx/mx-chain-tools-go/netComparator/config"
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

	outDirectory = cli.StringFlag{
		Name: "outDirectory",
		Usage: "This flag specifies in which directory the output files will be stored. " +
			"It consists of an html report with the differences and 2 json files with all the transactions that " +
			"have been compared.",
	}
)

func getFlags() []cli.Flag {
	return []cli.Flag{
		primaryURL,
		secondaryURL,
		timestamp,
		outDirectory,
		number,
	}
}

func getFlagsConfig(ctx *cli.Context) config.ContextFlagsNetComparator {
	flagsConfig := config.ContextFlagsNetComparator{}

	flagsConfig.PrimaryURL = ctx.GlobalString(primaryURL.Name)
	flagsConfig.SecondaryURL = ctx.GlobalString(secondaryURL.Name)
	flagsConfig.Timestamp = ctx.GlobalString(timestamp.Name)
	flagsConfig.OutDirectory = ctx.GlobalString(outDirectory.Name)
	flagsConfig.Number = ctx.GlobalInt(number.Name)

	return flagsConfig
}
