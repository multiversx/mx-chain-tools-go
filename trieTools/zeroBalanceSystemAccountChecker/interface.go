package main

import (
	"io"

	"github.com/ElrondNetwork/elrond-tools-go/trieTools/zeroBalanceSystemAccountChecker/common"
)

type crossTokenChecker interface {
	crossCheckExtraTokens(tokens map[string]struct{}) ([]string, error)
}

type tokenBalancesGetter interface {
	getBalance(address, token string) (string, error)
}

type elasticMultiSearchClient interface {
	GetMultiple(index string, requests []string) ([]byte, error)
}

type fileHandler interface {
	Open(name string) (io.Reader, error)
	ReadAll(r io.Reader) ([]byte, error)
	Getwd() (dir string, err error)
	ReadDir(dirname string) ([]common.FileInfo, error)
}
