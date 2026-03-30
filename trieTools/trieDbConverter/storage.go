package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/multiversx/mx-chain-go/common"
	config2 "github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/epochStart/notifier"
	"github.com/multiversx/mx-chain-go/storage"
	disabled3 "github.com/multiversx/mx-chain-go/storage/databaseremover/disabled"
	factory2 "github.com/multiversx/mx-chain-go/storage/factory"
	"github.com/multiversx/mx-chain-go/storage/pruning"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/pelletier/go-toml/v2"
)

type discoveredDbData struct {
	ShardID uint32
	Epoch   uint32
	DBCfg   config2.DBConfig
}

type trieDBData struct {
	ShardID          uint32
	AccountsTriePath string
	DBCfg            config2.DBConfig
}

func getStorer(
	dbData discoveredDbData,
	dbType string,
	pathManager storage.PathManagerHandler,
	snapshotStorageStats common.StateStatisticsHandler,
) (storage.Storer, error) {
	if dbType == staticType {
		return CreateStaticStorer(dbData.DBCfg)
	}

	return CreatePruningStorer(
		pathManager,
		dbData.Epoch,
		dbData.ShardID,
		snapshotStorageStats,
		dbData.DBCfg,
	)
}

func discoverTrieDbData(dbPath string, dbType string, epochDirs []string) (*trieDBData, error) {
	if dbType == staticType {
		return discoverStaticData(dbPath)
	}

	return discoverPruningData(dbPath, epochDirs)
}

func discoverStaticData(chainPath string) (*trieDBData, error) {
	staticPath := filepath.Join(chainPath, "Static")
	if !dirExists(staticPath) {
		return nil, nil
	}

	return discoverTrieDBData(staticPath)
}

func discoverPruningData(chainPath string, epochDirs []string) (*trieDBData, error) {
	if len(epochDirs) == 0 {
		return nil, nil
	}

	highestEpochPath := filepath.Join(chainPath, epochDirs[len(epochDirs)-1])

	return discoverTrieDBData(highestEpochPath)
}

func discoverEpochData(chainPath string) (uint32, []string, error) {
	entries, err := os.ReadDir(chainPath)
	if err != nil {
		return 0, nil, err
	}

	var epochs []int
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, "Epoch_") {
			continue
		}
		v, err := strconv.Atoi(strings.TrimPrefix(name, "Epoch_"))
		if err == nil && v >= 0 {
			epochs = append(epochs, v)
		}
	}
	if len(epochs) == 0 {
		return 0, nil, errors.New("no Epoch_* directories found")
	}

	sort.Ints(epochs)
	epochDirs := make([]string, 0, len(epochs))
	for _, ep := range epochs {
		epochDirs = append(epochDirs, fmt.Sprintf("Epoch_%d", ep))
	}
	highestEpoch := uint32(epochs[len(epochs)-1])

	return highestEpoch, epochDirs, nil
}

func discoverTrieDBData(root string) (*trieDBData, error) {
	shardName, shardPath, err := singleShardDir(root)
	if err != nil {
		return nil, err
	}

	accountsTriePath := filepath.Join(shardPath, "AccountsTrie")
	if err := mustBeDir(accountsTriePath); err != nil {
		return nil, err
	}

	shardID, err := parseSuffixUint32(shardName, "Shard_")
	if err != nil {
		return nil, err
	}

	dbCfg, err := loadDBConfigFromAccountsTrie(accountsTriePath)
	if err != nil {
		return nil, err
	}

	return &trieDBData{
		ShardID:          shardID,
		AccountsTriePath: accountsTriePath,
		DBCfg:            dbCfg,
	}, nil
}

func dirExists(p string) bool {
	st, err := os.Stat(p)
	if err != nil {
		return false
	}

	return st.IsDir()
}

func singleDir(root string) (name string, err error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return "", err
	}
	var dirs []os.DirEntry
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e)
		}
	}
	if len(dirs) != 1 {
		return "", fmt.Errorf("expected exactly 1 directory in %s, got %d", root, len(dirs))
	}
	name = dirs[0].Name()
	return name, nil
}

func singleShardDir(root string) (name string, fullPath string, err error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return "", "", err
	}
	var shards []os.DirEntry
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "Shard_") {
			shards = append(shards, e)
		}
	}
	if len(shards) != 1 {
		return "", "", fmt.Errorf("expected exactly 1 Shard_* directory in %s, got %d", root, len(shards))
	}
	name = shards[0].Name()
	return name, filepath.Join(root, name), nil
}

func parseSuffixUint32(value, prefix string) (uint32, error) {
	if !strings.HasPrefix(value, prefix) {
		return 0, fmt.Errorf("%q does not start with %q", value, prefix)
	}
	v, err := strconv.Atoi(strings.TrimPrefix(value, prefix))
	if err != nil || v < 0 {
		return 0, fmt.Errorf("invalid numeric value in %q", value)
	}
	return uint32(v), nil
}

func mustBeDir(p string) error {
	st, err := os.Stat(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("missing %s", p)
		}
		return fmt.Errorf("missing %s: %w", p, err)
	}
	if !st.IsDir() {
		return fmt.Errorf("%s is not a directory", p)
	}
	return nil
}

func loadDBConfigFromAccountsTrie(accountsTriePath string) (config2.DBConfig, error) {
	cfgPath := filepath.Join(accountsTriePath, "config.toml")

	raw, err := os.ReadFile(cfgPath)
	if err != nil {
		return config2.DBConfig{}, fmt.Errorf("read %s: %w", cfgPath, err)
	}

	var dbCfg config2.DBConfig
	if err := toml.Unmarshal(raw, &dbCfg); err != nil {
		return config2.DBConfig{}, fmt.Errorf("unmarshal %s: %w", cfgPath, err)
	}

	return dbCfg, nil
}

// CreateStaticStorer will create and return a static storer using the provided flags
func CreateStaticStorer(cfg config2.DBConfig) (storage.Storer, error) {
	cacheConfig := storageunit.CacheConfig{
		Type:        "SizeLRU",
		Capacity:    500000,
		SizeInBytes: 314572800, // 300MB
	}

	dbConf := storageunit.DBConfig{
		FilePath:          cfg.FilePath,
		Type:              storageunit.DBType(cfg.Type),
		BatchDelaySeconds: cfg.BatchDelaySeconds,
		MaxBatchSize:      cfg.MaxBatchSize,
		MaxOpenFiles:      cfg.MaxOpenFiles,
	}

	persisterFactory, err := factory2.NewPersisterFactory(cfg)
	if err != nil {
		return nil, err
	}

	return storageunit.NewStorageUnitFromConf(cacheConfig, dbConf, persisterFactory)
}

// CreatePruningStorer will create and return a pruning storer using the provided flags
func CreatePruningStorer(
	pathManager storage.PathManagerHandler,
	epoch uint32,
	shardId uint32,
	snapshotStorageStats common.StateStatisticsHandler,
	dbConfig config2.DBConfig,
) (storage.Storer, error) {
	persisterFactory, err := factory2.NewPersisterFactory(dbConfig)
	if err != nil {
		return nil, err
	}

	epochsData := pruning.EpochArgs{
		NumOfEpochsToKeep:     3,
		NumOfActivePersisters: 3,
		StartingEpoch:         epoch,
	}

	cacheConfig := storageunit.CacheConfig{
		Type:        "SizeLRU",
		Capacity:    500000,
		SizeInBytes: 314572800, // 300MB
	}

	shardCoordinatror := &testscommon.ShardsCoordinatorMock{
		SelfIDCalled: func() uint32 {
			return shardId
		},
	}

	pruningStorerArgs := pruning.StorerArgs{
		Identifier:                "AccountsTrie",
		ShardCoordinator:          shardCoordinatror,
		CacheConf:                 cacheConfig,
		PathManager:               pathManager,
		DbPath:                    "",
		PersisterFactory:          persisterFactory,
		Notifier:                  notifier.NewManualEpochStartNotifier(),
		OldDataCleanerProvider:    &testscommon.OldDataCleanerProviderStub{},
		CustomDatabaseRemover:     disabled3.NewDisabledCustomDatabaseRemover(),
		MaxBatchSize:              45000,
		EpochsData:                epochsData,
		PruningEnabled:            true,
		EnabledDbLookupExtensions: false,
		PersistersTracker:         pruning.NewPersistersTracker(epochsData),
		StateStatsHandler:         snapshotStorageStats,
	}

	return pruning.NewTriePruningStorer(pruningStorerArgs)
}
