package main

import (
	"context"
	"fmt"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/errChan"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/vm/systemSmartContracts"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-tools-go/trieTools/trieToolsCommon"
	"github.com/urfave/cli"
	"os"
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

	globalSettingsShard0Address := "erd1lllllllllllllllllllllllllllllllllllllllllllllllllllsckry7t"
	globalSettingsShard0Account, err := GetAccountFromBech32String(globalSettingsShard0Address, shardsState[Shard0])
	if err != nil {
		return err
	}

	nonMigratedTokensMap := make(map[string][][]byte)

	ESDTPrefix := []byte("ELRONDesdt")
	for tokenType, tokenIds := range tokensMap {
		if tokenType == core.FungibleESDT {
			continue
		}

		tokenTypeAsInt, err := core.ConvertESDTTypeToUint32(tokenType)
		if err != nil {
			log.Error("can not convert token type to int", "tokenType", tokenType, "error", err)
			return err
		}

		for _, tokenId := range tokenIds {
			key := append(ESDTPrefix, tokenId...)
			log.Info("checking token", "tokenId", tokenId, "tokenType", tokenType, "key", key)
			value, _, err := globalSettingsShard0Account.RetrieveValue(key)
			if err != nil {
				log.Error("can not get token data", "tokenId", tokenId, "error", err)
				return err
			}

			if len(value) != 2 {
				nonMigratedTokensMap[tokenType] = append(nonMigratedTokensMap[tokenType], tokenId)
				log.Warn("token data has wrong length", "tokenId", tokenId, "value", value)
				continue
			}

			retrievedTokenType := value[1]
			if uint32(retrievedTokenType) == 0 {
				nonMigratedTokensMap[tokenType] = append(nonMigratedTokensMap[tokenType], tokenId)
				log.Warn("token type is 0", "tokenId", tokenId)
				continue
			}

			if uint32(retrievedTokenType)-1 != tokenTypeAsInt {
				nonMigratedTokensMap[tokenType] = append(nonMigratedTokensMap[tokenType], tokenId)
				log.Warn("token type is not the same", "tokenId", tokenId, "retrievedTokenType", retrievedTokenType, "tokenType", tokenType)
			}
		}
		log.Info("tokens", "tokenType", tokenType, "tokenIds", tokenIds)
	}
	for _, tokenIds := range nonMigratedTokensMap {
		for _, tokenId := range tokenIds {
			fmt.Println(string(tokenId))
		}
	}

	return nil
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
