package config

import "github.com/ElrondNetwork/elrond-tools-go/trieTools/trieToolsCommon"

// ContextFlagsMetaDataRemover is the flags config for meta data remover
type ContextFlagsMetaDataRemover struct {
	trieToolsCommon.ContextFlagsConfig
	Outfile string
	Tokens  string
}
