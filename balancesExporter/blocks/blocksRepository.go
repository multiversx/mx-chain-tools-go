package blocks

import (
	"fmt"
	"path"
	"sort"

	"github.com/ElrondNetwork/elrond-go-core/data"
	dataBlock "github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

var (
	marshaller = &marshal.GogoProtoMarshalizer{}
)

type ArgsNewBlocksRepository struct {
	DbPath      string
	Epoch       uint32
	Shard       uint32
	TrieWrapper trieWrapper
}

type blocksRepository struct {
	dbPath string
	epoch  uint32
	shard  uint32
	trie   trieWrapper
}

func NewBlocksRepository(args ArgsNewBlocksRepository) *blocksRepository {
	return &blocksRepository{
		dbPath: args.DbPath,
		epoch:  args.Epoch,
		shard:  args.Shard,
		trie:   args.TrieWrapper,
	}
}

func (repository *blocksRepository) FindBestBlock() (data.HeaderHandler, error) {
	blocksInEpoch, err := repository.loadBlocksInEpoch()
	if err != nil {
		return nil, err
	}

	eligibleBlocks := repository.findEligibleBlocks(blocksInEpoch)

	for _, block := range eligibleBlocks {
		// For ease of reasoning, we exclude blocks that contain scheduled (or processed) miniblocks.
		if !hasScheduledMiniblocks(block) {
			return block, nil
		}
	}

	return nil, fmt.Errorf("no best block found")
}

func (repository *blocksRepository) findEligibleBlocks(headers []data.HeaderHandler) []data.HeaderHandler {
	log.Info("findEligibleBlocks() started")

	eligibleBlocks := make([]data.HeaderHandler, 0)

	for _, header := range headers {
		rootHash := header.GetRootHash()

		if repository.trie.IsRootHashAvailable(rootHash) {
			eligibleBlocks = append(eligibleBlocks, header)
		}
	}

	log.Info("findEligibleBlocks()", "found blocks", len(eligibleBlocks))

	return eligibleBlocks
}

func (repository *blocksRepository) loadBlocksInEpoch() ([]data.HeaderHandler, error) {
	marshalizedBlocks, err := repository.loadMarshalizedBlocksInEpoch()
	if err != nil {
		return nil, err
	}

	headers := make([]data.HeaderHandler, 0)

	for _, bytes := range marshalizedBlocks {
		header := &dataBlock.HeaderV2{}
		err := marshaller.Unmarshal(header, bytes)
		if err != nil {
			return nil, err
		}

		headers = append(headers, header)
	}

	// Sort blocks by nonce
	sort.Slice(headers, func(i, j int) bool {
		return headers[i].GetNonce() < headers[j].GetNonce()
	})

	return headers, nil
}

func (repository *blocksRepository) loadMarshalizedBlocksInEpoch() ([][]byte, error) {
	cacheConfig := getCacheConfig()
	unitPath := repository.getStorageUnitPath()
	dbConfig := getDbConfig(unitPath)

	unit, err := storageUnit.NewStorageUnitFromConf(cacheConfig, dbConfig)
	if err != nil {
		return nil, err
	}

	values := make([][]byte, 0)

	unit.RangeKeys(func(key, value []byte) bool {
		values = append(values, value)
		return true
	})

	return values, nil
}

func (repository *blocksRepository) getStorageUnitPath() string {
	epochPart := fmt.Sprintf("Epoch_%d", repository.epoch)
	shardPart := fmt.Sprintf("Shard_%d", repository.shard)
	unitPath := path.Join(repository.dbPath, epochPart, shardPart, "BlockHeaders")
	return unitPath
}

func hasScheduledMiniblocks(block data.HeaderHandler) bool {
	miniblocks := block.GetMiniBlockHeaderHandlers()

	for _, miniblock := range miniblocks {
		processingType := dataBlock.ProcessingType(miniblock.GetProcessingType())

		if processingType == dataBlock.Scheduled {
			return true
		}
		if processingType == dataBlock.Processed {
			return true
		}
	}

	return false
}
