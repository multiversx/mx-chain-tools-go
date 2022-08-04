package export

import "github.com/ElrondNetwork/elrond-tools-go/hdKeysGenerator/common"

type formatter interface {
	toText(keys []common.GeneratedKey) (string, error)
}
