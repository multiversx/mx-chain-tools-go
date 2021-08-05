package main

import (
	"io/ioutil"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-tools-go/elasticreindexer/config"
	"github.com/ElrondNetwork/elrond-tools-go/elasticreindexer/process"
	"github.com/pelletier/go-toml"
)

const tomlFile = "./config.toml"

var log = logger.GetOrCreate("main")

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Error("cannot load configuration", "error", err)
		return
	}

	reindexer, err := process.CreateReindexer(cfg)
	if err != nil {
		log.Error("cannot create reindexer", "error", err)
		return
	}

	err = reindexer.Process()
	if err != nil {
		log.Error(err.Error())
		return
	}
}

func loadConfig() (*config.GeneralConfig, error) {
	tomlBytes, err := loadBytesFromFile(tomlFile)
	if err != nil {
		return nil, err
	}

	var tc config.GeneralConfig
	err = toml.Unmarshal(tomlBytes, &tc)
	if err != nil {
		return nil, err
	}

	return &tc, nil
}

func loadBytesFromFile(file string) ([]byte, error) {
	return ioutil.ReadFile(file)
}
