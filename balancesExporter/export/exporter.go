package export

import (
	"fmt"
	"io/ioutil"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/state"
)

// ArgsNewExporter holds arguments for creating an exporter
type ArgsNewExporter struct {
	ShardCoordinator sharding.Coordinator
	TrieWrapper      trieWrapper
	Format           string
	Currency         string
	CurrencyDecimals uint
	WithContracts    bool
	WithZero         bool
}

type exporter struct {
	shardCoordinator sharding.Coordinator
	trie             trieWrapper
	format           string
	currency         string
	currencyDecimals uint
	withContracts    bool
	withZero         bool
}

// NewExporter creates a new exporter
func NewExporter(args ArgsNewExporter) *exporter {
	return &exporter{
		shardCoordinator: args.ShardCoordinator,
		trie:             args.TrieWrapper,
		format:           args.Format,
		currency:         args.Currency,
		currencyDecimals: args.CurrencyDecimals,
		withContracts:    args.WithContracts,
		withZero:         args.WithZero,
	}
}

// ExportBalancesAtBlock exports balances of accounts at a given block
func (e *exporter) ExportBalancesAtBlock(block data.HeaderHandler) error {
	rootHash := block.GetRootHash()

	accounts, err := e.trie.GetUserAccounts(rootHash, e.shouldExportAccount)
	if err != nil {
		return err
	}

	log.Info("Exporting:",
		"numAccounts", len(accounts),
		"blockNonce", block.GetNonce(),
		"blockRootHash", block.GetRootHash(),
		"formatType", e.format,
	)

	formatter, err := e.getFormatter(block)
	if err != nil {
		return err
	}

	formatterArgs := formatterArgs{
		currency:         e.currency,
		currencyDecimals: e.currencyDecimals,
	}

	filename := formatter.getFileName(block, formatterArgs)
	text, err := formatter.toText(accounts, formatterArgs)
	if err != nil {
		return err
	}

	err = e.saveFile(filename, text)
	if err != nil {
		return err
	}

	return nil
}

func (e *exporter) shouldExportAccount(account *state.UserAccountData) bool {
	if !e.withContracts && core.IsSmartContractAddress(account.Address) {
		return false
	}
	if !e.withZero && account.Balance.Sign() == 0 {
		return false
	}
	if e.shardCoordinator.ComputeId(account.Address) != e.shardCoordinator.SelfId() {
		return false
	}

	return true
}

func (e *exporter) getFormatter(block data.HeaderHandler) (formatter, error) {
	switch e.format {
	case FormatterNamePlainText:
		return &formatterPlainText{}, nil
	case FormatterNamePlainJson:
		return &formatterPlainJson{}, nil
	case FormatterNameRosettaJson:
		return &formatterRosettaJson{}, nil
	}

	return nil, fmt.Errorf("unknown format: %s", e.format)
}

func (e *exporter) saveFile(filename string, text string) error {
	err := ioutil.WriteFile(filename, []byte(text), core.FileModeReadWrite)
	if err != nil {
		return err
	}

	log.Info("Saved file:", "file", filename)
	return nil
}
