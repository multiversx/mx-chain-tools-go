package trieToolsCommon

import (
	"github.com/ElrondNetwork/elrond-go-core/hashing/blake2b"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
)

var (
	// Hasher represents the internal hasher used by the node
	Hasher = blake2b.NewBlake2b()
	// Marshaller represents the internal marshaller used by the node
	Marshaller = &marshal.GogoProtoMarshalizer{}
)
