package main

import (
	"context"
	"os"
	"sort"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	"github.com/ElrondNetwork/elrond-tools-go/hdKeysGenerator/common"
	"github.com/ElrondNetwork/elrond-tools-go/hdKeysGenerator/export"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"
)

const (
	appVersion = "1.0.0"
)

var log = logger.GetOrCreate("main")

func main() {
	app := cli.NewApp()
	cli.AppHelpTemplate = helpTemplate
	app.Version = appVersion
	app.Name = "HD keys generator app"
	app.Usage = "Tool for generating (deriving) HD keys from a given mnemonic"
	app.Flags = getAllCliFlags()
	app.Authors = []cli.Author{
		{
			Name:  "The Elrond Team",
			Email: "contact@elrond.com",
		},
	}

	app.Action = generateKeys

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func generateKeys(ctx *cli.Context) error {
	cliFlags := getParsedCliFlags(ctx)

	constraints, err := newConstraints(cliFlags.numShards, cliFlags.actualShard, cliFlags.projectedShard)
	if err != nil {
		return err
	}

	exporter, err := export.NewExporter(export.ArgsNewExporter{
		ActualShard:    cliFlags.actualShard,
		ProjectedShard: cliFlags.projectedShard,
		StartIndex:     cliFlags.startIndex,
		NumKeys:        int(cliFlags.numKeys),
		Format:         cliFlags.exportFormat,
	})
	if err != nil {
		return err
	}

	mnemonic, err := askMnemonic()
	if err != nil {
		return err
	}

	args := argsGenerateKeysInParallel{
		numTasks:        cliFlags.numTasks,
		startIndex:      cliFlags.startIndex,
		useAccountIndex: cliFlags.useAccountIndex,
		numKeys:         cliFlags.numTasks,
	}

	generatedKeys, err := generateKeysInParallel(context.Background(), args, mnemonic, constraints)
	if err != nil {
		return err
	}

	log.Info("done key generation")

	err = exporter.ExportKeys(generatedKeys)
	if err != nil {
		return err
	}

	return nil
}

func generateKeysInParallel(
	ctx context.Context,
	args argsGenerateKeysInParallel,
	mnemonic data.Mnemonic,
	constraints *constraints,
) ([]common.GeneratedKey, error) {
	allGeneratedKeys := make([]common.GeneratedKey, 0, args.numKeys)

	numKeys := int(args.numKeys)
	numTasks := args.numTasks
	slidingIndex := args.startIndex

	// Description of the parallel model:
	// In a loop, as long as there are keys to be generated:
	//	- a number of "numTasks" (parallel) tasks are created.
	// 	- each task checks a number of "fixedTaskSize" indexes account / address indexes for eligibility.
	//  - the output (selected indexes, generated keys) is accumulated in a slice
	// At the end of the loop, the accumulator slice is sorted by account & address indexes,
	// while redundantly-generated keys are ignored.
	//
	// This parallelization model leads to some redundant work when a small number of keys are requested
	// and / or the generation constraints are "easy" (they do not imply a lot of index skipping).
	// However, the model should behave well when a lot of account / address indexes have to be checked for eligibility.
	for len(allGeneratedKeys) < numKeys {
		generatedKeysByTask := make([][]common.GeneratedKey, numTasks)
		tasks, newSlidingIndex := createTasks(argsCreateTasks{
			numTasks:        numTasks,
			startIndex:      slidingIndex,
			useAccountIndex: args.useAccountIndex,
		})

		errs, _ := errgroup.WithContext(ctx)

		for taskIndex, task := range tasks {
			i := taskIndex
			t := task

			errs.Go(func() error {
				keys, err := t.doGenerateKeys(mnemonic, constraints)
				if err != nil {
					return err
				}

				generatedKeysByTask[i] = keys
				return nil
			})
		}

		// Wait for all tasks
		err := errs.Wait()
		if err != nil {
			return nil, err
		}

		// Gather output from all tasks
		for _, keys := range generatedKeysByTask {
			allGeneratedKeys = append(allGeneratedKeys, keys...)
		}

		slidingIndex = newSlidingIndex

		log.Info("progress", "numKeys", len(allGeneratedKeys))
	}

	inlineSortKeysByIndexes(allGeneratedKeys)

	// Ignore redundantly-generated keys
	allGeneratedKeys = allGeneratedKeys[:numKeys]
	return allGeneratedKeys, nil
}

func inlineSortKeysByIndexes(keys []common.GeneratedKey) {
	sort.Slice(keys, func(i, j int) bool {
		a := keys[i]
		b := keys[j]

		if a.AccountIndex != b.AccountIndex {
			return a.AccountIndex < b.AccountIndex
		}

		return a.AddressIndex < b.AddressIndex
	})
}
