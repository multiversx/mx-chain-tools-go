package config

import "github.com/ElrondNetwork/elrond-tools-go/trieTools/trieToolsCommon"

// ContextFlagsMetaDataRemover is the flags config for meta data remover
type ContextFlagsMetaDataRemover struct {
	trieToolsCommon.ContextFlagsConfig
	Outfile string
	Tokens  string
	Pems    string
}

// Config holds the config for meta data remover tool
type Config struct {
	ProxyUrl                     string `toml:"ProxyUrl"`
	TokensToDeletePerTransaction uint64 `toml:"TokensToDeletePerTransaction"`
	GasLimit                     uint64 `toml:"GasLimit"`
}
