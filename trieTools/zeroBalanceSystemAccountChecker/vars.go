package main

import (
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

var (
	log            = logger.GetOrCreate("main")
	jsonMarshaller = &marshal.JsonMarshalizer{}
)
