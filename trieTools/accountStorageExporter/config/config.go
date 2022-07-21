package config

import "github.com/ElrondNetwork/elrond-tools-go/common/trieToolsCommon"

// ContextFlagsConfigAddr the configuration for flags
type ContextFlagsConfigAddr struct {
	trieToolsCommon.ContextFlagsConfig
	Address string
}
