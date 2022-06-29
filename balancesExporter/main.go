package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/core/pubkeyConverter"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/pruning"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/trie"
	"github.com/ElrondNetwork/elrond-tools-go/balancesExporter/components"
	"github.com/urfave/cli"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

const (
	maxTrieLevelInMemory = 5
	rootHashLength       = 32
	addressLength        = 32
)

func main() {
	app := cli.NewApp()
	app.Name = "Balances exporter CLI app"
	app.Usage = "Tool for exporting balances of accounts (given a Node db)"
	app.Flags = getCliFlags()
	app.Authors = []cli.Author{
		{
			Name:  "The Elrond Team",
			Email: "contact@elrond.com",
		},
	}

	app.Action = startProcess

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func startProcess(c *cli.Context) error {
	cliFlags := getFlagsConfig(c)
	dbPath := cliFlags.DbPath
	epoch := cliFlags.Epoch
	shard := cliFlags.Shard

	fileLogging, err := initializeLogger(cliFlags.WorkingDir, cliFlags.LogLevel)
	if err != nil {
		return err
	}
	defer func() { _ = fileLogging.Close() }()

	blocks, err := loadBlocksInEpoch(dbPath, epoch, shard)
	if err != nil {
		return err
	}

	eligibleBlocks, err := findEligibleBlocks(dbPath, epoch, blocks)
	if err != nil {
		return err
	}

	chosenBlock, err := askBlockChoice(eligibleBlocks)
	if err != nil {
		return err
	}

	err = exportBalances(chosenBlock, dbPath, epoch)
	if err != nil {
		return err
	}

	return nil
}

func loadBlocksInEpoch(dbPath string, epoch uint32, shard uint32) ([]data.HeaderHandler, error) {
	marshalizedBlocks, err := loadMarshalizedBlocksInEpoch(dbPath, epoch, shard)
	if err != nil {
		return nil, err
	}

	headers := make([]data.HeaderHandler, 0)

	for _, bytes := range marshalizedBlocks {
		header := &block.HeaderV2{}
		err := marshaller.Unmarshal(header, bytes)
		if err != nil {
			return nil, err
		}

		headers = append(headers, header)
	}

	sort.Slice(headers, func(i, j int) bool {
		return headers[i].GetNonce() < headers[j].GetNonce()
	})

	return headers, nil
}

func loadMarshalizedBlocksInEpoch(dbPath string, epoch uint32, shard uint32) ([][]byte, error) {
	epochPart := fmt.Sprintf("Epoch_%d", epoch)
	shardPart := fmt.Sprintf("Shard_%d", shard)
	unitPath := path.Join(dbPath, epochPart, shardPart, "BlockHeaders")

	unit, err := storageUnit.NewStorageUnitFromConf(cacheConfig, storageUnit.DBConfig{
		FilePath:          unitPath,
		Type:              "LvlDBSerial",
		BatchDelaySeconds: 2,
		MaxBatchSize:      45000,
		MaxOpenFiles:      10,
	})
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

func findEligibleBlocks(dbPath string, epoch uint32, headers []data.HeaderHandler) ([]data.HeaderHandler, error) {
	tr, err := getTrie(dbPath, epoch)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tr.Close() }()

	log.Info("findEligibleBlocks() started")

	eligibleBlocks := make([]data.HeaderHandler, 0)

	for _, header := range headers {
		_, err := tr.GetAllLeavesOnChannel(header.GetRootHash())
		if err != nil {
			continue
		}

		eligibleBlocks = append(eligibleBlocks, header)
	}

	log.Info("findEligibleBlocks()", "found blocks", len(eligibleBlocks))
	return eligibleBlocks, nil
}

func askBlockChoice(headers []data.HeaderHandler) (data.HeaderHandler, error) {
	fmt.Println("> Choose a block:")

	for i, header := range headers {
		hasScheduled := header.HasScheduledMiniBlocks()
		fmt.Printf("%d: nonce = %d, hasScheduled = %t\n", i, header.GetNonce(), hasScheduled)
	}

	fmt.Println("> Choose a block:")

	choice, err := readNumber()
	if err != nil {
		return nil, err
	}

	return headers[choice], nil
}

func readNumber() (int, error) {
	line := readLine()
	return strconv.Atoi(line)
}

func readLine() string {
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}

func exportBalances(header data.HeaderHandler, dbPath string, epoch uint32) error {
	rootHash := header.GetRootHash()

	addressConverter, err := pubkeyConverter.NewBech32PubkeyConverter(addressLength, log)
	if err != nil {
		return err
	}

	tr, err := getTrie(dbPath, epoch)
	if err != nil {
		return err
	}
	defer func() { _ = tr.Close() }()

	ch, err := tr.GetAllLeavesOnChannel(rootHash)
	if err != nil {
		return err
	}

	numAccountsOnMainTrie := 0
	numCodeNodes := 0
	dataTriesRootHashes := make(map[string][]byte)
	numDataTriesLeaves := 0
	for kv := range ch {
		numAccountsOnMainTrie++

		userAccount := &state.UserAccountData{}
		errUnmarshal := marshaller.Unmarshal(userAccount, kv.Value())
		if errUnmarshal != nil {
			// probably a code node
			numCodeNodes++
			continue
		}
		if len(userAccount.RootHash) == 0 {
			continue
		}

		address := addressConverter.Encode(kv.Key())
		dataTriesRootHashes[address] = userAccount.RootHash
		fmt.Println(userAccount.Balance.String())
	}

	log.Info("parsed main trie",
		"num accounts", numAccountsOnMainTrie,
		"num code nodes", numCodeNodes,
		"num data tries", len(dataTriesRootHashes))

	if len(dataTriesRootHashes) == 0 {
		return nil
	}

	for address, dataRootHash := range dataTriesRootHashes {
		log.Debug("iterating data trie", "address", address, "data trie root hash", dataRootHash)

		chDataTrie, errGetAllLeaves := tr.GetAllLeavesOnChannel(dataRootHash)
		if errGetAllLeaves != nil {
			return errGetAllLeaves
		}

		for range chDataTrie {
			numDataTriesLeaves++
		}
	}

	log.Info("parsed all tries",
		"num accounts", numAccountsOnMainTrie,
		"num code nodes", numCodeNodes,
		"num data tries", len(dataTriesRootHashes),
		"num data tries leaves", numDataTriesLeaves)

	return nil
}

func getTrie(dbPath string, epoch uint32) (common.Trie, error) {
	shardCoordinator, err := sharding.NewMultiShardCoordinator(args.NumShards, args.ObservedActualShard)
	if err != nil {
		return nil, err
	}

	log.Info("getTrie()", "dbPath", dbPath, "epoch", epoch)

	localDbConfig := dbConfig // copy
	localDbConfig.FilePath = dbPath

	args := &pruning.StorerArgs{
		Identifier:                "",
		ShardCoordinator:          shardCoordinator
		CacheConf:                 cacheConfig,
		PathManager:               components.NewSimplePathManager(dbPath),
		DbPath:                    "",
		PersisterFactory:          factory.NewPersisterFactory(localDbConfig),
		Notifier:                  notifier.NewManualEpochStartNotifier(),
		OldDataCleanerProvider:    &testscommon.OldDataCleanerProviderStub{},
		CustomDatabaseRemover:     &testscommon.CustomDatabaseRemoverStub{},
		MaxBatchSize:              45000,
		NumOfEpochsToKeep:         epoch + 1,
		NumOfActivePersisters:     epoch + 1,
		StartingEpoch:             epoch,
		PruningEnabled:            true,
		EnabledDbLookupExtensions: false,
	}

	db, err := pruning.NewTriePruningStorer(args)
	if err != nil {
		return nil, err
	}

	tsm, err := trie.NewTrieStorageManagerWithoutPruning(db)
	if err != nil {
		return nil, err
	}

	tr, err := trie.NewTrie(tsm, marshaller, hasher, maxTrieLevelInMemory)
	if err != nil {
		return nil, err
	}

	log.Info("Trie loaded.")

	return tr, nil
}
