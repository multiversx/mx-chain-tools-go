package export

import (
	"github.com/ElrondNetwork/elrond-go/state"
)

type trieWrapper interface {
	GetUserAccounts(rootHash []byte, predicate func(*state.UserAccountData) bool) ([]*state.UserAccountData, error)
}

type formatter interface {
	toText(accounts []*state.UserAccountData, args formatterArgs) (string, error)
	getFileExtension() string
}
