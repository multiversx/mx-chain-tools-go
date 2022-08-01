package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	"github.com/urfave/cli"
)

const (
	appVersion = "1.0.0"
)

var log = logger.GetOrCreate("main")

func main() {
	app := cli.NewApp()
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

	mnemonic, err := askMnemonic()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	generatedKeysByTask := make([][]generatedKey, cliFlags.numTasks)

	for taskIndex := 0; taskIndex < cliFlags.numTasks; taskIndex++ {
		wg.Add(1)

		task := generatorTask{
			numTasks:        cliFlags.numTasks,
			taskIndex:       taskIndex,
			mnemonic:        mnemonic,
			constraints:     constraints,
			useAccountIndex: cliFlags.useAccountIndex,
			startIndex:      cliFlags.startIndex,
			numKeys:         int(cliFlags.numKeys)/cliFlags.numTasks + 1,
		}

		go func(t generatorTask) {
			generatedKeysByTask[t.taskIndex], err = t.doGenerateKeys()
			if err != nil {
				log.Error(err.Error())
				return
			}

			wg.Done()
		}(task)
	}

	wg.Wait()

	allGeneratedKeys := make([]generatedKey, 0, cliFlags.numKeys)

	for _, keys := range generatedKeysByTask {
		allGeneratedKeys = append(allGeneratedKeys, keys...)
	}

	sort.Slice(allGeneratedKeys, func(i, j int) bool {
		return allGeneratedKeys[i].Index < allGeneratedKeys[j].Index
	})

	fmt.Println("Generated", len(allGeneratedKeys))
	fmt.Println("...")

	for _, key := range allGeneratedKeys {
		fmt.Println(key.Index, key.Address)
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
