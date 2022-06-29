package trie

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/common"
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

func (tw *trieWrapper) GetAllLeaves(rootHash []byte) ([]core.KeyValueHolder, error) {
	ch, err := tw.trie.GetAllLeavesOnChannel(rootHash)
	if err != nil {
		return nil, err
	}

	pairs := make([]core.KeyValueHolder, 0)

	for keyValue := range ch {
		pairs = append(pairs, keyValue)
	}

	return pairs, nil
}

func (tw *trieWrapper) Close() {
	err := tw.trie.Close()
	if err != nil {
		log.Error("cannot close trie", "err", err)
	}
}
