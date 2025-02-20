package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/stateChange"
	"github.com/multiversx/mx-chain-core-go/marshal"
	marshalFactory "github.com/multiversx/mx-chain-core-go/marshal/factory"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/accounts"
	"github.com/multiversx/mx-chain-go/storage"
	storageFactory "github.com/multiversx/mx-chain-go/storage/factory"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-tools-go/trieTools/trieToolsCommon"
	"github.com/urfave/cli"
)

var log = logger.GetOrCreate("trie")

const (
	logFilePrefix        = "trie"
	fileHeader           = "----------------------------------------------------------"
	rootHashId           = "rootHash"
	scheduledRootHashId  = "scheduledRootHash"
	headerHashesFileName = "headerHashes"
)

func main() {
	app := cli.NewApp()
	app.Name = "Trie stats CLI app"
	app.Usage = "This is the entry point for the tool that checks collected state changes"
	app.Flags = getFlags()
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
	flagsConfig := getFlagsConfig(c)

	_, errLogger := trieToolsCommon.AttachFileLogger(log, logFilePrefix, flagsConfig.ContextFlagsConfig)
	if errLogger != nil {
		return errLogger
	}

	log.Info("sanity checks...")

	err := logger.SetLogLevel(flagsConfig.LogLevel)
	if err != nil {
		return err
	}

	log.Info("starting processing trie", "pid", os.Getpid())

	return applyStateChanges(flagsConfig)
}

func applyStateChanges(flags contextFlagsConfig) error {
	stateChangesDb, err := openStateChangesDB(flags.StateChangesDBPath)
	if err != nil {
		return err
	}
	gogoProtoMarshaller, err := marshalFactory.NewMarshalizer(marshalFactory.GogoProtobuf)
	if err != nil {
		return err
	}
	jsonMarshaller, err := marshalFactory.NewMarshalizer(marshalFactory.JsonMarshalizer)
	if err != nil {
		return err
	}
	headerHashes, err := getHeaderHashes()
	if err != nil {
		return err
	}
	if len(headerHashes) == 0 {
		return fmt.Errorf("no header hashes found")
	}
	mainRootHash, err := getStartingRootHash(headerHashes[0], stateChangesDb)
	if err != nil {
		return err
	}
	log.Debug("starting applying state changes from rootHash", "mainRootHash", mainRootHash)
	tr, err := getStateComponents(flags, mainRootHash)
	if err != nil {
		return err
	}

	for headerIndex := 0; headerIndex < len(headerHashes)-1; headerIndex++ {
		headerHash := headerHashes[headerIndex]
		log.Debug("started applying state changes for header hash", "headerHash", headerHash)

		txHashes, err := getOrderedTxHashes(headerHash, stateChangesDb, jsonMarshaller)
		if err != nil {
			return err
		}
		touchedAccounts := make(map[string]string)
		dataTries := make(map[string]common.Trie)

		for _, txHash := range txHashes {
			stateChanges, err := getStateChangesForTx(txHash, gogoProtoMarshaller, stateChangesDb)
			if err != nil {
				log.Error("error when getting state changes for tx", "txHash", txHash, "err", err.Error())
				return err
			}
			log.Debug("applying state changes for tx", "txHash", txHash, "numStateChanges", len(stateChanges.StateChanges))

			err = applyStateChange(tr, dataTries, stateChanges, touchedAccounts)
			if err != nil {
				log.Error("error when applying state changes for tx", "txHash", txHash, "err", err.Error())
				return err
			}
		}

		nextHeaderHash := headerHashes[headerIndex+1]
		err = checkStateChangesAppliedSuccessfully(tr, dataTries, touchedAccounts, gogoProtoMarshaller, headerHash, nextHeaderHash, stateChangesDb)
		if err != nil {
			return err
		}
	}

	log.Debug("state changes applied successfully")

	return nil
}

func getStartingRootHash(headerHash []byte, db storage.Persister) ([]byte, error) {
	rootHash, err := getScheduledRootHash(db, headerHash)
	if err != nil {
		return nil, err
	}
	if len(rootHash) != 0 {
		return rootHash, nil
	}

	expectedRootHashKey := append([]byte(rootHashId), headerHash...)
	return db.Get(expectedRootHashKey)
}

func checkStateChangesAppliedSuccessfully(
	tr common.Trie,
	dataTries map[string]common.Trie,
	touchedAccounts map[string]string,
	marshaller marshal.Marshalizer,
	headerHash []byte,
	nextHeaderHash []byte,
	db storage.Persister,
) error {
	err := tr.Commit()
	if err != nil {
		return err
	}

	rootHash, err := tr.RootHash()
	if err != nil {
		return err
	}

	expectedRootHash, err := getExpectedRootHash(headerHash, nextHeaderHash, db)
	if err != nil {
		return err
	}

	if !bytes.Equal(expectedRootHash, rootHash) {
		printTouchedAccounts(touchedAccounts, expectedRootHash, tr, marshaller)
		return fmt.Errorf("expected main trie root hash %x, got %x", expectedRootHash, rootHash)
	}

	for address, dataTrie := range dataTries {
		addressBytes, err := hex.DecodeString(address)
		if err != nil {
			return err
		}

		accountBytes, _, err := tr.Get(addressBytes)
		if err != nil {
			return err
		}
		if len(accountBytes) == 0 {
			return fmt.Errorf("account not found in original trie, address: %v", address)
		}

		userAcc := &accounts.UserAccountData{}
		err = marshaller.Unmarshal(userAcc, accountBytes)
		if err != nil {
			return err
		}

		err = dataTrie.Commit()
		if err != nil {
			return err
		}

		dataTrieRootHash, err := dataTrie.RootHash()
		if err != nil {
			return err
		}

		if !bytes.Equal(userAcc.RootHash, dataTrieRootHash) {
			recreateDataTrie, err := tr.Recreate(holders.NewDefaultRootHashesHolder(userAcc.RootHash))
			if err != nil {
				return err
			}
			log.Debug(recreateDataTrie.String())
			log.Debug(dataTrie.String())
			return fmt.Errorf("expected data trie root hash %v, got %v. Account address: %v", hex.EncodeToString(userAcc.RootHash), hex.EncodeToString(dataTrieRootHash), address)
		}
	}

	log.Debug("finished applying state changes for header hash", "headerHash", headerHash)
	return nil
}

func getExpectedRootHash(headerHash []byte, nextHeaderHash []byte, db storage.Persister) ([]byte, error) {
	if len(nextHeaderHash) != 0 {
		rootHash, err := getScheduledRootHash(db, nextHeaderHash)
		if err != nil {
			return nil, err
		}
		if len(rootHash) != 0 {
			return rootHash, nil
		}
	}

	expectedRootHashKey := append([]byte(rootHashId), headerHash...)
	return db.Get(expectedRootHashKey)
}

func getScheduledRootHash(db storage.Persister, headerHash []byte) ([]byte, error) {
	expectedRootHashKey := append([]byte(scheduledRootHashId), headerHash...)
	return db.Get(expectedRootHashKey)
}

func printTouchedAccounts(touchedAccounts map[string]string, expectedRootHash []byte, tr common.Trie, marshaller marshal.Marshalizer) {
	expectedRootHashHolder := holders.NewDefaultRootHashesHolder(expectedRootHash)
	expectedTrie, err := tr.Recreate(expectedRootHashHolder)
	if err != nil {
		log.Error("error when recreating trie", "err", err.Error())
		return
	}
	for accountKey, accountVal := range touchedAccounts {
		printAccountData("touched account", accountVal, marshaller)

		accountKeyBytes, err := hex.DecodeString(accountKey)
		if err != nil {
			log.Error("error when decoding account key", "account key", accountKey, "err", err.Error())
			continue
		}
		accountBytes, _, err := expectedTrie.Get(accountKeyBytes)
		if err != nil {
			log.Error("error when getting account from expected trie", "account key", accountKey, "err", err.Error())
			continue
		}
		if len(accountBytes) == 0 {
			log.Error("account not found in expected trie", "account key", accountKey)
			continue
		}
		printAccountData("expected account", hex.EncodeToString(accountBytes), marshaller)
	}
}

func getHeaderHashes() ([][]byte, error) {
	file, err := os.Open(headerHashesFileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	headerHashes := make([][]byte, 0)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		hexHeaderHash := scanner.Text()
		if hexHeaderHash == fileHeader {
			continue
		}
		headerHash, err := hex.DecodeString(hexHeaderHash)
		if err != nil {
			return nil, err
		}
		headerHashes = append(headerHashes, headerHash)
	}

	return headerHashes, nil
}

func getOrderedTxHashes(headerHash []byte, db storage.Persister, marshaller marshal.Marshalizer) ([][]byte, error) {
	marshalledTxsWithOrder, err := db.Get(headerHash)
	if err != nil {
		return nil, err
	}
	txsWithOrder := make(map[string]uint32)
	err = marshaller.Unmarshal(&txsWithOrder, marshalledTxsWithOrder)
	if err != nil {
		return nil, err
	}
	txHashes := make([][]byte, len(txsWithOrder))
	for txHash, index := range txsWithOrder {
		txHashBytes, err := hex.DecodeString(txHash)
		if err != nil {
			return nil, err
		}
		txHashes[index] = txHashBytes
	}

	return txHashes, nil
}

func getStateChangesForTx(txHash []byte, marshaller marshal.Marshalizer, db storage.Persister) (*stateChange.StateChanges, error) {
	stateChangesForTxs := &stateChange.StateChanges{}

	val, err := db.Get(txHash)
	if err != nil {
		return nil, err
	}

	err = marshaller.Unmarshal(stateChangesForTxs, val)
	if err != nil {
		return nil, err
	}

	return stateChangesForTxs, nil
}

func applyStateChange(
	mainTrie common.Trie,
	dataTries map[string]common.Trie,
	collectedChanges *stateChange.StateChanges,
	touchedAccounts map[string]string,
) error {
	for _, sc := range collectedChanges.StateChanges {
		if sc.Type == stateChange.Read {
			continue
		}

		switch sc.Operation {
		case stateChange.SaveAccount:
			err := applySaveAccountChanges(mainTrie, dataTries, sc, touchedAccounts)
			if err != nil {
				return err
			}
		case stateChange.WriteCode:
			touchedAccounts[hex.EncodeToString(sc.MainTrieKey)] = hex.EncodeToString(sc.MainTrieVal)
			err := mainTrie.Update(sc.MainTrieKey, sc.MainTrieVal)
			if err != nil {
				return err
			}
		case stateChange.RemoveDataTrie:
			return fmt.Errorf("remove data trie not implemented")
		default:
			return fmt.Errorf("unknown state change operation: %d", sc.Operation)
		}
	}

	return nil
}

func applySaveAccountChanges(
	mainTrie common.Trie,
	dataTries map[string]common.Trie,
	sc *stateChange.StateChange,
	touchedAccounts map[string]string,
) error {
	err := applyDataTrieChanges(mainTrie, sc, dataTries)
	if err != nil {
		return err
	}

	err = mainTrie.Update(sc.MainTrieKey, sc.MainTrieVal)
	if err != nil {
		return err
	}

	touchedAccounts[hex.EncodeToString(sc.MainTrieKey)] = hex.EncodeToString(sc.MainTrieVal)
	return nil
}

func applyDataTrieChanges(mainTrie common.Trie, sc *stateChange.StateChange, dataTries map[string]common.Trie) error {
	if len(sc.DataTrieChanges) == 0 {
		return nil
	}

	dataTrie, err := getDataTrie(sc.MainTrieKey, mainTrie, dataTries)
	if err != nil {
		return err
	}

	dt, ok := dataTrie.(state.DataTrie)
	if !ok {
		return fmt.Errorf("data trie is not a data trie")
	}

	for _, dataTrieChange := range sc.DataTrieChanges {
		if dataTrieChange.Type == stateChange.Read {
			continue
		}

		err = dt.UpdateWithVersion(dataTrieChange.Key, dataTrieChange.Val, core.TrieNodeVersion(dataTrieChange.Version))
		if err != nil {
			return err
		}
	}

	return nil
}

func getDataTrie(
	address []byte,
	mainTrie common.Trie,
	dataTries map[string]common.Trie,
) (common.Trie, error) {
	dataTrie, ok := dataTries[hex.EncodeToString(address)]
	if ok {
		return dataTrie, nil
	}

	accountBytes, _, err := mainTrie.Get(address)
	if err != nil {
		return nil, err
	}
	userAcc := &accounts.UserAccountData{}
	err = trieToolsCommon.Marshaller.Unmarshal(userAcc, accountBytes)
	if err != nil {
		return nil, err
	}
	dataTrieRootHash := holders.NewDefaultRootHashesHolder(userAcc.RootHash)
	dataTrie, err = mainTrie.Recreate(dataTrieRootHash)
	if err != nil {
		return nil, err
	}

	dataTries[hex.EncodeToString(address)] = dataTrie
	return dataTrie, nil
}

func getStateComponents(flags contextFlagsConfig, mainRootHash []byte) (common.Trie, error) {
	trieStorer, err := createStorer(flags.ContextFlagsConfig, log)
	if err != nil {
		return nil, err
	}

	enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
			return true
		},
	}

	tr, err := trieToolsCommon.CreateTrie(trieStorer, enableEpochsHandler)
	if err != nil {
		return nil, err
	}

	rootHashHolder := holders.NewDefaultRootHashesHolder(mainRootHash)
	newTr, err := tr.Recreate(rootHashHolder)
	if err != nil {
		return nil, err
	}

	return newTr, nil
}

func createStorer(flags trieToolsCommon.ContextFlagsConfig, log logger.Logger) (storage.Storer, error) {
	maxDBValue, err := trieToolsCommon.GetMaxDBValue(filepath.Join(flags.WorkingDir, flags.DbDir), log)
	if err == nil {
		return trieToolsCommon.CreatePruningStorer(flags, maxDBValue)
	}

	log.Info("no ordered DBs for a pruning storer operation, will switch to single directory operation...")

	return trieToolsCommon.CreateStorer(flags)
}

func openStateChangesDB(dbPath string) (storage.Persister, error) {
	dbConfig := config.DBConfig{
		FilePath:          dbPath,
		Type:              "LvlDBSerial",
		BatchDelaySeconds: 2,
		MaxBatchSize:      100,
		MaxOpenFiles:      10,
	}

	persisterFactory, err := storageFactory.NewPersisterFactory(dbConfig)
	if err != nil {
		return nil, err
	}

	db, err := persisterFactory.Create(dbConfig.FilePath)
	if err != nil {
		return nil, fmt.Errorf("%w while creating the db for the trie nodes", err)
	}

	return db, nil
}

func printAccountData(message string, accountData string, marshaller marshal.Marshalizer) {
	accountBytes, err := hex.DecodeString(accountData)
	if err != nil {
		log.Error("error when decoding account data", "err", err.Error())
		return
	}

	account := &accounts.UserAccountData{}
	err = marshaller.Unmarshal(account, accountBytes)
	if err != nil {
		log.Error("error when unmarshalling account data", "err", err.Error())
		return
	}

	balance := uint64(0)
	if account.Balance != nil {
		balance = account.Balance.Uint64()
	}
	devReward := uint64(0)
	if account.DeveloperReward != nil {
		devReward = account.DeveloperReward.Uint64()
	}

	log.Debug(message,
		"address", account.Address,
		"nonce", account.Nonce,
		"balance", balance,
		"codeHash", account.CodeHash,
		"rootHash", account.RootHash,
		"developerReward", devReward,
		"ownerAddress", account.OwnerAddress,
		"userName", account.UserName,
		"codeMetaData", account.CodeMetadata)
}
