package export

import (
	"github.com/multiversx/mx-chain-go/state/accounts"
)

type trieWrapper interface {
	GetUserAccounts(rootHash []byte, predicate func(data *accounts.UserAccountData) bool) ([]*accounts.UserAccountData, error)
}

type formatter interface {
	toText(accounts []*accounts.UserAccountData, args formatterArgs) (string, error)
	getFileExtension() string
}
