package export

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/state"
)

type rosettaBalance struct {
	AccountIdentifier rosettaAccountIdentifier `json:"account_identifier"`
	Currency          *rosettaCurrency         `json:"currency"`
	Value             string                   `json:"value"`
}

type rosettaAccountIdentifier struct {
	Address string `json:"address"`
}

type rosettaCurrency struct {
	Symbol   string `json:"symbol"`
	Decimals uint   `json:"decimals"`
}

type formatterRosettaJson struct {
}

func (f *formatterRosettaJson) toText(accounts []*state.UserAccountData, args formatterArgs) (string, error) {
	records := make([]rosettaBalance, 0, len(accounts))

	currency := &rosettaCurrency{
		Symbol:   args.currency,
		Decimals: args.currencyDecimals,
	}

	for _, account := range accounts {
		address := addressConverter.Encode(account.Address)
		balance := account.Balance.String()

		records = append(records, rosettaBalance{
			AccountIdentifier: rosettaAccountIdentifier{
				Address: address,
			},
			Currency: currency,
			Value:    balance,
		})
	}

	recordsJson, err := json.MarshalIndent(records, "", FourSpaces)
	if err != nil {
		return "", err
	}

	return string(recordsJson), nil
}

func (f *formatterRosettaJson) getFileName(block data.HeaderHandler, args formatterArgs) string {
	return fmt.Sprintf("%s_shard_%d_epoch_%d_nonce_%d_roothash_%s_%s.rosetta.json",
		block.GetChainID(),
		block.GetShardID(),
		block.GetEpoch(),
		block.GetNonce(),
		hex.EncodeToString(block.GetRootHash()),
		args.currency,
	)
}
