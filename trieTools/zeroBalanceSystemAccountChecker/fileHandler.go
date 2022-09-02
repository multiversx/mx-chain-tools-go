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

// Open opens the named file for reading
func (fh *osFileHandler) Open(name string) (io.Reader, error) {
	return os.Open(name)
}

// ReadAll reads from r until an error or EOF and returns the data it read.
func (fh *osFileHandler) ReadAll(r io.Reader) ([]byte, error) {
	return ioutil.ReadAll(r)
}

// Getwd returns a rooted path name corresponding to the current directory
func (fh *osFileHandler) Getwd() (dir string, err error) {
	return os.Getwd()
}

// ReadDir reads the directory and returns no directory entries along with the error.
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
