package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/urfave/cli"
)

const (
	logFilePrefix   = "meta-data-remover"
	tomlFile        = "./config.toml"
	outputFilePerms = 0644
)

var (
	log = logger.GetOrCreate("main")
)

func main() {
	app := cli.NewApp()
	app.Name = "Tokens exporter CLI app"
	app.Usage = "This is the entry point for the tool that deletes tokens meta-data"
	app.Authors = []cli.Author{
		{
			Name:  "The Elrond Team",
			Email: "contact@elrond.com",
		},
	}

	app.Action = func(c *cli.Context) error {
		return startProcess(c)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
		return
	}
}

func startProcess(c *cli.Context) error {
	tokensInSysAccount, err := readTokensInput("outputShard1SysAccount.json")
	if err != nil {
		return err
	}
	tokensToDelete, err := readTokensInput("tokens3.json")
	if err != nil {
		return err
	}

	for tokenInSysAcc := range tokensInSysAccount {
		missedToken, found := tokensToDelete[tokenInSysAcc]
		if found {
			log.Error("FOUND IT", "token", missedToken)
		}
	}

	for tokenToDelete := range tokensToDelete {
		missedToken, found := tokensInSysAccount[tokenToDelete]
		if found {
			log.Error("FOUND IT", "token", missedToken)
		}
	}

	return nil

}

func readTokensInput(tokensFile string) (map[string]struct{}, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	fullPath := filepath.Join(workingDir, tokensFile)
	jsonFile, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}

	bytesFromJson, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	shardTokensMap := make(map[string]struct{})
	err = json.Unmarshal(bytesFromJson, &shardTokensMap)
	if err != nil {
		return nil, err
	}

	log.Info("read from input", "file", tokensFile, "num tokens", len(shardTokensMap))
	return shardTokensMap, nil
}
