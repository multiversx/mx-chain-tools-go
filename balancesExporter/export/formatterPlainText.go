package export

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/state"
)

type plainBalance struct {
	Address string `json:"address"`
	Balance string `json:"balance"`
}

type formatterPlainText struct {
}

func (f *formatterPlainText) toText(accounts []*state.UserAccountData, args formatterArgs) (string, error) {
	var builder strings.Builder

	for _, account := range accounts {
		address := addressConverter.Encode(account.Address)
		balance := account.Balance.String()
		line := fmt.Sprintf("%s %s %s\n", address, balance, args.currency)
		_, err := builder.WriteString(line)
		if err != nil {
			return "", err
		}
	}

	return builder.String(), nil
}

func (f *formatterPlainText) getFileName(block data.HeaderHandler, args formatterArgs) string {
	return fmt.Sprintf("%s_shard_%d_epoch_%d_nonce_%d_roothash_%s_%s.txt",
		block.GetChainID(),
		block.GetShardID(),
		block.GetEpoch(),
		block.GetNonce(),
		hex.EncodeToString(block.GetRootHash()),
		args.currency,
	)
}
