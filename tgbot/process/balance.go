package process

import (
	"context"
	"errors"
	"math/big"
	"time"

	"github.com/multiversx/mx-sdk-go/blockchain"
	"github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
)

func GetAddressBalance(gatewayURL string, address string) (*big.Int, error) {
	proxy, err := blockchain.NewProxy(blockchain.ArgsProxy{
		ProxyURL:            gatewayURL,
		CacheExpirationTime: time.Second,
		EntityType:          core.Proxy,
	})
	if err != nil {
		return nil, err
	}

	addressHandler, err := data.NewAddressFromBech32String(address)
	if err != nil {
		return nil, err
	}

	accountResponse, err := proxy.GetAccount(context.Background(), addressHandler)
	if err != nil {
		return nil, err
	}

	balanceBig, ok := big.NewInt(0).SetString(accountResponse.Balance, 10)
	if !ok {
		return nil, errors.New("invalid balance from response")
	}

	return balanceBig, nil
}
