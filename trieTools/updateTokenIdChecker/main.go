package main

import (
	"context"
	"fmt"
	"math"
	"os"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/errChan"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/vm/systemSmartContracts"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-tools-go/trieTools/trieToolsCommon"
	"github.com/urfave/cli"
)

var log = logger.GetOrCreate("trie")

const (
	logFilePrefix  = "trie"
	rootHashLength = 32
	addressLength  = 32
)

func main() {
	app := cli.NewApp()
	app.Name = "Trie stats CLI app"
	app.Usage = "This is the entry point for the tool that prints stats about the state"
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

	shardsState, err := LoadStateForAllShards(flagsConfig)
	if err != nil {
		return err
	}

	systemEsdtAddress := "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqzllls8a5w6u"
	systemEsdtAccount, err := GetAccountFromBech32String(systemEsdtAddress, shardsState[Meta])
	if err != nil {
		return err
	}

	tokensMap, err := getAllESDTsFromSystemEsdtAccount(systemEsdtAccount)
	if err != nil {
		return err
	}

	systemAccounts, err := getSystemAccountsForShards(shardsState)
	if err != nil {
		return err
	}

	checkUpdateTokenTypeCalled(systemAccounts, tokensMap)
	return nil
}

func checkUpdateTokenTypeCalled(systemAccounts map[ShardID]state.UserAccountHandler, tokensMap map[string][][]byte) {
	esdtPrefix := []byte("ELRONDesdt")
	for tokenType, tokenIds := range tokensMap {
		if tokenType == core.FungibleESDT {
			continue
		}

		tokenTypeAsInt, err := core.ConvertESDTTypeToUint32(tokenType)
		if err != nil {
			log.Error("can not convert token type to int", "tokenType", tokenType, "error", err)
			continue
		}

		for _, tokenId := range tokenIds {
			key := append(esdtPrefix, tokenId...)

			checkTokenInAllShards(key, tokenTypeAsInt, systemAccounts)
		}
		log.Info("tokens", "tokenType", tokenType, "numTokens", len(tokenIds))
	}
}

func checkTokenInAllShards(tokenKey []byte, tokenType uint32, systemAccounts map[ShardID]state.UserAccountHandler) {
	log.Info("checking token", "tokenId", string(tokenKey), "tokenType", tokenType)

	for shardId, systemAccount := range systemAccounts {
		value, _, err := systemAccount.RetrieveValue(tokenKey)
		if err != nil {
			log.Error("can not get token data", "tokenId", tokenKey, "error", err)
			continue
		}

		if len(value) != 2 {
			log.Error("token data has wrong length", "tokenId", tokenKey, "value", value)
			continue
		}

		retrievedTokenType := value[1]
		esdtTokenType, err := convertToESDTTokenType(uint32(retrievedTokenType))
		if err != nil {
			log.Error("can not convert token type to int", "tokenType", tokenType, "error", err)
			continue
		}

		if esdtTokenType != tokenType {
			log.Warn("token type is not the same", "tokenId", tokenKey, "retrievedTokenType", retrievedTokenType, "tokenType", tokenType, "shard", shardId)
		}
	}
}

func getAllESDTsFromSystemEsdtAccount(systemEsdtAccount state.UserAccountHandler) (map[string][][]byte, error) {
	tokens := make(map[string][][]byte)

	iteratorChannels := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity),
		ErrChan:    errChan.NewErrChanWrapper(),
	}
	err := systemEsdtAccount.GetAllLeaves(iteratorChannels, context.Background())
	if err != nil {
		return nil, err
	}

	for leaf := range iteratorChannels.LeavesChan {
		data := &systemSmartContracts.ESDTDataV2{}
		errUnmarshal := trieToolsCommon.Marshaller.Unmarshal(data, leaf.Value())
		if errUnmarshal == nil {
			tokens[string(data.TokenType)] = append(tokens[string(data.TokenType)], leaf.Key())
			continue
		}

		dataV1 := &systemSmartContracts.ESDTDataV1{}
		errUnmarshal = trieToolsCommon.Marshaller.Unmarshal(dataV1, leaf.Value())
		if errUnmarshal == nil {
			tokens[string(data.TokenType)] = append(tokens[string(data.TokenType)], leaf.Key())
			continue
		}

		log.Warn("can not unmarshall data", "key", string(leaf.Key()), "value", leaf.Value())
	}

	return tokens, nil
}

func getSystemAccountsForShards(shardsState map[ShardID]state.AccountsAdapter) (map[ShardID]state.UserAccountHandler, error) {
	globalSettingsShard0Address := "erd1lllllllllllllllllllllllllllllllllllllllllllllllllllsckry7t"
	globalSettingsShard0Account, err := GetAccountFromBech32String(globalSettingsShard0Address, shardsState[Shard0])
	if err != nil {
		return nil, err
	}
	globalSettingsShard1Account, err := GetAccountFromBech32String(globalSettingsShard0Address, shardsState[Shard1])
	if err != nil {
		return nil, err
	}
	globalSettingsShard2Account, err := GetAccountFromBech32String(globalSettingsShard0Address, shardsState[Shard2])
	if err != nil {
		return nil, err
	}

	systemAccounts := make(map[ShardID]state.UserAccountHandler)
	systemAccounts[Shard0] = globalSettingsShard0Account
	systemAccounts[Shard1] = globalSettingsShard1Account
	systemAccounts[Shard2] = globalSettingsShard2Account
	return systemAccounts, nil
}

func convertToESDTTokenType(esdtType uint32) (uint32, error) {
	switch esdtType {
	case 0:
		return 0, fmt.Errorf("token type not set inside global settings handler")
	case 1:
		return uint32(core.Fungible), nil
	case 2:
		return uint32(core.NonFungible), nil
	case 3:
		return uint32(core.NonFungibleV2), nil
	case 4:
		return uint32(core.MetaFungible), nil
	case 5:
		return uint32(core.SemiFungible), nil
	case 6:
		return uint32(core.DynamicNFT), nil
	case 7:
		return uint32(core.DynamicSFT), nil
	case 8:
		return uint32(core.DynamicMeta), nil
	default:
		return math.MaxUint32, fmt.Errorf("invalid esdt type: %d", esdtType)
	}
}
