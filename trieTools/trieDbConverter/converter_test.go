package main

import (
	"bytes"
	"errors"
	"flag"
	"math/big"
	"sync"
	"testing"

	coredata "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/state/accounts"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/multiversx/mx-chain-tools-go/trieTools/trieToolsCommon"
	"github.com/urfave/cli"
)

func TestDecodeTrieNode(t *testing.T) {
	t.Run("branch node exposes all child hashes", func(t *testing.T) {
		serializedNode := mustSerializeTrieNode(t, &trie.CollapsedBn{
			EncodedChildren: [][]byte{{1, 2}, nil, {3, 4}},
		}, nodeTypeBranch)

		decodedNode, err := decodeTrieNode(serializedNode)
		if err != nil {
			t.Fatalf("decodeTrieNode returned error: %v", err)
		}

		expectedHashes := [][]byte{{1, 2}, {3, 4}}
		if !equalHashes(decodedNode.childHash, expectedHashes) {
			t.Fatalf("unexpected child hashes: got %v, want %v", decodedNode.childHash, expectedHashes)
		}
	})

	t.Run("extension node exposes encoded child", func(t *testing.T) {
		serializedNode := mustSerializeTrieNode(t, &trie.CollapsedEn{
			EncodedChild: []byte{9, 8, 7},
		}, nodeTypeExtension)

		decodedNode, err := decodeTrieNode(serializedNode)
		if err != nil {
			t.Fatalf("decodeTrieNode returned error: %v", err)
		}

		expectedHashes := [][]byte{{9, 8, 7}}
		if !equalHashes(decodedNode.childHash, expectedHashes) {
			t.Fatalf("unexpected child hashes: got %v, want %v", decodedNode.childHash, expectedHashes)
		}
	})

	t.Run("leaf node exposes leaf value", func(t *testing.T) {
		serializedNode := mustSerializeTrieNode(t, &trie.CollapsedLn{
			Value: []byte("leaf-value"),
		}, nodeTypeLeaf)

		decodedNode, err := decodeTrieNode(serializedNode)
		if err != nil {
			t.Fatalf("decodeTrieNode returned error: %v", err)
		}

		if !bytes.Equal(decodedNode.leafValue, []byte("leaf-value")) {
			t.Fatalf("unexpected leaf value: got %q", decodedNode.leafValue)
		}
	})

	t.Run("unknown node type returns error", func(t *testing.T) {
		_, err := decodeTrieNode([]byte{1, 2, 99})
		if err == nil {
			t.Fatal("expected error for unknown node type")
		}
	})
}

func TestCopyTrieHashesCopiesMainTrieAndDataTrie(t *testing.T) {
	originDb := newMemoryStorer()
	targetDb := newMemoryStorer()

	dataTrieRootHash := []byte("data-root")
	accountLeafHash := []byte("account-leaf")
	codeLeafHash := []byte("code-leaf")
	rootHash := []byte("main-root")

	accountValue := mustMarshalAccount(t, &accounts.UserAccountData{
		Nonce:    1,
		Balance:  big.NewInt(5),
		RootHash: dataTrieRootHash,
	})

	originDb.mustPut(t, rootHash, mustSerializeTrieNode(t, &trie.CollapsedBn{
		EncodedChildren: [][]byte{accountLeafHash, codeLeafHash},
	}, nodeTypeBranch))
	originDb.mustPut(t, accountLeafHash, mustSerializeTrieNode(t, &trie.CollapsedLn{Value: accountValue}, nodeTypeLeaf))
	originDb.mustPut(t, codeLeafHash, mustSerializeTrieNode(t, &trie.CollapsedLn{Value: []byte("code")}, nodeTypeLeaf))
	originDb.mustPut(t, dataTrieRootHash, mustSerializeTrieNode(t, &trie.CollapsedLn{Value: []byte("data-leaf")}, nodeTypeLeaf))

	stats, workerConfig, err := copyTrieHashes(originDb, targetDb, rootHash, defaultMaxTrieCopyGoroutines)
	if err != nil {
		t.Fatalf("copyTrieHashes returned error: %v", err)
	}
	if workerConfig.MaxGoroutines != defaultMaxTrieCopyGoroutines {
		t.Fatalf("unexpected max goroutines: got %d, want %d", workerConfig.MaxGoroutines, defaultMaxTrieCopyGoroutines)
	}

	if stats.StoredNodes != 4 {
		t.Fatalf("unexpected stored nodes: got %d, want 4", stats.StoredNodes)
	}
	if stats.MainTrieLeaves != 2 {
		t.Fatalf("unexpected main trie leaves: got %d, want 2", stats.MainTrieLeaves)
	}
	if stats.CodeLeaves != 1 {
		t.Fatalf("unexpected code leaves: got %d, want 1", stats.CodeLeaves)
	}
	if stats.DataTrieRoots != 1 {
		t.Fatalf("unexpected data trie roots: got %d, want 1", stats.DataTrieRoots)
	}

	for _, hash := range [][]byte{rootHash, accountLeafHash, codeLeafHash, dataTrieRootHash} {
		storedValue, err := targetDb.Get(hash)
		if err != nil {
			t.Fatalf("expected hash %q in target db, got error %v", hash, err)
		}
		originValue, _ := originDb.Get(hash)
		if !bytes.Equal(storedValue, originValue) {
			t.Fatalf("unexpected stored node for hash %q", hash)
		}
	}
}

func TestExtractDataTrieRootHash(t *testing.T) {
	expectedRootHash := []byte{1, 2, 3, 4}
	account := &accounts.UserAccountData{
		Nonce:    7,
		Balance:  big.NewInt(42),
		RootHash: expectedRootHash,
	}

	marshalledAccount, err := trieToolsCommon.Marshaller.Marshal(account)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	rootHash, found := extractDataTrieRootHash(marshalledAccount)
	if !found {
		t.Fatal("expected account root hash to be discovered")
	}
	if !bytes.Equal(rootHash, expectedRootHash) {
		t.Fatalf("unexpected root hash: got %v, want %v", rootHash, expectedRootHash)
	}

	rootHash[0] = 99
	if bytes.Equal(rootHash, expectedRootHash) {
		t.Fatal("expected a defensive copy of the root hash")
	}
}

func TestExtractDataTrieRootHashReturnsFalseForNonAccountLeaf(t *testing.T) {
	rootHash, found := extractDataTrieRootHash([]byte("not-an-account"))
	if found {
		t.Fatalf("expected no root hash, got %v", rootHash)
	}
}

func TestExtractDataTrieRootHashReturnsFalseForAccountWithoutDataTrie(t *testing.T) {
	account := &accounts.UserAccountData{
		Nonce:   11,
		Balance: big.NewInt(55),
	}

	marshalledAccount, err := trieToolsCommon.Marshaller.Marshal(account)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	rootHash, found := extractDataTrieRootHash(marshalledAccount)
	if found {
		t.Fatalf("expected no root hash, got %v", rootHash)
	}
}

func TestSplitTrieCopyWorkers(t *testing.T) {
	t.Run("rejects less than two goroutines", func(t *testing.T) {
		_, err := splitTrieCopyWorkers(1)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("keeps one goroutine for data tries", func(t *testing.T) {
		workerConfig, err := splitTrieCopyWorkers(2)
		if err != nil {
			t.Fatalf("splitTrieCopyWorkers returned error: %v", err)
		}

		if workerConfig.MainTrieWorker != 1 || workerConfig.DataTrieWorker != 1 {
			t.Fatalf("unexpected worker split: %+v", workerConfig)
		}
	})

	t.Run("caps main trie workers and gives the rest to data tries", func(t *testing.T) {
		workerConfig, err := splitTrieCopyWorkers(defaultMaxTrieCopyGoroutines)
		if err != nil {
			t.Fatalf("splitTrieCopyWorkers returned error: %v", err)
		}

		if workerConfig.MainTrieWorker != defaultMainTrieWorkers {
			t.Fatalf("unexpected main trie workers: got %d, want %d", workerConfig.MainTrieWorker, defaultMainTrieWorkers)
		}
		expectedDataWorkers := defaultMaxTrieCopyGoroutines - defaultMainTrieWorkers
		if workerConfig.DataTrieWorker != expectedDataWorkers {
			t.Fatalf("unexpected data trie workers: got %d, want %d", workerConfig.DataTrieWorker, expectedDataWorkers)
		}
	})
}

func TestGetFlagsConfigReadsMaxGoroutines(t *testing.T) {
	t.Run("default value is used when flag is missing", func(t *testing.T) {
		ctx := createCLIContextForConverterFlags(t)

		flagsConfig := getFlagsConfig(ctx)
		if flagsConfig.MaxGoroutines != defaultMaxTrieCopyGoroutines {
			t.Fatalf("unexpected default max goroutines: got %d, want %d", flagsConfig.MaxGoroutines, defaultMaxTrieCopyGoroutines)
		}
	})

	t.Run("explicit flag overrides default", func(t *testing.T) {
		ctx := createCLIContextForConverterFlags(t)
		err := ctx.GlobalSet(maxGoroutines.Name, "17")
		if err != nil {
			t.Fatalf("GlobalSet returned error: %v", err)
		}

		flagsConfig := getFlagsConfig(ctx)
		if flagsConfig.MaxGoroutines != 17 {
			t.Fatalf("unexpected max goroutines: got %d, want 17", flagsConfig.MaxGoroutines)
		}
	})
}

func createCLIContextForConverterFlags(t *testing.T) *cli.Context {
	t.Helper()

	app := cli.NewApp()
	app.Flags = getFlags()
	flagSet := flag.NewFlagSet("trie-db-converter", flag.ContinueOnError)

	for _, cliFlag := range app.Flags {
		cliFlag.Apply(flagSet)
	}

	return cli.NewContext(app, flagSet, nil)
}

func mustSerializeTrieNode(t *testing.T, node interface{}, nodeType byte) []byte {
	t.Helper()

	marshalledNode, err := trieToolsCommon.Marshaller.Marshal(node)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	return append(marshalledNode, nodeType)
}

func mustMarshalAccount(t *testing.T, account *accounts.UserAccountData) []byte {
	t.Helper()

	marshalledAccount, err := trieToolsCommon.Marshaller.Marshal(account)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	return marshalledAccount
}

func equalHashes(first [][]byte, second [][]byte) bool {
	if len(first) != len(second) {
		return false
	}

	for index := range first {
		if !bytes.Equal(first[index], second[index]) {
			return false
		}
	}

	return true
}

type memoryStorer struct {
	mut   sync.Mutex
	data  map[string][]byte
	epoch uint32
}

func newMemoryStorer() *memoryStorer {
	return &memoryStorer{data: make(map[string][]byte)}
}

func (ms *memoryStorer) Put(key, data []byte) error {
	ms.mut.Lock()
	defer ms.mut.Unlock()

	ms.data[string(key)] = append([]byte(nil), data...)
	return nil
}

func (ms *memoryStorer) PutInEpoch(key, data []byte, _ uint32) error { return ms.Put(key, data) }
func (ms *memoryStorer) Get(key []byte) ([]byte, error) {
	ms.mut.Lock()
	defer ms.mut.Unlock()

	data, ok := ms.data[string(key)]
	if !ok {
		return nil, errors.New("not found")
	}

	return append([]byte(nil), data...), nil
}
func (ms *memoryStorer) Has(key []byte) error {
	_, err := ms.Get(key)
	return err
}
func (ms *memoryStorer) SearchFirst(key []byte) ([]byte, error)  { return ms.Get(key) }
func (ms *memoryStorer) RemoveFromCurrentEpoch(key []byte) error { return ms.Remove(key) }
func (ms *memoryStorer) Remove(key []byte) error {
	ms.mut.Lock()
	defer ms.mut.Unlock()

	delete(ms.data, string(key))
	return nil
}
func (ms *memoryStorer) ClearCache()                                       {}
func (ms *memoryStorer) DestroyUnit() error                                { return nil }
func (ms *memoryStorer) GetFromEpoch(key []byte, _ uint32) ([]byte, error) { return ms.Get(key) }
func (ms *memoryStorer) GetBulkFromEpoch(_ [][]byte, _ uint32) ([]coredata.KeyValuePair, error) {
	return nil, nil
}
func (ms *memoryStorer) GetOldestEpoch() (uint32, error) { return ms.epoch, nil }
func (ms *memoryStorer) RangeKeys(handler func(key []byte, val []byte) bool) {
	ms.mut.Lock()
	defer ms.mut.Unlock()

	for key, value := range ms.data {
		if !handler([]byte(key), append([]byte(nil), value...)) {
			return
		}
	}
}
func (ms *memoryStorer) Close() error         { return nil }
func (ms *memoryStorer) IsInterfaceNil() bool { return ms == nil }

func (ms *memoryStorer) mustPut(t *testing.T, key, value []byte) {
	t.Helper()

	if err := ms.Put(key, value); err != nil {
		t.Fatalf("Put returned error: %v", err)
	}
}
