package export

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ElrondNetwork/elrond-tools-go/hdKeysGenerator/common"
)

type formatterPlainText struct {
}

func (f *formatterPlainText) toText(keys []common.GeneratedKey) (string, error) {
	var builder strings.Builder

	header := "AccountIndex\tAddressIndex\tAddress\tPublicKey\tSecretKey\n"
	_, err := builder.WriteString(header)
	if err != nil {
		return "", err
	}

	for _, key := range keys {
		line := fmt.Sprintf("%d\t%d\t%s\t%s\t%s\n",
			key.AccountIndex,
			key.AddressIndex,
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
