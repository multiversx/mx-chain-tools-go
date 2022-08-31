package main

import (
	"encoding/json"
	"io"
	"testing"

	"github.com/ElrondNetwork/elrond-tools-go/trieTools/trieToolsCommon"
	"github.com/stretchr/testify/require"
)

type FileHandlerStub struct {
	OpenCalled    func(name string) (io.Reader, error)
	ReadAllCalled func(r io.Reader) ([]byte, error)
	GetwdCalled   func() (dir string, err error)
	ReadDirCalled func(dirname string) ([]FileInfo, error)
}

func (fhs *FileHandlerStub) Open(name string) (io.Reader, error) {
	if fhs.OpenCalled != nil {
		return fhs.OpenCalled(name)
	}

	return nil, nil
}

func (fhs *FileHandlerStub) ReadAll(r io.Reader) ([]byte, error) {
	if fhs.ReadAllCalled != nil {
		return fhs.ReadAllCalled(r)
	}

	return nil, nil
}

func (fhs *FileHandlerStub) Getwd() (dir string, err error) {
	if fhs.GetwdCalled != nil {
		return fhs.GetwdCalled()
	}

	return "", nil
}

func (fhs *FileHandlerStub) ReadDir(dirname string) ([]FileInfo, error) {
	if fhs.ReadDirCalled != nil {
		return fhs.ReadDirCalled(dirname)
	}

	return nil, nil
}

type FileStub struct {
	NameCalled  func() string
	IsDirCalled func() bool
}

func (fs *FileStub) Name() string {
	if fs.NameCalled != nil {
		return fs.NameCalled()
	}

	return ""
}

func (fs *FileStub) IsDir() bool {
	if fs.IsDirCalled != nil {
		return fs.IsDirCalled()
	}

	return false
}

type ReaderStub struct {
	ReadCalled func(p []byte) (n int, err error)
}

func (rs *ReaderStub) Read(p []byte) (n int, err error) {
	if rs.ReadCalled != nil {
		return rs.ReadCalled(p)
	}

	return 0, nil
}

func TestReadInputs(t *testing.T) {
	workingDir := "working-dir"
	tokensDir := "tokens-dir"

	file1Name := "shard0"
	file2Name := "shard1"

	file1 := &FileStub{
		NameCalled: func() string {
			return file1Name
		},
	}
	file2 := &FileStub{
		NameCalled: func() string {
			return file2Name
		},
	}

	adr1 := "adr1"
	sysAccAddr := "sysAccAddr"

	adr1Tokens := map[string]struct{}{
		"token1-r-0": {},
		"token2-r-0": {},
	}
	sysAccTokensShard0 := map[string]struct{}{
		"token3-r-0": {},
		"token3-r-1": {},
	}
	addressTokensMapShard0 := trieToolsCommon.NewAddressTokensMap()
	addressTokensMapShard0.Add(adr1, adr1Tokens)
	addressTokensMapShard0.Add(sysAccAddr, sysAccTokensShard0)

	adr2 := "adr2"
	adr2Tokens := map[string]struct{}{
		"token3-r-0": {},
		"token4-r-0": {},
	}
	sysAccTokensShard1 := map[string]struct{}{
		"token3-r-0": {},
		"token3-r-2": {},
	}
	addressTokensMapShard1 := trieToolsCommon.NewAddressTokensMap()
	addressTokensMapShard1.Add(adr2, adr2Tokens)
	addressTokensMapShard1.Add(sysAccAddr, sysAccTokensShard1)

	openCt := 0
	fileHandlerStub := &FileHandlerStub{
		GetwdCalled: func() (dir string, err error) {
			return workingDir, nil
		},
		ReadDirCalled: func(dirname string) ([]FileInfo, error) {
			require.Equal(t, workingDir+"/"+tokensDir, dirname)
			return []FileInfo{file1, file2}, nil
		},

		OpenCalled: func(name string) (io.Reader, error) {
			openCt++

			switch openCt {
			case 1:
				require.Equal(t, workingDir+"/"+tokensDir+"/"+file1Name, name)
			case 2:
				require.Equal(t, workingDir+"/"+tokensDir+"/"+file2Name, name)
			}

			return nil, nil
		},
		ReadAllCalled: func(r io.Reader) ([]byte, error) {
			switch openCt {
			case 1:
				return json.Marshal(addressTokensMapShard0.GetMapCopy())
			case 2:
				return json.Marshal(addressTokensMapShard1.GetMapCopy())
			}

			return nil, nil
		},
	}

	reader := newAddressTokensMapFileReader(fileHandlerStub)
	globalTokens, shardTokens, err := reader.readInputs(tokensDir)
	require.Nil(t, err)

	expectedGlobalTokensMap := trieToolsCommon.NewAddressTokensMap()
	expectedGlobalTokensMap.Add(adr1, adr1Tokens)
	expectedGlobalTokensMap.Add(adr2, adr2Tokens)
	expectedGlobalTokensMap.Add(sysAccAddr, map[string]struct{}{
		"token3-r-0": {},
		"token3-r-1": {},
		"token3-r-2": {},
	})
	require.Equal(t, expectedGlobalTokensMap, globalTokens)

	expectedShardTokens := make(map[uint32]trieToolsCommon.AddressTokensMap)
	expectedShardTokens[0] = addressTokensMapShard0
	expectedShardTokens[1] = addressTokensMapShard1
	require.Equal(t, expectedShardTokens, shardTokens)
}
