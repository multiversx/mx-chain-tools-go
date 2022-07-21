package trieToolsCommon

import (
	"github.com/ElrondNetwork/elrond-go-core/hashing/blake2b"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
)

var (
	Hasher     = blake2b.NewBlake2b()
	Marshaller = &marshal.GogoProtoMarshalizer{}
)
