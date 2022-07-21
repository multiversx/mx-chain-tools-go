package config

import "github.com/ElrondNetwork/elrond-tools-go/trieTools/trieToolsCommon"

// ContextFlagsConfigAddr the configuration for flags
type ContextFlagsConfigAddr struct {
	trieToolsCommon.ContextFlagsConfig
	Address string
}
