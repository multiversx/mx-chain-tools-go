package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/trie/keyBuilder"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-tools-go/trieTools/trieToolsCommon"
	"github.com/urfave/cli"
)

const (
	logFilePrefix                    = "generic"
	rootHashLength                   = 32
	addressLength                    = 32
	outputFilePerms                  = 0644
	trieLeavesChannelDefaultCapacity = 100
)

var log = logger.GetOrCreate("main")
var marshaller = &marshal.GogoProtoMarshalizer{}
var addressConverter, _ = pubkeyConverter.NewBech32PubkeyConverter(32, log)

func main() {
	app := cli.NewApp()
	app.Name = "Sample application"
	app.Usage = "..."
	app.Flags = getFlags()
	app.Authors = []cli.Author{}

	app.Action = func(c *cli.Context) error {
		return doMain(c)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
		return
	}
}

func doMain(c *cli.Context) error {
	flagsConfig := getFlagsConfig(c)

	_, errLogger := trieToolsCommon.AttachFileLogger(log, logFilePrefix, flagsConfig.ContextFlagsConfig)
	if errLogger != nil {
		return errLogger
	}

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

	db, err := trieToolsCommon.CreatePruningStorer(flagsConfig.ContextFlagsConfig, maxDBValue)
	if err != nil {
		return err
	}

	tr, err := trieToolsCommon.CreateTrie(db)
	if err != nil {
		return err
	}

	defer func() {
		errNotCritical := tr.Close()
		log.LogIfError(errNotCritical)
	}()

	iteratorChannels := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder, trieLeavesChannelDefaultCapacity),
		ErrChan:    make(chan error, 1),
	}

	log.Info("Roothash", "roothash", rootHash)

	err = tr.GetAllLeavesOnChannel(iteratorChannels, context.Background(), rootHash, keyBuilder.NewKeyBuilder())
	if err != nil {
		return err
	}

	accDb, err := trieToolsCommon.NewAccountsAdapter(tr)
	if err != nil {
		return err
	}

	err = accDb.RecreateTrie(rootHash)
	if err != nil {
		return err
	}

	numAccountsOnMainTrie := 0

	for range iteratorChannels.LeavesChan {
		numAccountsOnMainTrie++

		if numAccountsOnMainTrie%1000 == 0 {
			fmt.Println(numAccountsOnMainTrie)
		}
	}

	err = common.GetErrorFromChanNonBlocking(iteratorChannels.ErrChan)
	if err != nil {
		return err
	}

	return nil
}
