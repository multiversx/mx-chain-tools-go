package export

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ElrondNetwork/elrond-tools-go/hdKeysGenerator/common"
)

type formatterPlainText struct {
}

func (f *formatterPlainText) toText(keys []common.GeneratedKey, args formatterArgs) (string, error) {
	var builder strings.Builder

	header := "Index\tAddress\tPublicKey\tSecretKey"
	_, err := builder.WriteString(header)
	if err != nil {
		return "", err
	}

	for _, key := range keys {
		line := fmt.Sprintf("%d\t%s\t%s\t%s\n",
			key.Index,
			key.Address,
			hex.EncodeToString(key.PublicKey),
			hex.EncodeToString(key.SecretKey),
		)
		_, err := builder.WriteString(line)
		if err != nil {
			return "", err
		}
	}

	return builder.String(), nil
}

func (f *formatterPlainText) getFileExtension() string {
	return "txt"
}
