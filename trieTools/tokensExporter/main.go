package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/core/pubkeyConverter"
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
	"github.com/ElrondNetwork/elrond-tools-go/trieTools/trieToolsCommon"
	"github.com/ElrondNetwork/elrond-tools-go/trieTools/trieToolsCommon/components"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/urfave/cli"
	"io/fs"
	"io/ioutil"
	"math/big"
	"os"
	"path"
	"path/filepath"
)

const (
	defaultLogsPath      = "logs"
	logFilePrefix        = "accounts-tokens-exporter"
	maxTrieLevelInMemory = 5
	rootHashLength       = 32
	addressLength        = 32
	maxDirs              = 100
	outputFileName       = "output.json"
	outputFilePerms      = 0644
)

func main() {
	app := cli.NewApp()
	app.Name = "Tokens exporter CLI app"
	app.Usage = "This is the entry point for the tool that exports all tokens for a given root hash"
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

	log.Info("finished exporting address-tokens map")
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

	log.Info("starting processing trie", "pid", os.Getpid())

	return exportTokens(flagsConfig, rootHash, maxDBValue)
}

func attachFileLogger(log logger.Logger, flagsConfig trieToolsCommon.ContextFlagsConfig) (elrondFactory.FileLoggingHandler, error) {
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

func exportTokens(flags trieToolsCommon.ContextFlagsConfig, mainRootHash []byte, maxDBValue int) error {
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

	ch := make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity)
	err = tr.GetAllLeavesOnChannel(ch, context.Background(), mainRootHash)
	if err != nil {
		return err
	}

	accDb, err := newAccountsAdapter(tr)
	if err != nil {
		return err
	}

	err = accDb.RecreateTrie(mainRootHash)
	if err != nil {
		return err
	}

	numAccountsOnMainTrie := 0
	addressTokensMap := make(map[string]map[string]struct{})
	for keyValue := range ch {
		address, found := getAddress(keyValue)
		if !found {
			continue
		}

		numAccountsOnMainTrie++

		account, errGetAccount := accDb.GetExistingAccount(address)
		if errGetAccount != nil {
			return errGetAccount
		}

		esdtTokens, errGetESDT := getAllESDTTokens(account)
		if errGetESDT != nil {
			return errGetESDT
		}

		if len(esdtTokens) > 0 {
			encodedAddress := addressConverter.Encode(address)
			addressTokensMap[encodedAddress] = esdtTokens
		}
	}

	log.Info("parsed main trie",
		"num accounts", numAccountsOnMainTrie,
		"num accounts with tokens", len(addressTokensMap))

	return saveAndPrintResult(addressTokensMap)
}

func getAddress(kv core.KeyValueHolder) ([]byte, bool) {
	userAccount := &state.UserAccountData{}
	errUnmarshal := trieToolsCommon.Marshaller.Unmarshal(userAccount, kv.Value())
	if errUnmarshal != nil {
		// probably a code node
		return nil, false
	}
	if len(userAccount.RootHash) == 0 {
		return nil, false
	}

	return kv.Key(), true
}

func saveAndPrintResult(addressTokensMap map[string]map[string]struct{}) error {
	jsonBytes, err := json.MarshalIndent(addressTokensMap, "", " ")
	if err != nil {
		return err
	}

	log.Info("parsing result written in", "file", outputFileName)
	err = ioutil.WriteFile(outputFileName, jsonBytes, fs.FileMode(outputFilePerms))
	if err != nil {
		return err
	}

	for address, tokens := range addressTokensMap {
		for token := range tokens {
			log.Info("", "address", address, "token", token)
		}
	}

	return nil
}

func getTrie(flags trieToolsCommon.ContextFlagsConfig, maxDBValue int) (common.Trie, error) {
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

	return trie.NewTrie(tsm, trieToolsCommon.Marshaller, trieToolsCommon.Hasher, maxTrieLevelInMemory)
}

func newAccountsAdapter(trie common.Trie) (state.AccountsAdapter, error) {
	accCreator := stateFactory.NewAccountCreator()
	storagePruningManager := disabled2.NewDisabledStoragePruningManager()
	accountsAdapter, err := state.NewAccountsDB(state.ArgsAccountsDB{
		Trie:                  trie,
		Hasher:                trieToolsCommon.Hasher,
		Marshaller:            trieToolsCommon.Marshaller,
		AccountFactory:        accCreator,
		StoragePruningManager: storagePruningManager,
		ProcessingMode:        common.Normal,
		ProcessStatusHandler:  commonDisabled.NewProcessStatusHandler(),
	})

	return accountsAdapter, err
}

func getAllESDTTokens(account vmcommon.AccountHandler) (map[string]struct{}, error) {
	userAccount, ok := account.(state.UserAccountHandler)
	if !ok {
		return nil, fmt.Errorf("could not convert account to user account, address = %s",
			hex.EncodeToString(account.AddressBytes()))
	}

	allESDTs := make(map[string]struct{})
	if check.IfNil(userAccount.DataTrie()) {
		return allESDTs, nil
	}

	rootHash, err := userAccount.DataTrie().RootHash()
	if err != nil {
		return nil, err
	}

	chLeaves := make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity)
	err = userAccount.DataTrie().GetAllLeavesOnChannel(chLeaves, context.Background(), rootHash)
	if err != nil {
		return nil, err
	}

	esdtPrefix := []byte(core.ElrondProtectedKeyPrefix + core.ESDTKeyIdentifier)
	for leaf := range chLeaves {
		if !bytes.HasPrefix(leaf.Key(), esdtPrefix) {
			continue
		}

		tokenKey := leaf.Key()
		lenESDTPrefix := len(esdtPrefix)
		tokenName := getPrettyTokenName(tokenKey[lenESDTPrefix:])

		allESDTs[tokenName] = struct{}{}
	}

	return allESDTs, nil
}

func getPrettyTokenName(tokenName []byte) string {
	token, nonce := common.ExtractTokenIDAndNonceFromTokenStorageKey(tokenName)
	if nonce != 0 {
		tokens := bytes.Split(token, []byte("-"))

		token = append(tokens[0], []byte("-")...)                           // ticker-
		token = append(token, tokens[1]...)                                 // ticker-randSequence
		token = append(token, []byte("-")...)                               // ticker-randSequence-
		token = append(token, []byte(big.NewInt(int64(nonce)).String())...) // ticker-randSequence-nonce
	}

	return string(token)
}
