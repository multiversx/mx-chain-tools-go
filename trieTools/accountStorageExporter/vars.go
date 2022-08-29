package main

import (
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

var (
	log             = logger.GetOrCreate("main")
	outputFileName  = "output.json"
	outputFilePerms = 0644
)
