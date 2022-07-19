package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"math/big"
	"os"
	"path"
	"path/filepath"

	"github.com/ElrondNetwork/elrond-go-core/core/pubkeyConverter"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	elrondFactory "github.com/ElrondNetwork/elrond-go/cmd/node/factory"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/common/logging"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	"github.com/ElrondNetwork/elrond-go/state"
	stateFactory "github.com/ElrondNetwork/elrond-go/state/factory"
	"github.com/ElrondNetwork/elrond-go/state/storagePruningManager/disabled"
	"github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/pruning"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/trie"
	"github.com/ElrondNetwork/elrond-tools-go/accountStorageExporter/components"
	"github.com/ElrondNetwork/elrond-tools-go/accountStorageExporter/config"
	"github.com/urfave/cli"
)

const (
	defaultLogsPath      = "logs"
	logFilePrefix        = "account-storage-exporter"
	maxTrieLevelInMemory = 5
	rootHashLength       = 32
	addressLength        = 32
	maxDirs              = 100
)

func main() {
	app := cli.NewApp()
	app.Name = "Accounts Storage Exporter CLI app"
	app.Usage = "This is the entry point for the tool that exports the storage of a given account"
	app.Flags = getFlags()
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

	log.Info("finished exporting the storage")
}

func startProcess(c *cli.Context) error {
	flagsConfig := getFlagsConfig(c)

	_, errLogger := attachFileLogger(log, flagsConfig)
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

	maxDBValue, err := getMaxDBValue(filepath.Join(flagsConfig.WorkingDir, flagsConfig.DbDir))
	if err != nil {
		return err
	}

	log.Info("starting exporting storage", "pid", os.Getpid())

	return exportStorage(flagsConfig.Address, flagsConfig, rootHash, maxDBValue)
}

func attachFileLogger(log logger.Logger, flagsConfig config.ContextFlagsConfig) (elrondFactory.FileLoggingHandler, error) {
	var fileLogging elrondFactory.FileLoggingHandler
	var err error
	if flagsConfig.SaveLogFile {
		fileLogging, err = logging.NewFileLogging(flagsConfig.WorkingDir, defaultLogsPath, logFilePrefix)
		if err != nil {
			return nil, fmt.Errorf("%w creating a log file", err)
		}
	}

	err = logger.SetDisplayByteSlice(logger.ToHex)
	log.LogIfError(err)
	logger.ToggleLoggerName(flagsConfig.EnableLogName)
	logLevelFlagValue := flagsConfig.LogLevel
	err = logger.SetLogLevel(logLevelFlagValue)
	if err != nil {
		return nil, err
	}

	if flagsConfig.DisableAnsiColor {
		err = logger.RemoveLogObserver(os.Stdout)
		if err != nil {
			return nil, err
		}

		err = logger.AddLogObserver(os.Stdout, &logger.PlainFormatter{})
		if err != nil {
			return nil, err
		}
	}
	log.Trace("logger updated", "level", logLevelFlagValue, "disable ANSI color", flagsConfig.DisableAnsiColor)

	return fileLogging, nil
}

func getMaxDBValue(parentDir string) (int, error) {
	contents, err := ioutil.ReadDir(parentDir)
	if err != nil {
		return 0, err
	}

	directories := make([]string, 0)
	for _, c := range contents {
		if !c.IsDir() {
			continue
		}

		_, ok := big.NewInt(0).SetString(c.Name(), 10)
		if !ok {
			log.Debug("DB directory found that will not be taken into account", "name", c.Name())
			continue
		}

		directories = append(directories, c.Name())
	}

	numDirs := 0
	for i := 0; i < maxDirs; i++ {
		expectedDir := fmt.Sprintf("%d", i)
		if !contains(directories, expectedDir) {
			break
		}

		numDirs++
	}

	if numDirs == 0 {
		return 0, fmt.Errorf("missing ordered directories in %s, like 0, 1 and so on", parentDir)
	}
	if numDirs != len(directories) {
		return 0, fmt.Errorf("unordered directories in %s, like 0, 1 and so on", parentDir)
	}

	return numDirs - 1, nil
}

func contains(haystack []string, needle string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}

	return false
}

func exportStorage(address string, flags config.ContextFlagsConfig, mainRootHash []byte, maxDBValue int) error {
	addressConverter, err := pubkeyConverter.NewBech32PubkeyConverter(addressLength, log)
	if err != nil {
		return err
	}

	tr, err := getTrie(flags, maxDBValue)
	if err != nil {
		return err
	}

	defer func() {
		errNotCritical := tr.Close()
		log.LogIfError(errNotCritical)
	}()

	accDb, err := newAccountsAdapter(tr)
	if err != nil {
		return err
	}

	err = accDb.RecreateTrie(mainRootHash)
	if err != nil {
		return err
	}

	addressBytes, err := addressConverter.Decode(address)
	if err != nil {
		return err
	}

	account, err := accDb.GetExistingAccount(addressBytes)
	if err != nil {
		return err
	}

	userAccount, ok := account.(state.UserAccountHandler)
	if !ok {
		return fmt.Errorf("cannot cast AccountHandler to UserAccountHandler")
	}

	leavesCh, err := userAccount.DataTrie().GetAllLeavesOnChannel(mainRootHash)
	if err != nil {
		return err
	}

	keyValueMap := make(map[string]string)
	for leaf := range leavesCh {
		keyValueMap[hex.EncodeToString(leaf.Key())] = hex.EncodeToString(leaf.Value())
	}

	fmt.Println(keyValueMap)

	jsonBytes, err := json.MarshalIndent(keyValueMap, "", " ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(outputFileName, jsonBytes, fs.FileMode(outputFilePerms))
	if err != nil {
		return err
	}

	return nil
}

func getTrie(flags config.ContextFlagsConfig, maxDBValue int) (common.Trie, error) {
	localDbConfig := dbConfig // copy
	localDbConfig.FilePath = path.Join(flags.WorkingDir, flags.DbDir)

	dbPath := path.Join(flags.WorkingDir, flags.DbDir)
	args := &pruning.StorerArgs{
		Identifier:                "",
		ShardCoordinator:          testscommon.NewMultiShardsCoordinatorMock(1),
		CacheConf:                 cacheConfig,
		PathManager:               components.NewSimplePathManager(dbPath),
		DbPath:                    "",
		PersisterFactory:          factory.NewPersisterFactory(localDbConfig),
		Notifier:                  notifier.NewManualEpochStartNotifier(),
		OldDataCleanerProvider:    &testscommon.OldDataCleanerProviderStub{},
		MaxBatchSize:              45000,
		NumOfEpochsToKeep:         uint32(maxDBValue) + 1,
		NumOfActivePersisters:     uint32(maxDBValue) + 1,
		StartingEpoch:             uint32(maxDBValue),
		PruningEnabled:            true,
		EnabledDbLookupExtensions: false,
	}

	db, err := pruning.NewTriePruningStorer(args)
	if err != nil {
		return nil, err
	}

	tsm, err := trie.NewTrieStorageManagerWithoutPruning(db)
	if err != nil {
		return nil, err
	}

	return trie.NewTrie(tsm, marshaller, hasher, maxTrieLevelInMemory)
}

func newAccountsAdapter(trie common.Trie) (state.AccountsAdapter, error) {
	accCreator := stateFactory.NewAccountCreator()
	storagePruningManager := disabled.NewDisabledStoragePruningManager()
	accountsAdapter, err := state.NewAccountsDB(
		trie,
		hasher,
		marshaller,
		accCreator,
		storagePruningManager,
		common.Normal,
	)

	return accountsAdapter, err
}
