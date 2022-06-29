package export

import "github.com/ElrondNetwork/elrond-go/state"

type trieWrapper interface {
	GetAllUserAccounts(rootHash []byte) ([]*state.UserAccountData, error)
}
