package main

import (
	"io"
	"io/ioutil"
	"os"

	"github.com/ElrondNetwork/elrond-tools-go/trieTools/zeroBalanceSystemAccountChecker/common"
)

type osFileHandler struct {
}

func newOSFileHandler() *osFileHandler {
	return &osFileHandler{}
}

func (fh *osFileHandler) Open(name string) (io.Reader, error) {
	return os.Open(name)
}

func (fh *osFileHandler) ReadAll(r io.Reader) ([]byte, error) {
	return ioutil.ReadAll(r)
}

func (fh *osFileHandler) Getwd() (dir string, err error) {
	return os.Getwd()
}

func (fh *osFileHandler) ReadDir(dirname string) ([]common.FileInfo, error) {
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		return nil, err
	}

	ret := make([]common.FileInfo, 0, len(files))
	for _, f := range files {
		ret = append(ret, f)
	}

	return ret, nil
}
