package trieToolsCommon

import (
	"github.com/multiversx/mx-chain-core-go/hashing/blake2b"
	"github.com/multiversx/mx-chain-core-go/marshal"
	nodeConfig "github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
)

var (
	// Hasher represents the internal hasher used by the node
	Hasher = blake2b.NewBlake2b()
	// Marshaller represents the internal marshaller used by the node
	Marshaller = &marshal.GogoProtoMarshalizer{}

	cacheConfig = storageunit.CacheConfig{
		Type:        "SizeLRU",
		Capacity:    500000,
		SizeInBytes: 314572800, // 300MB
	}
	dbConfig = nodeConfig.DBConfig{
		Type:              "LvlDBSerial",
		BatchDelaySeconds: 2,
		MaxBatchSize:      45000,
		MaxOpenFiles:      10,
	}
)
