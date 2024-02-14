package trie

import (
	"context"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/errChan"
	"github.com/multiversx/mx-chain-go/state/accounts"
	"github.com/multiversx/mx-chain-go/state/parsers"
	"github.com/multiversx/mx-chain-go/trie/keyBuilder"
)

type trieWrapper struct {
	trie common.Trie
}

func newTrieWrapper(t common.Trie) *trieWrapper {
	return &trieWrapper{trie: t}
}

// IsRootHashAvailable checks whether a rootHash is available in the trie database (e.g. for trie reconstruction)
func (tw *trieWrapper) IsRootHashAvailable(rootHash []byte) bool {
	storageManager := tw.trie.GetStorageManager()
	_, err := storageManager.Get(rootHash)
	if err != nil {
		return false
	}

	return true
}

func (tw *trieWrapper) GetUserAccounts(rootHash []byte, predicate func(*accounts.UserAccountData) bool) ([]*accounts.UserAccountData, error) {
	iteratorChannels := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity),
		ErrChan:    errChan.NewErrChanWrapper(),
	}
	err := tw.trie.GetAllLeavesOnChannel(iteratorChannels, context.Background(), rootHash, keyBuilder.NewDisabledKeyBuilder(), parsers.NewMainTrieLeafParser())
	if err != nil {
		return nil, err
	}

	users := make([]*accounts.UserAccountData, 0)

	for keyValue := range iteratorChannels.LeavesChan {
		user := &accounts.UserAccountData{}
		errUnmarshal := marshaller.Unmarshal(user, keyValue.Value())
		if errUnmarshal != nil {
			// Probably a code node
			continue
		}

		if predicate(user) {
			users = append(users, user)
		}
	}

	err = iteratorChannels.ErrChan.ReadFromChanNonBlocking()
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (tw *trieWrapper) Close() {
	err := tw.trie.Close()
	if err != nil {
		log.Error("cannot close trie", "err", err)
	}
}
