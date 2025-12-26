package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/throttler"
	"github.com/multiversx/mx-chain-core-go/hashing/blake2b"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	commonDisabled "github.com/multiversx/mx-chain-go/common/disabled"
	"github.com/multiversx/mx-chain-go/common/errChan"
	"github.com/multiversx/mx-chain-go/common/holders"
	statisticsDisabled "github.com/multiversx/mx-chain-go/common/statistics/disabled"
	nodeConfig "github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/epochStart/notifier"
	"github.com/multiversx/mx-chain-go/state/hashesCollector"
	"github.com/multiversx/mx-chain-go/state/parsers"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/databaseremover/disabled"
	"github.com/multiversx/mx-chain-go/storage/factory"
	"github.com/multiversx/mx-chain-go/storage/pruning"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/multiversx/mx-chain-go/trie/keyBuilder"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/urfave/cli"
)

var log = logger.GetOrCreate("trie")

var (
	// Hasher represents the internal hasher used by the node
	Hasher = blake2b.NewBlake2b()
	// Marshaller represents the internal marshaller used by the node
	Marshaller = &marshal.GogoProtoMarshalizer{}
)

func main() {
	app := cli.NewApp()
	app.Name = "Trie stats CLI app"
	app.Usage = "This is the entry point for the tool that benchmarks trie accesses"
	app.Authors = []cli.Author{
		{
			Name:  "The MultiversX Team",
			Email: "contact@multiversx.com",
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

	log.Info("finished processing trie")
}

func startProcess(_ *cli.Context) error {
	log.Info("starting processing trie", "pid", os.Getpid())
	return benchmarkTrieAccess()
}

func benchmarkTrieAccess() error {
	db, err := createStorer("path-to-db", false)
	if err != nil {
		return err
	}

	defer func() {
		err = db.Close()
		if err != nil {
			log.Error(err.Error())
		}
	}()

	tr, err := createTrie(db)
	if err != nil {
		return err
	}

	rootHashHex := "fb686c15770e0945f2efe3644287129d22629321ec20d3b5ebdde3733541c74a"
	rootHash, err := hex.DecodeString(rootHashHex)
	if err != nil {
		return fmt.Errorf("failed to decode root hash %s: %w", rootHashHex, err)
	}
	newTr, err := tr.Recreate(holders.NewDefaultRootHashesHolder(rootHash), "main")
	if err != nil {
		return fmt.Errorf("failed to recreate trie: %w", err)
	}

	// TODO: to avoid iterating the whole trie each time, save the leaves in a file and read them from there
	initialTrieKeys := make([][]byte, 0)
	leavesChannel := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity),
		ErrChan:    errChan.NewErrChanWrapper(),
	}
	err = newTr.GetAllLeavesOnChannel(leavesChannel, context.Background(), rootHash, keyBuilder.NewKeyBuilder(), parsers.NewMainTrieLeafParser())
	if err != nil {
		return fmt.Errorf("failed to get all leaves on channel: %w", err)
	}

	for leaf := range leavesChannel.LeavesChan {
		initialTrieKeys = append(initialTrieKeys, leaf.Key())
	}

	err = leavesChannel.ErrChan.ReadFromChanNonBlocking()
	if err != nil {
		return fmt.Errorf("error reading from leaves channel: %w", err)
	}

	numInserts := 5000
	numUpdates := 4500
	numDeletes := 500
	numGets := 10000
	numBlocks := 100

	avgInsertTime := time.Duration(0)
	avgUpdateTime := time.Duration(0)
	avgDeleteTime := time.Duration(0)
	avgGetTime := time.Duration(0)
	avgCommitTime := time.Duration(0)

	newData := generateDataForTrie(0, numInserts*numBlocks)
	wg := sync.WaitGroup{}
	wgGet := sync.WaitGroup{}
	for i := 0; i < numBlocks; i++ {
		wg.Add(3)
		wgGet.Add(1)
		startTime := time.Now()
		go func(blockIndex int) {
			insertStart := time.Now()
			insertNewData(numInserts, blockIndex*numInserts, newTr, newData)
			elapsedTimeInsert := time.Since(insertStart)
			avgInsertTime += elapsedTimeInsert
			log.Info("insert done", "blockIndex", blockIndex, "elapsedTimeInsert", elapsedTimeInsert)
			wg.Done()
		}(i)
		go func(blockIndex int) {
			updateStart := time.Now()
			updateExistingData(numUpdates, blockIndex*numUpdates, newTr, initialTrieKeys)
			elapsedTimeUpdate := time.Since(updateStart)
			avgUpdateTime += elapsedTimeUpdate
			log.Info("update done", "blockIndex", blockIndex, "elapsedTimeUpdate", elapsedTimeUpdate)
			wg.Done()
		}(i)
		go func(blockIndex int) {
			startTimeDelete := time.Now()
			deleteData(numDeletes, blockIndex*numDeletes, newTr, initialTrieKeys)
			elapsedTimeDelete := time.Since(startTimeDelete)
			avgDeleteTime += elapsedTimeDelete
			log.Info("delete done", "blockIndex", blockIndex, "elapsedTimeDelete", elapsedTimeDelete)
			wg.Done()
		}(i)

		go func(blockIndex int) {
			startTimeGet := time.Now()
			getData(numGets, newTr, initialTrieKeys)
			elapsedTimeGet := time.Since(startTimeGet)
			avgGetTime += elapsedTimeGet
			log.Info("get done", "blockIndex", blockIndex, "elapsedTimeGet", elapsedTimeGet)
			wgGet.Done()
		}(i)

		wg.Wait()
		elapsedTime := time.Since(startTime)
		log.Info("block processed", "blockIndex", i, "elapsedTime", elapsedTime)

		startTimeCommit := time.Now()
		err = newTr.Commit(hashesCollector.NewDisabledHashesCollector())
		elapsedTimeCommit := time.Since(startTimeCommit)
		avgCommitTime += elapsedTimeCommit
		log.Info("trie commit done", "blockIndex", i, "elapsedTimeCommit", elapsedTimeCommit)
		if err != nil {
			return err
		}

		wgGet.Wait()
	}

	log.Info("Average Insert Time", "avgInsertTime", avgInsertTime/time.Duration(numBlocks))
	log.Info("Average Update Time", "avgUpdateTime", avgUpdateTime/time.Duration(numBlocks))
	log.Info("Average Delete Time", "avgDeleteTime", avgDeleteTime/time.Duration(numBlocks))
	log.Info("Average Get Time", "avgGetTime", avgGetTime/time.Duration(numBlocks))
	log.Info("Average Commit Time", "avgCommitTime", avgCommitTime/time.Duration(numBlocks))

	return nil
}

func generateDataForTrie(startIndex int, numDataEntries int) [][]byte {
	data := make([][]byte, numDataEntries)

	for i := startIndex; i < startIndex+numDataEntries; i++ {
		data[i-startIndex] = Hasher.Compute(strconv.Itoa(i))
	}

	return data
}

func insertNewData(numInserts int, startIndex int, tr common.Trie, data [][]byte) {
	for i := startIndex; i < startIndex+numInserts; i++ {
		tr.Update(data[i], data[i])
	}
}

func updateExistingData(numUpdates int, startIndex int, tr common.Trie, data [][]byte) {
	maxIndex := len(data)
	for i := startIndex; i < startIndex+numUpdates; i++ {
		index := i % maxIndex
		tr.Update(data[index], append(data[index], []byte("_updated")...))
	}
}

func deleteData(numDeletes int, startIndex int, tr common.Trie, data [][]byte) {
	maxIndex := len(data)
	for i := startIndex; i < startIndex+numDeletes; i++ {
		index := i % maxIndex
		tr.Delete(data[index])
	}
}

func getData(numGets int, tr common.Trie, data [][]byte) {
	maxIndex := len(data)
	for i := 0; i < numGets; i++ {
		index := i % maxIndex
		val := data[index]
		_, _, _ = tr.Get(val)
	}
}

func createTrie(db storage.Storer) (common.Trie, error) {
	tsmArgs := trie.NewTrieStorageManagerArgs{
		MainStorer:  db,
		Marshalizer: Marshaller,
		Hasher:      Hasher,
		GeneralConfig: nodeConfig.TrieStorageManagerConfig{
			SnapshotsBufferLen:    10,
			SnapshotsGoroutineNum: 100,
		},
		IdleProvider:   commonDisabled.NewProcessStatusHandler(),
		Identifier:     dataRetriever.UserAccountsUnit.String(),
		StatsCollector: statisticsDisabled.NewStateStatistics(),
	}

	options := trie.StorageManagerOptions{
		PruningEnabled:   false,
		SnapshotsEnabled: false,
	}
	tsm, err := trie.CreateTrieStorageManager(tsmArgs, options)
	if err != nil {
		return nil, err
	}

	enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledCalled: func(_ core.EnableEpochFlag) bool {
			return true
		},
	}

	th, err := throttler.NewNumGoRoutinesThrottler(8)
	if err != nil {
		return nil, err
	}

	trieArgs := trie.TrieArgs{
		TrieStorage:          tsm,
		Marshalizer:          Marshaller,
		Hasher:               Hasher,
		EnableEpochsHandler:  enableEpochsHandler,
		MaxTrieLevelInMemory: 5,
		Throttler:            th,
		Identifier:           "mainTrie",
	}
	tr, err := trie.NewTrie(trieArgs)
	if err != nil {
		return nil, err
	}

	return tr, nil
}

func createStorer(path string, useTmpAsFilePath bool) (storage.Storer, error) {
	dbConfig := nodeConfig.DBConfig{
		FilePath:            path,
		Type:                "LvlDBSerial",
		BatchDelaySeconds:   2,
		MaxBatchSize:        45000,
		MaxOpenFiles:        10,
		UseTmpAsFilePath:    useTmpAsFilePath,
		ShardIDProviderType: "BinarySplit",
		NumShards:           4,
	}
	cacheConfig := storageunit.CacheConfig{
		Type:        "SizeLRU",
		Capacity:    500000,
		SizeInBytes: 31457280,
	}
	persisterFactory, err := factory.NewPersisterFactory(dbConfig)
	if err != nil {
		return nil, err
	}
	epochsData := pruning.EpochArgs{
		NumOfEpochsToKeep:     2,
		NumOfActivePersisters: 2,
		StartingEpoch:         1,
	}

	pathManager := &testscommon.PathManagerStub{
		PathForEpochCalled: func(shardId string, epoch uint32, identifier string) string {
			return filepath.Join(dbConfig.FilePath, fmt.Sprintf("%d", epoch))
		},
		PathForStaticCalled: func(shardId string, identifier string) string {
			return filepath.Join(dbConfig.FilePath, "Static")
		},
		DatabasePathCalled: func() string {
			return dbConfig.FilePath
		},
	}

	args := pruning.StorerArgs{
		Identifier:                "",
		ShardCoordinator:          testscommon.NewMultiShardsCoordinatorMock(1),
		CacheConf:                 cacheConfig,
		PathManager:               pathManager,
		DbPath:                    "",
		PersisterFactory:          persisterFactory,
		Notifier:                  notifier.NewManualEpochStartNotifier(),
		OldDataCleanerProvider:    &testscommon.OldDataCleanerProviderStub{},
		CustomDatabaseRemover:     disabled.NewDisabledCustomDatabaseRemover(),
		MaxBatchSize:              45000,
		EpochsData:                epochsData,
		PruningEnabled:            true,
		EnabledDbLookupExtensions: false,
		PersistersTracker:         pruning.NewPersistersTracker(epochsData),
		StateStatsHandler:         statisticsDisabled.NewStateStatistics(),
	}

	return pruning.NewTriePruningStorer(args)
}
