package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func readPemsData(pemsFile string, pemDataProvider pemProvider) (map[uint32]*pkAddress, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	fullPath := filepath.Join(workingDir, pemsFile)
	contents, err := ioutil.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	shardPemDataMap := make(map[uint32]*pkAddress)
	for _, file := range contents {
		if file.IsDir() {
			continue
		}

		shardID, err := getShardID(file.Name())
		if err != nil {
			return nil, err
		}

		pemData, err := pemDataProvider.getPrivateKeyAndAddress(filepath.Join(fullPath, file.Name()))
		if err != nil {
			return nil, err
		}

		shardPemDataMap[shardID] = pemData
	}

	return shardPemDataMap, nil
}

func getShardID(file string) (uint32, error) {
	shardIDStr := strings.TrimPrefix(file, "shard")
	shardIDStr = strings.TrimSuffix(shardIDStr, ".pem")
	shardID, err := strconv.Atoi(shardIDStr)
	if err != nil {
		return 0, fmt.Errorf("invalid file input name = %s; expected pem file name to be <shardX.pem>, where X = number(e.g. shard0.pem)", file)
	}

	return uint32(shardID), nil
}
