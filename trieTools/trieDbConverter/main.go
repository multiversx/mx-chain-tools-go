package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"strconv"

	"github.com/multiversx/mx-chain-go/common/statistics"
	nodeConfig "github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/storage"
	factory2 "github.com/multiversx/mx-chain-go/storage/factory"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-tools-go/trieTools/trieToolsCommon"
	"github.com/urfave/cli"
)

var log = logger.GetOrCreate("trieDbConverter")

func main() {
	app := cli.NewApp()
	app.Name = "Trie stats CLI app"
	app.Usage = "This is the entry point for the tool that converts between a Static trie db and a pruning storer"
	app.Flags = getFlags()
	app.Authors = []cli.Author{
		{
			Name:  "The MultiversX Team",
			Email: "contact@multiversx.com",
		},
	}

	app.Action = startProcess

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
		return
	}

	log.Info("execution finished successfully")
}

func startProcess(c *cli.Context) error {
	flagsConfig := getFlagsConfig(c)

	fileLogger, err := createProcessFileLogger(flagsConfig.ContextFlagsConfig)
	if err != nil {
		return err
	}
	if fileLogger != nil && !fileLogger.IsInterfaceNil() {
		defer closeClosable(fileLogger)
	}

	log.Info("sanity checks...")

	err = logger.SetLogLevel(flagsConfig.LogLevel)
	if err != nil {
		return err
	}

	//flagsConfig.WorkingDir = ""
	//static - flagsConfig.HexRootHash = "ff9d2f40f7d9ac7cc0ada8ea59d487810817d54023aa3892c10b8f17c0908ae3"
	//pruning - flagsConfig.HexRootHash = "8ed46937ea20a176dc3e9c64dbc1f3880659b074eca0c4300b6b5bb222c88155"
	//flagsConfig.OriginDbType = pruningDbType
	//flagsConfig.TargetDbType = staticType

	rootHash, err := hex.DecodeString(flagsConfig.HexRootHash)
	if err != nil {
		return fmt.Errorf("%w when decoding the provided hex root hash", err)
	}
	if len(rootHash) != rootHashLength {
		return fmt.Errorf("wrong root hash length: expected %d, got %d", rootHashLength, len(rootHash))
	}

	log.Info("starting processing state", "pid", os.Getpid())

	return convertDb(flagsConfig, rootHash)
}

func convertDb(flagsConfig ContextFlagsDbConverter, rootHash []byte) error {
	dbDir := "db"
	chainID, err := singleDir(dbDir)
	if err != nil {
		return fmt.Errorf("discover chainID in %s: %w", dbDir, err)
	}

	epochsPath := path.Join(dbDir, chainID)
	epoch, epochDirs, err := discoverEpochData(epochsPath)
	if err != nil {
		return err
	}

	trieDbData, err := discoverTrieDbData(epochsPath, flagsConfig.OriginDbType, epochDirs)
	if err != nil {
		return err
	}

	log.Info("discovered trie db data",
		"chainID", chainID,
		"shardID", trieDbData.ShardID,
		"lastEpoch", epoch,
		"epochDirs", epochDirs,
		"accountsTriePath", trieDbData.AccountsTriePath,
		"dbConfig", trieDbData.DBCfg,
	)

	pathManager, err := factory2.CreatePathManager(factory2.ArgCreatePathManager{
		WorkingDir: "",
		ChainID:    chainID,
	})
	if err != nil {
		return err
	}

	dbData := getDbData(trieDbData.ShardID, epoch, trieDbData, flagsConfig.OriginDbType, pathManager)
	originDb, err := getStorer(dbData, flagsConfig.OriginDbType, pathManager, statistics.NewStateStatistics())
	if err != nil {
		return err
	}
	defer closeStorer(originDb)

	dbData = getDbData(trieDbData.ShardID, epoch, trieDbData, flagsConfig.TargetDbType, pathManager)
	targetDb, err := getStorer(dbData, flagsConfig.TargetDbType, pathManager, statistics.NewStateStatistics())
	if err != nil {
		return err
	}
	defer closeStorer(targetDb)

	return copyTrieAndAllDataTries(originDb, targetDb, rootHash, flagsConfig.MaxGoroutines)
}

func getDbData(
	shardID uint32,
	epoch uint32,
	data *trieDBData,
	dbType string,
	pathManager storage.PathManagerHandler,
) discoveredDbData {
	if dbType == staticType {
		pathForStatic := pathManager.PathForStatic(strconv.Itoa(int(shardID)), "AccountsTrie")
		return discoveredDbData{
			ShardID: shardID,
			Epoch:   epoch,
			DBCfg: nodeConfig.DBConfig{
				FilePath:            pathForStatic,
				Type:                data.DBCfg.Type,
				BatchDelaySeconds:   data.DBCfg.BatchDelaySeconds,
				MaxBatchSize:        data.DBCfg.MaxBatchSize,
				MaxOpenFiles:        data.DBCfg.MaxOpenFiles,
				UseTmpAsFilePath:    data.DBCfg.UseTmpAsFilePath,
				ShardIDProviderType: data.DBCfg.ShardIDProviderType,
				NumShards:           data.DBCfg.NumShards,
			},
		}
	}
	return discoveredDbData{
		ShardID: shardID,
		Epoch:   epoch,
		DBCfg: nodeConfig.DBConfig{
			FilePath:            "",
			Type:                data.DBCfg.Type,
			BatchDelaySeconds:   data.DBCfg.BatchDelaySeconds,
			MaxBatchSize:        data.DBCfg.MaxBatchSize,
			MaxOpenFiles:        data.DBCfg.MaxOpenFiles,
			UseTmpAsFilePath:    data.DBCfg.UseTmpAsFilePath,
			ShardIDProviderType: data.DBCfg.ShardIDProviderType,
			NumShards:           data.DBCfg.NumShards,
		},
	}
}

type closable interface {
	Close() error
	IsInterfaceNil() bool
}

func closeClosable(resource closable) {
	if resource == nil || resource.IsInterfaceNil() {
		return
	}

	log.LogIfError(resource.Close())
}

func closeStorer(storer storage.Storer) {
	closeClosable(storer)
}

func createProcessFileLogger(flagsConfig trieToolsCommon.ContextFlagsConfig) (closable, error) {
	fileLogger, err := trieToolsCommon.AttachFileLogger(log, "dbConverter", flagsConfig)
	if err != nil {
		return nil, err
	}

	if fileLogger == nil || fileLogger.IsInterfaceNil() {
		return nil, nil
	}

	return fileLogger, nil
}
