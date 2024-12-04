package trie

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/hashing/blake2b"
	"github.com/multiversx/mx-chain-core-go/marshal"
	nodeCommon "github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/epochStart/notifier"
	"github.com/multiversx/mx-chain-go/sharding"
	storageFactory "github.com/multiversx/mx-chain-go/storage/factory"
	"github.com/multiversx/mx-chain-go/storage/pruning"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/multiversx/mx-chain-tools-go/trieTools/balancesExporter/common"
	"github.com/multiversx/mx-chain-tools-go/trieTools/trieToolsCommon"
)

const (
	maxTrieLevelInMemory  = 5
	maxBatchSize          = 45000
	storageUnitIdentifier = "AccountsTrie"
)

var (
	hasher     = blake2b.NewBlake2b()
	marshaller = &marshal.GogoProtoMarshalizer{}
)

// ArgsNewTrieFactory holds arguments for creating a trieFactory
type ArgsNewTrieFactory struct {
	ShardCoordinator sharding.Coordinator
	DbPath           string
	Epoch            uint32
}

type trieFactory struct {
	shardCoordinator sharding.Coordinator
	dbPath           string
	epoch            uint32
}

// NewTrieFactory creates a new trieFactory
func NewTrieFactory(args ArgsNewTrieFactory) *trieFactory {
	return &trieFactory{
		shardCoordinator: args.ShardCoordinator,
		dbPath:           args.DbPath,
		epoch:            args.Epoch,
	}
}

// CreateTrie creates a trie (actually, a wrapper over the actual trie)
func (factory *trieFactory) CreateTrie() (*trieWrapper, error) {
	cacheConfig := getCacheConfig()
	dbConfig := getDbConfig(factory.dbPath)
	pathManager := common.NewSimplePathManager(factory.dbPath)

	persisterFactory, err := storageFactory.NewPersisterFactory(dbConfig)
	if err != nil {
		return nil, err
	}

	args := pruning.StorerArgs{
		Identifier:             storageUnitIdentifier,
		ShardCoordinator:       factory.shardCoordinator,
		CacheConf:              cacheConfig,
		PathManager:            pathManager,
		DbPath:                 "",
		PersisterFactory:       persisterFactory,
		Notifier:               notifier.NewManualEpochStartNotifier(),
		OldDataCleanerProvider: &testscommon.OldDataCleanerProviderStub{},
		CustomDatabaseRemover:  &testscommon.CustomDatabaseRemoverStub{},
		MaxBatchSize:           maxBatchSize,
		EpochsData: pruning.EpochArgs{
			NumOfEpochsToKeep:     factory.epoch + 1,
			NumOfActivePersisters: factory.epoch + 1,
			StartingEpoch:         factory.epoch,
		},
		PruningEnabled:            true,
		EnabledDbLookupExtensions: false,
	}

	db, err := pruning.NewTriePruningStorer(args)
	if err != nil {
		return nil, err
	}

	storageManager, err := trieToolsCommon.CreateStorageManager(db)
	if err != nil {
		return nil, err
	}

	enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
			return flag == nodeCommon.AutoBalanceDataTriesFlag ||
				flag == nodeCommon.DynamicESDTFlag
		},
	}

	t, err := trie.NewTrie(storageManager, marshaller, hasher, enableEpochsHandler, maxTrieLevelInMemory)
	if err != nil {
		return nil, err
	}

	return newTrieWrapper(t), nil
}
