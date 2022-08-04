package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

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

	generatedKeys, err := generateKeysInParallel(context.Background(), cliFlags, mnemonic, constraints)
	if err != nil {
		return err
	}

	generatedKeys = generatedKeys[:cliFlags.numKeys]
	log.Info("done key generation")

	err = exporter.ExportKeys(generatedKeys)
	if err != nil {
		return err
	}

	return nil
}

func askMnemonic() (data.Mnemonic, error) {
	fmt.Println("Enter an Elrond-compatible mnemonic:")
	line, err := readLine()
	if err != nil {
		return "", err
	}

	mnemonic := data.Mnemonic(line)
	return mnemonic, nil
}

func readLine() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(line), nil
}

func generateKeysInParallel(
	ctx context.Context,
	params parsedCliFlags,
	mnemonic data.Mnemonic,
	constraints *constraints,
) ([]common.GeneratedKey, error) {
	allGeneratedKeys := make([]common.GeneratedKey, 0, params.numKeys)

	numKeys := int(params.numKeys)
	numTasks := params.numTasks
	slidingIndex := params.startIndex

	for len(allGeneratedKeys) < numKeys {
		generatedKeysByTask := make([][]common.GeneratedKey, numTasks)
		errs, _ := errgroup.WithContext(ctx)

		tasks, newSlidingIndex := createTasks(numTasks, slidingIndex, params.useAccountIndex)

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

		err := errs.Wait()
		if err != nil {
			return nil, err
		}

		for _, keys := range generatedKeysByTask {
			allGeneratedKeys = append(allGeneratedKeys, keys...)
		}

		slidingIndex = newSlidingIndex

		log.Info("progress", "numKeys", len(allGeneratedKeys))
	}

	// Sort generated keys by index
	sort.Slice(allGeneratedKeys, func(i, j int) bool {
		return allGeneratedKeys[i].AddressIndex < allGeneratedKeys[j].AddressIndex
	})

	return allGeneratedKeys, nil
}
