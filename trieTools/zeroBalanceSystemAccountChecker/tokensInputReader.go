package main

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ElrondNetwork/elrond-tools-go/trieTools/trieToolsCommon"
)

type FileInfo interface {
	Name() string
	IsDir() bool
}

type fileHandler interface {
	Open(name string) (io.Reader, error)
	ReadAll(r io.Reader) ([]byte, error)
	Getwd() (dir string, err error)
	ReadDir(dirname string) ([]FileInfo, error)
}

type addressTokensMapFileReader struct {
	fileHandler fileHandler
}

func newAddressTokensMapFileReader(
	fileHandler fileHandler,
) *addressTokensMapFileReader {
	return &addressTokensMapFileReader{
		fileHandler: fileHandler,
	}
}

func (atr *addressTokensMapFileReader) readInputs(tokensDir string) (trieToolsCommon.AddressTokensMap, map[uint32]trieToolsCommon.AddressTokensMap, error) {
	workingDir, err := atr.fileHandler.Getwd()
	if err != nil {
		return nil, nil, err
	}

	fullPath := filepath.Join(workingDir, tokensDir)
	contents, err := atr.fileHandler.ReadDir(fullPath)
	if err != nil {
		return nil, nil, err
	}

	globalAddressTokensMap := trieToolsCommon.NewAddressTokensMap()
	shardAddressTokensMap := make(map[uint32]trieToolsCommon.AddressTokensMap)
	for _, file := range contents {
		if file.IsDir() {
			continue
		}

		shardID, err := getShardID(file.Name())
		if err != nil {
			return nil, nil, err
		}

		addressTokensMapInCurrFile, err := atr.getFileContent(filepath.Join(fullPath, file.Name()))
		if err != nil {
			return nil, nil, err
		}

		shardAddressTokensMap[shardID] = addressTokensMapInCurrFile.ShallowClone()
		merge(globalAddressTokensMap, addressTokensMapInCurrFile)

		log.Info("read data from",
			"file", file.Name(),
			"shard", shardID,
			"num tokens in shard", shardAddressTokensMap[shardID].NumTokens(),
			"num addresses in shard", shardAddressTokensMap[shardID].NumAddresses(),
			"total num addresses in all shards", globalAddressTokensMap.NumAddresses())
	}

	return globalAddressTokensMap, shardAddressTokensMap, nil
}

func getShardID(file string) (uint32, error) {
	shardIDStr := strings.TrimPrefix(file, "shard")
	shardIDStr = strings.TrimSuffix(shardIDStr, ".json")
	shardID, err := strconv.Atoi(shardIDStr)
	if err != nil {
		return 0, fmt.Errorf("invalid file input name; expected tokens shard file name to be <shardX.json>, where X = number(e.g. shard0.json)")
	}

	return uint32(shardID), nil
}

func (atr *addressTokensMapFileReader) getFileContent(file string) (trieToolsCommon.AddressTokensMap, error) {
	jsonFile, err := atr.fileHandler.Open(file)
	if err != nil {
		return nil, err
	}

	bytesFromJson, err := atr.fileHandler.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	addressTokensMapInCurrFile := make(map[string]map[string]struct{})
	err = json.Unmarshal(bytesFromJson, &addressTokensMapInCurrFile)
	if err != nil {
		return nil, err
	}

	ret := trieToolsCommon.NewAddressTokensMap()
	for address, tokens := range addressTokensMapInCurrFile {
		tokensWithNonce := getTokensWithNonce(tokens)
		ret.Add(address, tokensWithNonce)
	}

	return ret, nil
}

func getTokensWithNonce(tokens map[string]struct{}) map[string]struct{} {
	ret := make(map[string]struct{})

	for token := range tokens {
		addTokenInMapIfHasNonce(token, ret)
	}

	return ret
}

func addTokenInMapIfHasNonce(token string, tokens map[string]struct{}) {
	if hasNonce(token) {
		tokens[token] = struct{}{}
	}
}

func hasNonce(token string) bool {
	return strings.Count(token, "-") == 2
}

func merge(dest, src trieToolsCommon.AddressTokensMap) {
	for addressSrc, tokensSrc := range src.GetMapCopy() {
		if dest.HasAddress(addressSrc) {
			log.Debug("same address found in multiple files", "address", addressSrc)
		}

		dest.Add(addressSrc, tokensSrc)
	}
}
