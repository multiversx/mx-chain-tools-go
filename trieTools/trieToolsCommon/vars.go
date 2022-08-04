package trieToolsCommon

import (
	"github.com/ElrondNetwork/elrond-go-core/hashing/blake2b"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	elrondConfig "github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

var (
	// Hasher represents the internal hasher used by the node
	Hasher = blake2b.NewBlake2b()
	// Marshaller represents the internal marshaller used by the node
	Marshaller = &marshal.GogoProtoMarshalizer{}

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
)
