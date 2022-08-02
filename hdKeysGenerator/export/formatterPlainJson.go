package export

import (
	"encoding/hex"
	"encoding/json"

	"github.com/ElrondNetwork/elrond-tools-go/hdKeysGenerator/common"
)

type plainExportedKey struct {
	Index     int
	SecretKey string `json:"secretKey"`
	PublicKey string `json:"publicKey"`
	Address   string `json:"address"`
}

type formatterPlainJson struct {
}

func (f *formatterPlainJson) toText(keys []common.GeneratedKey, args formatterArgs) (string, error) {
	records := make([]plainExportedKey, 0, len(keys))

	for _, key := range keys {
		records = append(records, plainExportedKey{
			Index:     key.Index,
			Address:   key.Address,
			SecretKey: hex.EncodeToString(key.SecretKey),
			PublicKey: hex.EncodeToString(key.PublicKey),
		})
	}

	recordsJson, err := json.MarshalIndent(records, "", fourSpaces)
	if err != nil {
		return "", err
	}

	return string(recordsJson), nil
}

func (f *formatterPlainJson) getFileExtension() string {
	return "json"
}
