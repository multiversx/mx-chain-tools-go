package export

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core/pubkeyConverter"
	"github.com/ElrondNetwork/elrond-go-core/data"
)

const (
	addressLength = 32
)

var (
	addressConverter, _ = pubkeyConverter.NewBech32PubkeyConverter(addressLength, log)
)

type ArgsNewExporter struct {
	TrieWrapper      trieWrapper
	Currency         string
	CurrencyDecimals uint32
}

type exporter struct {
	trie             trieWrapper
	currency         string
	currencyDecimals uint32
}

func NewExporter(args ArgsNewExporter) *exporter {
	return &exporter{
		trie:             args.TrieWrapper,
		currency:         args.Currency,
		currencyDecimals: args.CurrencyDecimals,
	}
}

func (e *exporter) ExportBalancesAfterBlock(block data.HeaderHandler) error {
	rootHash := block.GetRootHash()

	users, err := e.trie.GetAllUserAccounts(rootHash)
	if err != nil {
		return err
	}

	// TODO: flag include zero
	// TODO: flag include contracts

	fmt.Println(len(users))
	return nil
}
