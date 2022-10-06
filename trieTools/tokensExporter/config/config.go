package config

import "github.com/ElrondNetwork/elrond-tools-go/trieTools/trieToolsCommon"

// ContextFlagsTokensExporter is the flags config for tokens exporter
type ContextFlagsTokensExporter struct {
	trieToolsCommon.ContextFlagsConfig
	Outfile string
	Token   string
}
