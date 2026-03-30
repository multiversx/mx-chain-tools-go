package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestDiscoverDBLayoutStaticExtractsChainShardEpochsAndConfig(t *testing.T) {
	root := t.TempDir()

	writeTrieDBFixture(t, root, "chainID", "Static", "Shard_2", dbConfigFixture{
		Type:              "LvlDBSerial",
		BatchDelaySeconds: 3,
		MaxBatchSize:      111,
		MaxOpenFiles:      17,
	})
	writeTrieDBFixture(t, root, "chainID", "Epoch_3", "Shard_2", dbConfigFixture{Type: "LvlDBSerial", MaxBatchSize: 203})
	writeTrieDBFixture(t, root, "chainID", "Epoch_7", "Shard_2", dbConfigFixture{Type: "LvlDB", MaxBatchSize: 207})
	createDir(t, filepath.Join(root, "chainID", "Epoch_invalid"))

	discovered, err := discoverDBLayout(root, staticType)
	if err != nil {
		t.Fatalf("discoverDBLayout returned error: %v", err)
	}

	if discovered.ChainID != "chainID" {
		t.Fatalf("unexpected chain ID: got %q", discovered.ChainID)
	}
	if discovered.ShardID != 2 {
		t.Fatalf("unexpected shard ID: got %d", discovered.ShardID)
	}
	if discovered.Epoch == nil || *discovered.Epoch != 7 {
		t.Fatalf("unexpected highest epoch: got %v", discovered.Epoch)
	}

	expectedEpochDirs := []string{"Epoch_3", "Epoch_7"}
	if !reflect.DeepEqual(discovered.EpochDirs, expectedEpochDirs) {
		t.Fatalf("unexpected epoch dirs: got %#v, want %#v", discovered.EpochDirs, expectedEpochDirs)
	}

	expectedPath := filepath.Join(root, "chainID", "Static", "Shard_2", "AccountsTrie")
	if discovered.AccountsTriePath != expectedPath {
		t.Fatalf("unexpected AccountsTrie path: got %q, want %q", discovered.AccountsTriePath, expectedPath)
	}
	if discovered.DBCfg.Type != "LvlDBSerial" || discovered.DBCfg.BatchDelaySeconds != 3 || discovered.DBCfg.MaxBatchSize != 111 || discovered.DBCfg.MaxOpenFiles != 17 {
		t.Fatalf("unexpected config: %#v", discovered.DBCfg)
	}
}

func TestDiscoverDBLayoutPruningExtractsHighestEpochShardAndConfig(t *testing.T) {
	root := t.TempDir()

	writeTrieDBFixture(t, root, "chainID", "Static", "Shard_5", dbConfigFixture{Type: "LvlDBSerial", MaxBatchSize: 100})
	writeTrieDBFixture(t, root, "chainID", "Epoch_4", "Shard_5", dbConfigFixture{Type: "LvlDB", BatchDelaySeconds: 4, MaxBatchSize: 204, MaxOpenFiles: 24})
	writeTrieDBFixture(t, root, "chainID", "Epoch_12", "Shard_5", dbConfigFixture{Type: "LvlDB", BatchDelaySeconds: 12, MaxBatchSize: 212, MaxOpenFiles: 32})

	discovered, err := discoverDBLayout(root, pruningDbType)
	if err != nil {
		t.Fatalf("discoverDBLayout returned error: %v", err)
	}

	if discovered.ChainID != "chainID" {
		t.Fatalf("unexpected chain ID: got %q", discovered.ChainID)
	}
	if discovered.ShardID != 5 {
		t.Fatalf("unexpected shard ID: got %d", discovered.ShardID)
	}
	if discovered.Epoch == nil || *discovered.Epoch != 12 {
		t.Fatalf("unexpected highest epoch: got %v", discovered.Epoch)
	}

	expectedEpochDirs := []string{"Epoch_4", "Epoch_12"}
	if !reflect.DeepEqual(discovered.EpochDirs, expectedEpochDirs) {
		t.Fatalf("unexpected epoch dirs: got %#v, want %#v", discovered.EpochDirs, expectedEpochDirs)
	}

	expectedPath := filepath.Join(root, "chainID", "Epoch_12", "Shard_5", "AccountsTrie")
	if discovered.AccountsTriePath != expectedPath {
		t.Fatalf("unexpected AccountsTrie path: got %q, want %q", discovered.AccountsTriePath, expectedPath)
	}
	if discovered.DBCfg.BatchDelaySeconds != 12 || discovered.DBCfg.MaxBatchSize != 212 || discovered.DBCfg.MaxOpenFiles != 32 {
		t.Fatalf("unexpected config: %#v", discovered.DBCfg)
	}
}

func TestDiscoverDBLayoutStaticWithoutEpochsReturnsEmptyEpochMetadata(t *testing.T) {
	root := t.TempDir()
	writeTrieDBFixture(t, root, "chainID", "Static", "Shard_0", dbConfigFixture{Type: "LvlDBSerial"})

	discovered, err := discoverDBLayout(root, staticType)
	if err != nil {
		t.Fatalf("discoverDBLayout returned error: %v", err)
	}

	if discovered.Epoch != nil {
		t.Fatalf("expected nil epoch, got %v", *discovered.Epoch)
	}
	if len(discovered.EpochDirs) != 0 {
		t.Fatalf("expected no epoch dirs, got %#v", discovered.EpochDirs)
	}
}

func TestDiscoverDBLayoutRejectsMismatchedShardIDsAcrossLayouts(t *testing.T) {
	root := t.TempDir()
	writeTrieDBFixture(t, root, "chainID", "Static", "Shard_1", dbConfigFixture{Type: "LvlDBSerial"})
	writeTrieDBFixture(t, root, "chainID", "Epoch_9", "Shard_2", dbConfigFixture{Type: "LvlDB"})

	_, err := discoverDBLayout(root, staticType)
	if err == nil {
		t.Fatal("expected shard mismatch error")
	}
	if !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("unexpected error: %v", err)
	}
}

type dbConfigFixture struct {
	Type              string
	BatchDelaySeconds int
	MaxBatchSize      int
	MaxOpenFiles      int
}

func writeTrieDBFixture(t *testing.T, root string, chainID string, storageDir string, shardDir string, cfg dbConfigFixture) {
	t.Helper()

	accountsTriePath := filepath.Join(root, chainID, storageDir, shardDir, "AccountsTrie")
	createDir(t, accountsTriePath)

	configToml := strings.Join([]string{
		"Type = \"" + cfg.Type + "\"",
		"BatchDelaySeconds = " + strconv.Itoa(cfg.BatchDelaySeconds),
		"MaxBatchSize = " + strconv.Itoa(cfg.MaxBatchSize),
		"MaxOpenFiles = " + strconv.Itoa(cfg.MaxOpenFiles),
	}, "\n") + "\n"

	configPath := filepath.Join(accountsTriePath, "config.toml")
	if err := os.WriteFile(configPath, []byte(configToml), 0o644); err != nil {
		t.Fatalf("WriteFile failed for %q: %v", configPath, err)
	}
}

func createDir(t *testing.T, dirPath string) {
	t.Helper()

	if err := os.MkdirAll(dirPath, 0o755); err != nil {
		t.Fatalf("MkdirAll failed for %q: %v", dirPath, err)
	}
}
