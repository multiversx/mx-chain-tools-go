package trieToolsCommon

import (
	"fmt"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	elrondFactory "github.com/ElrondNetwork/elrond-go/cmd/node/factory"
	"github.com/ElrondNetwork/elrond-go/common"
	commonDisabled "github.com/ElrondNetwork/elrond-go/common/disabled"
	"github.com/ElrondNetwork/elrond-go/common/logging"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	"github.com/ElrondNetwork/elrond-go/state"
	stateFactory "github.com/ElrondNetwork/elrond-go/state/factory"
	disabled2 "github.com/ElrondNetwork/elrond-go/state/storagePruningManager/disabled"
	"github.com/ElrondNetwork/elrond-go/storage/databaseremover/disabled"
	"github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/pruning"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/trie"
	"github.com/ElrondNetwork/elrond-tools-go/trieTools/trieToolsCommon/components"
	"io/ioutil"
	"math/big"
	"os"
	"path"
)

const (
	defaultLogsPath      = "logs"
	maxTrieLevelInMemory = 5
	maxDirs              = 100
)

// AttachFileLogger will attach the file logger, using provided flags
func AttachFileLogger(log logger.Logger, logFilePrefix string, flagsConfig ContextFlagsConfig) (elrondFactory.FileLoggingHandler, error) {
	var fileLogging elrondFactory.FileLoggingHandler
	var err error
	if flagsConfig.SaveLogFile {
		fileLogging, err = logging.NewFileLogging(logging.ArgsFileLogging{
			WorkingDir:      flagsConfig.WorkingDir,
			DefaultLogsPath: defaultLogsPath,
			LogFilePrefix:   logFilePrefix,
		})
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

// GetMaxDBValue will search in parentDir for all dbs directories and return max db value
func GetMaxDBValue(parentDir string, log logger.Logger) (int, error) {
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

// GetTrie will create and return a trie using the provided flags
func GetTrie(flags ContextFlagsConfig, maxDBValue int) (common.Trie, error) {
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
		CustomDatabaseRemover:     disabled.NewDisabledCustomDatabaseRemover(),
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

	return trie.NewTrie(tsm, Marshaller, Hasher, maxTrieLevelInMemory)
}

// NewAccountsAdapter will create a new accounts adapter using provided trie
func NewAccountsAdapter(trie common.Trie) (state.AccountsAdapter, error) {
	accCreator := stateFactory.NewAccountCreator()
	storagePruningManager := disabled2.NewDisabledStoragePruningManager()
	accountsAdapter, err := state.NewAccountsDB(state.ArgsAccountsDB{
		Trie:                  trie,
		Hasher:                Hasher,
		Marshaller:            Marshaller,
		AccountFactory:        accCreator,
		StoragePruningManager: storagePruningManager,
		ProcessingMode:        common.Normal,
		ProcessStatusHandler:  commonDisabled.NewProcessStatusHandler(),
	})

	return accountsAdapter, err
}
