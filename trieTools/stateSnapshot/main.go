package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/statistics/disabled"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/factory"
	"github.com/multiversx/mx-chain-go/state/iteratorChannelsProvider"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	state2 "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/multiversx/mx-chain-go/update/mock"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-tools-go/trieTools/trieToolsCommon"
	"github.com/urfave/cli"
)

const (
	rootHashLength = 32
	addressLength  = 32
)

var log = logger.GetOrCreate("trieSnapshot")

func main() {
	app := cli.NewApp()
	app.Name = "Trie stats CLI app"
	app.Usage = "This is the entry point for the tool that does a state snapshot"
	app.Flags = trieToolsCommon.GetFlags()
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

	log.Info("execution finished successfully")
}

func startProcess(c *cli.Context) error {
	flagsConfig := trieToolsCommon.GetFlagsConfig(c)

	_, errLogger := trieToolsCommon.AttachFileLogger(log, "state-snapshot", flagsConfig)
	if errLogger != nil {
		return errLogger
	}

	log.Info("sanity checks...")

	err := logger.SetLogLevel(flagsConfig.LogLevel)
	if err != nil {
		return err
	}

	rootHash, err := hex.DecodeString(flagsConfig.HexRootHash)
	if err != nil {
		return fmt.Errorf("%w when decoding the provided hex root hash", err)
	}
	if len(rootHash) != rootHashLength {
		return fmt.Errorf("wrong root hash length: expected %d, got %d", rootHashLength, len(rootHash))
	}

	log.Info("starting processing state", "pid", os.Getpid())

	return snapshotState(flagsConfig, rootHash)
}

func snapshotState(flagsConfig trieToolsCommon.ContextFlagsConfig, rootHash []byte) error {
	storer, maxDbValue, err := createStorer(flagsConfig, log)
	if err != nil {
		return err
	}

	tsm, err := CreateTrieStorageManager(storer)
	if err != nil {
		return err
	}
	defer tsm.Close()

	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	lsm := &LastSnapshotMarkerStub{
		RemoveMarkerCalled: func(trieStorageManager common.StorageManager, epoch uint32, rootHash []byte) {
			waitGroup.Done()
		},
	}

	snapshotManager, err := CreateSnapshotsManager(lsm)
	if err != nil {
		return err
	}

	snapshotManager.SnapshotState(rootHash, uint32(maxDbValue), tsm)
	waitGroup.Wait()

	log.Info("state snapshot process finished successfully")
	time.Sleep(time.Second * 20)
	return nil
}

func CreateSnapshotsManager(lastSnapshotMarker state.LastSnapshotMarker) (state.SnapshotsManager, error) {
	addressConverter, err := pubkeyConverter.NewBech32PubkeyConverter(addressLength, trieToolsCommon.WalletHRP)
	if err != nil {
		return nil, err
	}

	argsAccCreator := factory.ArgsAccountCreator{
		Hasher:              trieToolsCommon.Hasher,
		Marshaller:          trieToolsCommon.Marshaller,
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	}
	accountFactory, err := factory.NewAccountCreator(argsAccCreator)
	if err != nil {
		return nil, err
	}

	args := state.ArgsNewSnapshotsManager{
		ShouldSerializeSnapshots: false,
		ProcessingMode:           common.Normal,
		Marshaller:               trieToolsCommon.Marshaller,
		AddressConverter:         addressConverter,
		ProcessStatusHandler:     &testscommon.ProcessStatusHandlerStub{},
		StateMetrics:             &state2.StateMetricsStub{},
		AccountFactory:           accountFactory,
		ChannelsProvider:         iteratorChannelsProvider.NewUserStateIteratorChannelsProvider(),
		StateStatsHandler:        disabled.NewStateStatistics(),
		LastSnapshotMarker:       lastSnapshotMarker,
	}

	snapshotManager, err := state.NewSnapshotsManager(args)
	if err != nil {
		return nil, err
	}
	err = snapshotManager.SetSyncer(&mock.AccountsDBSyncerStub{})
	if err != nil {
		return nil, err
	}

	return snapshotManager, nil
}

func CreateTrieStorageManager(
	persister storage.Storer,
) (common.StorageManager, error) {
	args := trie.NewTrieStorageManagerArgs{
		MainStorer:  persister,
		Marshalizer: trieToolsCommon.Marshaller,
		Hasher:      trieToolsCommon.Hasher,
		GeneralConfig: config.TrieStorageManagerConfig{
			PruningBufferLen:      1000,
			SnapshotsBufferLen:    1000000,
			SnapshotsGoroutineNum: 200,
		},
		IdleProvider:   &testscommon.ProcessStatusHandlerStub{},
		Identifier:     dataRetriever.UserAccountsUnit.String(),
		StatsCollector: disabled.NewStateStatistics(),
	}

	return trie.NewTrieStorageManager(args)
}

func createStorer(flags trieToolsCommon.ContextFlagsConfig, log logger.Logger) (storage.Storer, int, error) {
	maxDBValue, err := trieToolsCommon.GetMaxDBValue(filepath.Join(flags.WorkingDir, flags.DbDir), log)
	if err == nil {
		db, err := trieToolsCommon.CreatePruningStorer(flags, maxDBValue)
		return db, maxDBValue, err
	}

	log.Info("no ordered DBs for a pruning storer operation, will switch to single directory operation...")

	db, err := trieToolsCommon.CreateStorer(flags)
	return db, maxDBValue, err
}
