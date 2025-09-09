package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/marshal"
	marshalFactory "github.com/multiversx/mx-chain-core-go/marshal/factory"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/state/accounts"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-tools-go/trieTools/trieToolsCommon"
	"github.com/urfave/cli"
)

var log = logger.GetOrCreate("trie")

const (
	logFilePrefix    = "trie"
	accountsFileName = "accountsAddresses"
	rootHashLength   = 32
)

func main() {
	app := cli.NewApp()
	app.Name = "Trie stats CLI app"
	app.Usage = "This is the entry point for the tool that checks collected state changes"
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

	_, errLogger := trieToolsCommon.AttachFileLogger(log, logFilePrefix, flagsConfig)
	if errLogger != nil {
		return errLogger
	}

	err := logger.SetLogLevel(flagsConfig.LogLevel)
	if err != nil {
		return err
	}

	log.Info("starting processing trie", "pid", os.Getpid())

	rootHash, err := hex.DecodeString(flagsConfig.HexRootHash)
	if err != nil {
		return fmt.Errorf("%w when decoding the provided hex root hash", err)
	}
	if len(rootHash) != rootHashLength {
		return fmt.Errorf("wrong root hash length: expected %d, got %d", rootHashLength, len(rootHash))
	}

	return printAccounts(rootHash, flagsConfig)
}

func getAccounts() ([][]byte, error) {
	file, err := os.Open(accountsFileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	accountsAddressesHex := make([][]byte, 0)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		accountHex := scanner.Text()
		accountBytes, err := hex.DecodeString(accountHex)
		if err != nil {
			return nil, err
		}
		accountsAddressesHex = append(accountsAddressesHex, accountBytes)
	}

	return accountsAddressesHex, nil
}

func printAccounts(mainRootHash []byte, flags trieToolsCommon.ContextFlagsConfig) error {
	gogoProtoMarshaller, err := marshalFactory.NewMarshalizer(marshalFactory.GogoProtobuf)
	if err != nil {
		return err
	}

	tr, err := getStateComponents(flags, mainRootHash)
	if err != nil {
		return err
	}

	acc, err := getAccounts()
	if err != nil {
		return err
	}
	for _, account := range acc {
		accData, _, err := tr.Get(account)
		if err != nil {
			return err
		}
		printAccountData("account data", account, accData, gogoProtoMarshaller)
	}

	return nil
}

func getStateComponents(flags trieToolsCommon.ContextFlagsConfig, mainRootHash []byte) (common.Trie, error) {
	trieStorer, err := createStorer(flags, log)
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

func printAccountData(message string, trieKey []byte, accountBytes []byte, marshaller marshal.Marshalizer) {
	account := &accounts.UserAccountData{}
	err := marshaller.Unmarshal(account, accountBytes)
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

	log.Debug("trie account data", "trie key", trieKey, "trieVal", accountBytes)
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
