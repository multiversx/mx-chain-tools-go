package main

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/multiversx/mx-chain-tools-go/trieTools/generic/config"
	"github.com/multiversx/mx-chain-tools-go/trieTools/generic/filter"
	"github.com/multiversx/mx-chain-tools-go/trieTools/trieToolsCommon"
)

var (
	limit = cli.Uint64Flag{
		Name:  "limit",
		Usage: "Limit can be used to limit the number of records to a specified amount",
		Value: 1000,
	}

	filters = cli.GenericFlag{
		Name: "filters",
		Usage: "Filters can be used to match a set of resources by specific criteria.\n" +
			"Accepted format is: field=<field_type>,comparator=<cmp_type>,value=<value_type>\n" +
			"field_type: account, balance, nonce, pair, token\n" +
			"cmp_type: eq, ne, gt, lt, ge, le\n" +
			"value_type: either an address/token string, a number or a key:value pair",
		Value: &filtersType{filters: make([]filter.Operation, 0)},
	}

	outfile = cli.StringFlag{
		Name:  "outfile",
		Usage: "This flag specifies where the output will be stored. It consists of a map<address, tokens>",
		Value: "output.json",
	}
)

type filtersType struct {
	filters []filter.Operation
}

func (f *filtersType) Set(input string) error {
	fo, err := filter.ParseFlag(input)
	if err != nil {
		return err
	}

	f.filters = append(f.filters, fo)
	return nil
}

func (f *filtersType) String() string {
	return ""
}

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
		limit,
		filters,
		outfile,
	}
}

func getFlagsConfig(ctx *cli.Context) (*config.ContextFlagsGeneric, error) {
	var err error
	flagsConfig := config.ContextFlagsGeneric{}

	flagsConfig.WorkingDir = ctx.GlobalString(trieToolsCommon.WorkingDirectory.Name)
	flagsConfig.DbDir = ctx.GlobalString(trieToolsCommon.DbDirectory.Name)
	flagsConfig.LogLevel = ctx.GlobalString(trieToolsCommon.LogLevel.Name)
	flagsConfig.SaveLogFile = ctx.GlobalBool(trieToolsCommon.LogSaveFile.Name)
	flagsConfig.EnableLogName = ctx.GlobalBool(trieToolsCommon.LogWithLoggerName.Name)
	flagsConfig.EnablePprof = ctx.GlobalBool(trieToolsCommon.ProfileMode.Name)
	flagsConfig.HexRootHash = ctx.GlobalString(trieToolsCommon.HexRootHash.Name)
	flagsConfig.Limit = ctx.GlobalUint64(limit.Name)
	flagsConfig.Outfile = ctx.GlobalString(outfile.Name)

	flagsConfig.Filters, err = getFiltersFlag(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get flags config: %v", err)
	}

	return &flagsConfig, nil
}

func getFiltersFlag(ctx *cli.Context) ([]filter.Operation, error) {
	in := ctx.GlobalGeneric(filters.Name)
	ft, ok := in.(*filtersType)
	if !ok {
		return nil, fmt.Errorf("failed to parse --filters flag")
	}

	return ft.filters, nil
}
