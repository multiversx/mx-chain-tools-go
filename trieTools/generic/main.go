package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/trie/keyBuilder"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/urfave/cli"

	"github.com/multiversx/mx-chain-tools-go/trieTools/generic/filter"
	"github.com/multiversx/mx-chain-tools-go/trieTools/trieToolsCommon"
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
var output []*filter.AccountDetails

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
	flagsConfig, err := getFlagsConfig(c)
	if err != nil {
		return err
	}

	output = make([]*filter.AccountDetails, 0)

	_, errLogger := trieToolsCommon.AttachFileLogger(log, logFilePrefix, flagsConfig.ContextFlagsConfig)
	if errLogger != nil {
		return errLogger
	}

	err = logger.SetLogLevel(flagsConfig.LogLevel)
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

	for keyValue := range iteratorChannels.LeavesChan {
		userAccount, found := getUserAccount(keyValue)
		if !found {
			continue
		}

		address := addressConverter.Encode(userAccount.Address)

		var pairs map[string]string
		var tokens map[string]struct{}
		pairs, tokens, err = retrievePairsAndTokensPerAccount(accDb, address)
		if err != nil {
			return fmt.Errorf("failed to retrieve key-value store for address [%s]: %v", address, err)
		}

		accDetails := &filter.AccountDetails{
			Address: address,
			Balance: userAccount.Balance,
			Nonce:   userAccount.Nonce,
			Pairs:   pairs,
			Tokens:  tokens,
		}

		ok := true
		if flagsConfig.Filters != nil {
			for _, f := range flagsConfig.Filters {
				if !f.ApplyFilter(accDetails) {
					ok = false
					break
				}
			}
		}

		if ok {
			output = append(output, accDetails)
		}

		numAccountsOnMainTrie++
		if int(flagsConfig.Limit) == numAccountsOnMainTrie {
			break
		}
	}

	err = common.GetErrorFromChanNonBlocking(iteratorChannels.ErrChan)
	if err != nil {
		return err
	}

	err = saveResult(output, flagsConfig.Outfile)
	if err != nil {
		return err
	}

	return nil
}

// TODO: these are duplicates, move to a common package to import from
func getUserAccount(kv core.KeyValueHolder) (*state.UserAccountData, bool) {
	userAccount := &state.UserAccountData{}
	errUnmarshal := trieToolsCommon.Marshaller.Unmarshal(userAccount, kv.Value())
	if errUnmarshal != nil {
		// probably a code node
		return nil, false
	}
	if len(userAccount.RootHash) == 0 {
		return nil, false
	}

	return userAccount, true
}

func saveResult(result []*filter.AccountDetails, outfile string) error {
	jsonBytes, err := json.MarshalIndent(result, "", " ")
	if err != nil {
		return err
	}

	log.Info("writing result in", "file", outfile)
	err = os.WriteFile(outfile, jsonBytes, fs.FileMode(outputFilePerms))
	if err != nil {
		return err
	}

	log.Info("finished exporting address-tokens map")
	return nil
}

func retrievePairsAndTokensPerAccount(accDb state.AccountsAdapter, address string) (map[string]string, map[string]struct{}, error) {
	addressBytes, err := addressConverter.Decode(address)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode address [%s]: %v", address, err)
	}

	account, err := accDb.GetExistingAccount(addressBytes)
	if err != nil {
		return nil, nil, err
	}

	userAccount, ok := account.(state.UserAccountHandler)
	if !ok {
		return nil, nil, fmt.Errorf("cannot cast AccountHandler to UserAccountHandler")
	}

	if check.IfNil(userAccount.DataTrie()) {
		return nil, nil, fmt.Errorf("the provided address doesn't have a data trie")
	}

	rootHash, err := userAccount.DataTrie().RootHash()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve root hash: %v", err)
	}

	iteratorChannels := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity),
		ErrChan:    make(chan error, 1),
	}
	err = userAccount.DataTrie().GetAllLeavesOnChannel(iteratorChannels, context.Background(), rootHash, keyBuilder.NewKeyBuilder())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve leaves on channels: %v", err)
	}

	allESDTs := make(map[string]struct{})
	keyValueMap := make(map[string]string)
	esdtPrefix := []byte(core.ProtectedKeyPrefix + core.ESDTKeyIdentifier)
	for leaf := range iteratorChannels.LeavesChan {
		getKeyValuePairs(leaf, userAccount, keyValueMap)
		getESDTTokens(leaf, esdtPrefix, allESDTs)
	}

	err = common.GetErrorFromChanNonBlocking(iteratorChannels.ErrChan)
	if err != nil {
		return nil, nil, err
	}

	return keyValueMap, allESDTs, nil
}

func getKeyValuePairs(leaf core.KeyValueHolder, userAccount state.UserAccountHandler, keyValueMap map[string]string) {
	suffix := append(leaf.Key(), userAccount.AddressBytes()...)
	value, errVal := leaf.ValueWithoutSuffix(suffix)
	if errVal != nil {
		log.Warn("cannot get value without suffix", "error", errVal, "key", leaf.Key())
		return
	}

	keyValueMap[hex.EncodeToString(leaf.Key())] = hex.EncodeToString(value)
}

func getESDTTokens(leaf core.KeyValueHolder, esdtPrefix []byte, allESDTs map[string]struct{}) {
	if !bytes.HasPrefix(leaf.Key(), esdtPrefix) {
		return
	}

	// TODO: Try to unmarshal it when the new meta data storage model will be live
	tokenKey := leaf.Key()
	lenESDTPrefix := len(esdtPrefix)
	tokenName := getPrettyTokenName(tokenKey[lenESDTPrefix:])

	allESDTs[tokenName] = struct{}{}
}

func getPrettyTokenName(tokenName []byte) string {
	token, nonce := common.ExtractTokenIDAndNonceFromTokenStorageKey(tokenName)
	if nonce != 0 {
		tokens := bytes.Split(token, []byte("-"))

		token = append(tokens[0], []byte("-")...)          // ticker-
		token = append(token, tokens[1]...)                // ticker-randSequence
		token = append(token, []byte("-")...)              // ticker-randSequence-
		token = append(token, getPrettyHexNonce(nonce)...) // ticker-randSequence-nonce
	}

	return string(token)
}

func getPrettyHexNonce(nonce uint64) []byte {
	nonceStr := fmt.Sprintf("%x", nonce)
	if len(nonceStr)%2 != 0 {
		nonceStr = "0" + nonceStr
	}

	return []byte(nonceStr)
}
