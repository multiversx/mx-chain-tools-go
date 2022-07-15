package export

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/state"
)

// ArgsNewExporter holds arguments for creating an exporter
type ArgsNewExporter struct {
	TrieWrapper         trieWrapper
	Format              string
	ProjectedShard      uint32
	ProjectedShardIsSet bool
	Currency            string
	CurrencyDecimals    uint
	WithContracts       bool
	WithZero            bool
}

type exporter struct {
	trie                      trieWrapper
	format                    string
	projectedShard            uint32
	projectedShardIsSet       bool
	projectedShardCoordinator sharding.Coordinator
	currency                  string
	currencyDecimals          uint
	withContracts             bool
	withZero                  bool
}

// NewExporter creates a new exporter
func NewExporter(args ArgsNewExporter) (*exporter, error) {
	projectedShardCoordinator, err := sharding.NewMultiShardCoordinator(core.MaxNumShards, args.ProjectedShard)
	if err != nil {
		return nil, err
	}

	return &exporter{
		trie:                      args.TrieWrapper,
		format:                    args.Format,
		projectedShard:            args.ProjectedShard,
		projectedShardIsSet:       args.ProjectedShardIsSet,
		projectedShardCoordinator: projectedShardCoordinator,
		currency:                  args.Currency,
		currencyDecimals:          args.CurrencyDecimals,
		withContracts:             args.WithContracts,
		withZero:                  args.WithZero,
	}, nil
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

	err = e.saveBalancesFile(block, accounts)
	if err != nil {
		return err
	}

	err = e.saveMetadataFile(block, len(accounts))
	if err != nil {
		return err
	}

	return nil
}

func (e *exporter) shouldExportAccount(account *state.UserAccountData) bool {
	isContract := core.IsSmartContractAddress(account.Address)
	if !e.withContracts && isContract {
		return false
	}

	hasZeroBalance := account.Balance.Sign() == 0
	if !e.withZero && hasZeroBalance {
		return false
	}

	hasDesiredProjectedShard := e.projectedShardCoordinator.ComputeId(account.Address) == e.projectedShardCoordinator.SelfId()
	if e.projectedShardIsSet && !hasDesiredProjectedShard {
		return false
	}

	return true
}

func (e *exporter) saveBalancesFile(block data.HeaderHandler, accounts []*state.UserAccountData) error {
	formatter, err := e.getFormatter(block)
	if err != nil {
		return err
	}

	formatterArgs := formatterArgs{
		currency:         e.currency,
		currencyDecimals: e.currencyDecimals,
	}

	text, err := formatter.toText(accounts, formatterArgs)
	if err != nil {
		return err
	}

	fileBasename := e.getOutputFileBasename(block)
	balancesFilename := fmt.Sprintf("%s.%s", fileBasename, formatter.getFileExtension())
	err = e.saveFile(balancesFilename, text)
	if err != nil {
		return err
	}

	return nil
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

func (e *exporter) getOutputFileBasename(block data.HeaderHandler) string {
	if e.projectedShardIsSet {
		return fmt.Sprintf("%s_shard_%d(%d)_epoch_%d_nonce_%d_%s",
			block.GetChainID(),
			block.GetShardID(),
			e.projectedShard,
			block.GetEpoch(),
			block.GetNonce(),
			e.currency,
		)
	}

	return fmt.Sprintf("%s_shard_%d_epoch_%d_nonce_%d_%s",
		block.GetChainID(),
		block.GetShardID(),
		block.GetEpoch(),
		block.GetNonce(),
		e.currency,
	)
}

func (e *exporter) saveMetadataFile(block data.HeaderHandler, numAccounts int) error {
	metadata := &exportMetadata{
		ChainID:             string(block.GetChainID()),
		ActualShardID:       block.GetShardID(),
		ProjectedShardID:    e.projectedShard,
		ProjectedShardIsSet: e.projectedShardIsSet,
		Epoch:               block.GetEpoch(),
		BlockNonce:          block.GetNonce(),
		BlockRootHash:       hex.EncodeToString(block.GetRootHash()),
		Format:              e.format,
		Currency:            e.currency,
		CurrencyDecimals:    e.currencyDecimals,
		WithContracts:       e.withContracts,
		WithZero:            e.withZero,
		NumAccounts:         numAccounts,
	}

	metadataJson, err := json.MarshalIndent(metadata, "", fourSpaces)
	if err != nil {
		return err
	}

	fileBasename := e.getOutputFileBasename(block)
	metadataFilename := fmt.Sprintf("%s.%s.metadata.json", fileBasename, e.format)
	err = e.saveFile(metadataFilename, string(metadataJson))
	if err != nil {
		return err
	}

	return nil
}

func (e *exporter) saveFile(filename string, text string) error {
	err := ioutil.WriteFile(filename, []byte(text), core.FileModeReadWrite)
	if err != nil {
		return err
	}

	log.Info("Saved file:", "file", filename)
	return nil
}
