package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state/accounts"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/multiversx/mx-chain-tools-go/trieTools/trieToolsCommon"
)

const (
	rootHashLength         = 32
	defaultMainTrieWorkers = 4
	nodeTypeExtension      = byte(0)
	nodeTypeLeaf           = byte(1)
	nodeTypeBranch         = byte(2)
)

type decodedTrieNode struct {
	nodeType  byte
	childHash [][]byte
	leafValue []byte
}

type trieCopyStats struct {
	StoredNodes    int64
	MainTrieLeaves int64
	CodeLeaves     int64
	DataTrieRoots  int64
}

type trieCopyCoordinator struct {
	originDb storage.Storer
	targetDb storage.Storer

	ctx    context.Context
	cancel context.CancelFunc

	// wait group for data trie goroutines
	dataTrieWG sync.WaitGroup

	stats trieCopyStats

	mutErr   sync.Mutex
	firstErr error
}

func copyTrieAndAllDataTries(originDb storage.Storer, targetDb storage.Storer, rootHash []byte, maxGoroutines int) error {
	stats, err := copyTrieHashes(originDb, targetDb, rootHash, maxGoroutines)
	if err != nil {
		return err
	}

	log.Info("copied trie hierarchy",
		"rootHash", rootHash,
		"stored nodes", stats.StoredNodes,
		"main trie leaves", stats.MainTrieLeaves,
		"code leaves", stats.CodeLeaves,
		"data trie roots", stats.DataTrieRoots,
	)

	return nil
}

func copyTrieHashes(originDb storage.Storer, targetDb storage.Storer, rootHash []byte, maxGoroutines int) (trieCopyStats, error) {
	if len(rootHash) == 0 {
		return trieCopyStats{}, nil
	}

	copyCoordinator := newTrieCopyCoordinator(originDb, targetDb)
	defer copyCoordinator.cancel()

	// process the main trie sequentially (no worker pool)
	if err := copyCoordinator.copyMainTrieSequential(rootHash); err != nil {
		return trieCopyStats{}, err
	}

	// wait for spawned data trie goroutines to finish
	copyCoordinator.dataTrieWG.Wait()

	if copyCoordinator.firstError() != nil {
		return trieCopyStats{}, copyCoordinator.firstError()
	}

	return copyCoordinator.stats, nil
}

func newTrieCopyCoordinator(originDb storage.Storer, targetDb storage.Storer) *trieCopyCoordinator {
	ctx, cancel := context.WithCancel(context.Background())

	return &trieCopyCoordinator{
		originDb:   originDb,
		targetDb:   targetDb,
		ctx:        ctx,
		cancel:     cancel,
		dataTrieWG: sync.WaitGroup{},
	}
}

// copyMainTrieSequential traverses the main trie sequentially and writes nodes to the target DB.
// When a main-trie leaf references a data trie root, a goroutine is spawned to copy that full data trie.
func (copyCoordinator *trieCopyCoordinator) copyMainTrieSequential(rootHash []byte) error {
	pendingNodes := make([][]byte, 0, 1)
	pendingNodes = append(pendingNodes, append([]byte(nil), rootHash...))

	for len(pendingNodes) > 0 {
		lastIndex := len(pendingNodes) - 1
		nodeHash := pendingNodes[lastIndex]
		pendingNodes = pendingNodes[:lastIndex]

		decodedNode, err := copyCoordinator.copyNodeByHash(nodeHash)
		if err != nil {
			return err
		}

		// add children to pending list
		pendingNodes = appendChildHashes(pendingNodes, decodedNode.childHash)

		// if leaf, check for data trie and spawn copy if present
		if decodedNode.nodeType == nodeTypeLeaf {
			dataTrieRootHash, found := extractDataTrieRootHash(decodedNode.leafValue)
			if !found {
				atomic.AddInt64(&copyCoordinator.stats.CodeLeaves, 1)
			} else {
				rootCopy := append([]byte(nil), dataTrieRootHash...)
				copyCoordinator.dataTrieWG.Add(1)
				atomic.AddInt64(&copyCoordinator.stats.DataTrieRoots, 1)
				go func(rh []byte) {
					defer copyCoordinator.dataTrieWG.Done()
					if copyCoordinator.ctx.Err() != nil {
						return
					}
					if err := copyCoordinator.copyWholeTrie(rh); err != nil {
						copyCoordinator.setFirstError(fmt.Errorf("copy data trie %x: %w", rh, err))
						return
					}
					log.Debug("data trie copied", "rootHash", rh)
				}(rootCopy)
			}

			atomic.AddInt64(&copyCoordinator.stats.MainTrieLeaves, 1)
		}
	}

	return nil
}

func (copyCoordinator *trieCopyCoordinator) copyWholeTrie(rootHash []byte) error {
	pendingNodes := make([][]byte, 0, 1)
	pendingNodes = append(pendingNodes, append([]byte(nil), rootHash...))

	for len(pendingNodes) > 0 {
		if copyCoordinator.ctx.Err() != nil {
			return nil
		}

		lastIndex := len(pendingNodes) - 1
		nodeHash := pendingNodes[lastIndex]
		pendingNodes = pendingNodes[:lastIndex]

		decodedNode, err := copyCoordinator.copyNodeByHash(nodeHash)
		if err != nil {
			return err
		}

		pendingNodes = appendChildHashes(pendingNodes, decodedNode.childHash)
	}

	return nil
}

func (copyCoordinator *trieCopyCoordinator) copyNodeByHash(hash []byte) (decodedTrieNode, error) {
	serializedNode, err := copyCoordinator.originDb.Get(hash)
	if err != nil {
		return decodedTrieNode{}, fmt.Errorf("get trie node %x: %w", hash, err)
	}

	err = copyCoordinator.targetDb.Put(hash, serializedNode)
	if err != nil {
		return decodedTrieNode{}, fmt.Errorf("store trie node %x: %w", hash, err)
	}
	atomic.AddInt64(&copyCoordinator.stats.StoredNodes, 1)

	// TRACE when any hash/node is copied into the target DB
	log.Trace("copied trie node", "hash", hash)

	decodedNode, err := decodeTrieNode(serializedNode)
	if err != nil {
		return decodedTrieNode{}, fmt.Errorf("decode trie node %x: %w", hash, err)
	}

	return decodedNode, nil
}

// worker/enqueue helpers removed: main trie is processed sequentially and data tries are
// copied via dedicated goroutines spawned from `copyMainTrieSequential`.

func (copyCoordinator *trieCopyCoordinator) setFirstError(err error) {
	copyCoordinator.mutErr.Lock()
	defer copyCoordinator.mutErr.Unlock()

	if copyCoordinator.firstErr != nil {
		return
	}

	copyCoordinator.firstErr = err
	copyCoordinator.cancel()
}

func (copyCoordinator *trieCopyCoordinator) firstError() error {
	copyCoordinator.mutErr.Lock()
	defer copyCoordinator.mutErr.Unlock()

	return copyCoordinator.firstErr
}

func extractDataTrieRootHash(collapsedLeafValue []byte) ([]byte, bool) {
	userAccount := &accounts.UserAccountData{}
	err := trieToolsCommon.Marshaller.Unmarshal(userAccount, collapsedLeafValue)
	if err != nil || common.IsEmptyTrie(userAccount.RootHash) {
		return nil, false
	}

	return userAccount.RootHash, true
}

func decodeTrieNode(serializedNode []byte) (decodedTrieNode, error) {
	if len(serializedNode) == 0 {
		return decodedTrieNode{}, fmt.Errorf("empty trie node")
	}

	nodeType := serializedNode[len(serializedNode)-1]
	payload := serializedNode[:len(serializedNode)-1]

	switch nodeType {
	case nodeTypeBranch:
		collapsedBranch := &trie.CollapsedBn{}
		if err := trieToolsCommon.Marshaller.Unmarshal(collapsedBranch, payload); err != nil {
			return decodedTrieNode{}, err
		}

		childHashes := make([][]byte, 0, len(collapsedBranch.EncodedChildren))
		for _, childHash := range collapsedBranch.EncodedChildren {
			if len(childHash) == 0 {
				continue
			}

			copiedHash := append([]byte(nil), childHash...)
			childHashes = append(childHashes, copiedHash)
		}

		return decodedTrieNode{
			nodeType:  nodeType,
			childHash: childHashes,
		}, nil
	case nodeTypeExtension:
		collapsedExtension := &trie.CollapsedEn{}
		if err := trieToolsCommon.Marshaller.Unmarshal(collapsedExtension, payload); err != nil {
			return decodedTrieNode{}, err
		}

		childHashes := make([][]byte, 0, 1)
		if len(collapsedExtension.EncodedChild) != 0 {
			childHashes = append(childHashes, append([]byte(nil), collapsedExtension.EncodedChild...))
		}

		return decodedTrieNode{
			nodeType:  nodeType,
			childHash: childHashes,
		}, nil
	case nodeTypeLeaf:
		collapsedLeaf := &trie.CollapsedLn{}
		if err := trieToolsCommon.Marshaller.Unmarshal(collapsedLeaf, payload); err != nil {
			return decodedTrieNode{}, err
		}

		return decodedTrieNode{
			nodeType:  nodeType,
			leafValue: append([]byte(nil), collapsedLeaf.Value...),
		}, nil
	default:
		return decodedTrieNode{}, fmt.Errorf("unknown trie node type %d", nodeType)
	}
}

func appendChildHashes(destination [][]byte, childHashes [][]byte) [][]byte {
	for _, childHash := range childHashes {
		destination = append(destination, append([]byte(nil), childHash...))
	}

	return destination
}
