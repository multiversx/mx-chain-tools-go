package config

import (
	"math/big"

	"github.com/multiversx/mx-chain-tools-go/trieTools/generic/filter"
	"github.com/multiversx/mx-chain-tools-go/trieTools/trieToolsCommon"
)

// ContextFlagsTokensExporter is the flags config for tokens exporter
type ContextFlagsGeneric struct {
	trieToolsCommon.ContextFlagsConfig
	Limit   uint64
	Filters []filter.Operation
	Outfile string
}

type AccountDetails struct {
	Nonce   uint64   `json:"nonce"`
	Address string   `json:"address"`
	Balance *big.Int `json:"balance"`
}
