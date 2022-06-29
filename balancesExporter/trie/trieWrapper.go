package trie

import (
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/state"
)

type trieWrapper struct {
	trie common.Trie
}

func newTrieWrapper(t common.Trie) *trieWrapper {
	return &trieWrapper{trie: t}
}

func (tw *trieWrapper) IsRootHashAvailable(rootHash []byte) bool {
	_, err := tw.trie.GetAllLeavesOnChannel(rootHash)
	if err != nil {
		return false
	}

	return true
}

func (tw *trieWrapper) GetAllUserAccounts(rootHash []byte) ([]*state.UserAccountData, error) {
	ch, err := tw.trie.GetAllLeavesOnChannel(rootHash)
	if err != nil {
		return nil, err
	}

	users := make([]*state.UserAccountData, 0)

	for keyValue := range ch {
		user := &state.UserAccountData{}
		errUnmarshal := marshaller.Unmarshal(user, keyValue.Value())
		if errUnmarshal != nil {
			// Probably a code node
			continue
		}

		users = append(users, user)
	}

	return users, nil
}

func (tw *trieWrapper) Close() {
	err := tw.trie.Close()
	if err != nil {
		log.Error("cannot close trie", "err", err)
	}
}
