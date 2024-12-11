package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-tools-go/trieTools/trieToolsCommon"
)

func NewShardState(rootHash []byte, workingDir, dbDir string) (state.AccountsAdapter, error) {
	createStorerFlags := trieToolsCommon.ContextFlagsConfig{
		WorkingDir: workingDir,
		DbDir:      dbDir,
	}

	maxDBValue, err := trieToolsCommon.GetMaxDBValue(filepath.Join(createStorerFlags.WorkingDir, createStorerFlags.DbDir), log)
	if err != nil {
		return nil, err
	}

	storer, err := trieToolsCommon.CreatePruningStorer(createStorerFlags, maxDBValue)
	if err != nil {
		return nil, err
	}

	enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
			return true
		},
	}
	tr, err := trieToolsCommon.CreateTrie(storer, enableEpochsHandler)
	if err != nil {
		return nil, err
	}

	accDb, err := trieToolsCommon.NewAccountsAdapter(tr, enableEpochsHandler)
	if err != nil {
		return nil, err
	}

	rootHashHolder := holders.NewDefaultRootHashesHolder(rootHash)
	err = accDb.RecreateTrie(rootHashHolder)
	if err != nil {
		return nil, err
	}

	return accDb, nil
}

func LoadStateForAllShards(flagsConfig contextFlagsConfig) (map[ShardID]state.AccountsAdapter, error) {
	log.Info("starting loading the state for each shard", "pid", os.Getpid())

	shardsIds := make([]ShardID, 4)
	shardsIds[0] = Shard0
	shardsIds[1] = Shard1
	shardsIds[2] = Shard2
	shardsIds[3] = Meta

	shardsState := make(map[ShardID]state.AccountsAdapter, len(shardsIds))
	workingDir := filepath.Join(flagsConfig.WorkingDir, flagsConfig.DbDir)
	for _, shardId := range shardsIds {
		rootHash, err := getRootHashForShard(flagsConfig, shardId)
		if err != nil {
			return nil, fmt.Errorf("%w when decoding the provided hex root hash", err)
		}
		if len(rootHash) != rootHashLength {
			return nil, fmt.Errorf("wrong root hash length: expected %d, got %d", rootHashLength, len(rootHash))
		}

		ss, err := NewShardState(rootHash, workingDir, string(shardId))
		if err != nil {
			return nil, err
		}

		shardsState[shardId] = ss
	}

	log.Info("loading state successful")
	return shardsState, nil
}

func getRootHashForShard(config contextFlagsConfig, id ShardID) ([]byte, error) {
	switch id {
	case Shard0:
		return hex.DecodeString(config.HexRootHash0)
	case Shard1:
		return hex.DecodeString(config.HexRootHash1)
	case Shard2:
		return hex.DecodeString(config.HexRootHash2)
	case Meta:
		return hex.DecodeString(config.HexRootHashMeta)
	default:
		return nil, fmt.Errorf("unknown shard id: %v", id)
	}
}

func GetAccountFromBech32String(address string, accDb state.AccountsAdapter) (state.UserAccountHandler, error) {
	addressConverter, err := pubkeyConverter.NewBech32PubkeyConverter(addressLength, trieToolsCommon.WalletHRP)
	if err != nil {
		return nil, err
	}
	accBytes, err := addressConverter.Decode(address)
	if err != nil {
		return nil, err
	}
	acc, err := accDb.GetExistingAccount(accBytes)
	if err != nil {
		return nil, err
	}
	userAcc, ok := acc.(state.UserAccountHandler)
	if !ok {
		return nil, fmt.Errorf("account is not a user account")
	}
	return userAcc, nil
}
