package export

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/state"
)

type formatterPlainJson struct {
}

func (f *formatterPlainJson) toText(accounts []*state.UserAccountData, args formatterArgs) (string, error) {
	records := make([]plainBalance, 0, len(accounts))

	for _, account := range accounts {
		address := addressConverter.Encode(account.Address)
		balance := account.Balance.String()

		records = append(records, plainBalance{
			Address: address,
			Balance: balance,
		})
	}

	recordsJson, err := json.MarshalIndent(records, "", FourSpaces)
	if err != nil {
		return "", err
	}

	return string(recordsJson), nil
}

func (f *formatterPlainJson) getFileName(block data.HeaderHandler, args formatterArgs) string {
	return fmt.Sprintf("%s_shard_%d_epoch_%d_nonce_%d_roothash_%s_%s.json",
		block.GetChainID(),
		block.GetShardID(),
		block.GetEpoch(),
		block.GetNonce(),
		hex.EncodeToString(block.GetRootHash()),
		args.currency,
	)
}
