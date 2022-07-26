package main

import (
	logger "github.com/ElrondNetwork/elrond-go-logger"
	elrondConfig "github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

var (
	cacheConfig = storageUnit.CacheConfig{
		Type:        "SizeLRU",
		Capacity:    500000,
		SizeInBytes: 314572800, // 300MB
	}
	dbConfig = elrondConfig.DBConfig{
		Type:              "LvlDBSerial",
		BatchDelaySeconds: 2,
		MaxBatchSize:      45000,
		MaxOpenFiles:      10,
	}
	log = logger.GetOrCreate("main")
)
