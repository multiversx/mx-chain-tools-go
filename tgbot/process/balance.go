package process

import (
	"errors"
	"math/big"

	"github.com/ElrondNetwork/elrond-sdk-erdgo/blockchain"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
)

func GetAddressBalance(gatewayURL string, address string) (*big.Int, error) {
	proxy := blockchain.NewElrondProxy(gatewayURL, nil)

	addressHandler, err := data.NewAddressFromBech32String(address)
	if err != nil {
		return nil, err
	}

	accountResponse, err := proxy.GetAccount(addressHandler)
	if err != nil {
		return nil, err
	}

	balanceBig, ok := big.NewInt(0).SetString(accountResponse.Balance, 10)
	if !ok {
		return nil, errors.New("invalid balance from response")
	}

	return balanceBig, nil
}
