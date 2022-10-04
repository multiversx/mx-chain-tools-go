package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/core/pubkeyConverter"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-tools-go/trieTools/accountStorageExporter/config"
	"github.com/ElrondNetwork/elrond-tools-go/trieTools/trieToolsCommon"
	"github.com/urfave/cli"
)

const (
	logFilePrefix  = "account-storage-exporter"
	rootHashLength = 32
	addressLength  = 32
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

	_, errLogger := trieToolsCommon.AttachFileLogger(log, logFilePrefix, flagsConfig.ContextFlagsConfig)
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

	maxDBValue, err := trieToolsCommon.GetMaxDBValue(filepath.Join(flagsConfig.WorkingDir, flagsConfig.DbDir), log)
	if err != nil {
		return err
	}

	log.Info("starting exporting storage", "pid", os.Getpid())

	return exportStorage(flagsConfig.Address, flagsConfig, rootHash, maxDBValue)
}

func exportStorage(address string, flags config.ContextFlagsConfigAddr, mainRootHash []byte, maxDBValue int) error {
	addressConverter, err := pubkeyConverter.NewBech32PubkeyConverter(addressLength, log)
	if err != nil {
		return err
	}

	db, err := trieToolsCommon.GetPruningStorer(flags.ContextFlagsConfig, maxDBValue)
	if err != nil {
		return err
	}

	tr, err := trieToolsCommon.GetTrie(db)
	if err != nil {
		return err
	}

	defer func() {
		errNotCritical := tr.Close()
		log.LogIfError(errNotCritical)
	}()

	accDb, err := trieToolsCommon.NewAccountsAdapter(tr)
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

	if check.IfNil(userAccount.DataTrie()) {
		return fmt.Errorf("the provided address doesn't have a data trie")
	}

	rootHash, err := userAccount.DataTrie().RootHash()
	if err != nil {
		return err
	}

	leavesCh := make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity)
	err = userAccount.DataTrie().GetAllLeavesOnChannel(leavesCh, context.Background(), rootHash)
	if err != nil {
		return err
	}

	keyValueMap := make(map[string]string)
	for leaf := range leavesCh {
		suffix := append(leaf.Key(), userAccount.AddressBytes()...)
		value, errVal := leaf.ValueWithoutSuffix(suffix)
		if errVal != nil {
			log.Warn("cannot get value without suffix", "error", errVal, "key", leaf.Key())
			continue
		}

		keyValueMap[hex.EncodeToString(leaf.Key())] = hex.EncodeToString(value)
	}

	jsonBytes, err := json.MarshalIndent(keyValueMap, "", " ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(outputFileName, jsonBytes, fs.FileMode(outputFilePerms))
	if err != nil {
		return err
	}

	log.Info("key-value map", "value", keyValueMap)

	return nil
}
