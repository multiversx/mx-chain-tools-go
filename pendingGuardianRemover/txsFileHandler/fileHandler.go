package txsFileHandler

import (
	"encoding/json"
	"io/fs"
	"io/ioutil"
	"sort"

	"github.com/multiversx/mx-sdk-go/data"
)

type fileHandler struct {
	file string
}

// NewFileHandler returns a new fileHandler
func NewFileHandler(file string) (*fileHandler, error) {

	return &fileHandler{
		file: file,
	}, nil
}

// Save saves the provided txs to a json file
func (handler *fileHandler) Save(txs []*data.Transaction) error {
	txsMap := make(map[uint64]*data.Transaction, len(txs))

	sort.Slice(txs, func(i, j int) bool {
		return txs[i].Nonce > txs[j].Nonce
	})

	for i := 0; i < len(txs); i++ {
		txsMap[txs[i].Nonce] = txs[i]
	}

	jsonBytes, err := json.Marshal(txsMap)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(handler.file, jsonBytes, fs.ModePerm)
}

// Load loads the transactions from the managed json file
func (handler *fileHandler) Load() (map[uint64]*data.Transaction, error) {
	txsBuff, err := ioutil.ReadFile(handler.file)
	if err != nil {
		return nil, err
	}

	txsMap := make(map[uint64]*data.Transaction)
	err = json.Unmarshal(txsBuff, &txsMap)
	if err != nil {
		return nil, err
	}

	return txsMap, nil

}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *fileHandler) IsInterfaceNil() bool {
	return handler == nil
}
