package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/core/pubkeyConverter"
	"github.com/ElrondNetwork/elrond-go-core/data/esdt"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-tools-go/trieTools/tokensExporter/config"
	"github.com/ElrondNetwork/elrond-tools-go/trieTools/trieToolsCommon"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/urfave/cli"
)

const (
	logFilePrefix   = "accounts-tokens-exporter"
	rootHashLength  = 32
	addressLength   = 32
	outputFilePerms = 0644
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

	log.Info("starting processing trie", "pid", os.Getpid())

	return exportTokens(flagsConfig, rootHash, maxDBValue)
}

func exportTokens(flags config.ContextFlagsTokensExporter, mainRootHash []byte, maxDBValue int) error {
	addressConverter, err := pubkeyConverter.NewBech32PubkeyConverter(addressLength, log)
	if err != nil {
		return err
	}

	tr, err := trieToolsCommon.GetTrie(flags.ContextFlagsConfig, maxDBValue)
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

	accDb, err := trieToolsCommon.NewAccountsAdapter(tr)
	if err != nil {
		return err
	}

	err = accDb.RecreateTrie(mainRootHash)
	if err != nil {
		return err
	}

	numAccountsOnMainTrie := 0
	allBalances := big.NewInt(0)
	numAccountsWithToken := 0
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

		if bytes.Compare(account.AddressBytes(), vmcommon.SystemAccountAddress) == 0 {
			log.Debug("found sys account address in trie, ignoring it...")
			continue
		}

		esdtBalance, errGetESDT := getESDTBalance(flags.Token, account, addressConverter)
		if errGetESDT != nil {
			return errGetESDT
		}

		if esdtBalance.Cmp(big.NewInt(0)) > 0 {
			allBalances = big.NewInt(0).Add(allBalances, esdtBalance)
			numAccountsWithToken++
		}
	}

	log.Info("found ", "global balance", allBalances.String(), "num accounts with token", numAccountsWithToken)
	return nil
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

func saveResult(addressTokensMap map[string]map[string]struct{}, outfile string) error {
	jsonBytes, err := json.MarshalIndent(addressTokensMap, "", " ")
	if err != nil {
		return err
	}

	log.Info("writing result in", "file", outfile)
	err = ioutil.WriteFile(outfile, jsonBytes, fs.FileMode(outputFilePerms))
	if err != nil {
		return err
	}

	log.Info("finished exporting address-tokens map")
	return nil
}

func getESDTBalance(token string, account vmcommon.AccountHandler, pubKeyConverter core.PubkeyConverter) (*big.Int, error) {
	userAccount, ok := account.(state.UserAccountHandler)
	if !ok {
		return nil, fmt.Errorf("could not convert account to user account, address = %s",
			pubKeyConverter.Encode(account.AddressBytes()))
	}

	balance := big.NewInt(0)
	if check.IfNil(userAccount.DataTrie()) {
		return balance, nil
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

	marshaller := marshal.GogoProtoMarshalizer{}
	esdtPrefix := []byte(core.ElrondProtectedKeyPrefix + core.ESDTKeyIdentifier + token)
	for leaf := range chLeaves {
		if !bytes.HasPrefix(leaf.Key(), esdtPrefix) {
			continue
		}

		marshalledData, err := userAccount.RetrieveValueFromDataTrieTracker(leaf.Key())
		esdtData := &esdt.ESDigitalToken{}
		err = marshaller.Unmarshal(esdtData, marshalledData)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshall in ESDigitalToken address:%s, err: %w",
				pubKeyConverter.Encode(account.AddressBytes()), err,
			)
		}

		balance = big.NewInt(0).Add(balance, esdtData.Value)
	}

	return balance, nil
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
