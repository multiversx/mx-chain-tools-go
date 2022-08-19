package mocks

import (
	"context"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
)

type ProxyStub struct {
	GetNetworkConfigCalled               func(ctx context.Context) (*data.NetworkConfig, error)
	GetDefaultTransactionArgumentsCalled func(
		ctx context.Context,
		address core.AddressHandler,
		networkConfigs *data.NetworkConfig,
	) (data.ArgCreateTransaction, error)
}

func (ps *ProxyStub) GetNetworkConfig(ctx context.Context) (*data.NetworkConfig, error) {
	if ps.GetNetworkConfigCalled != nil {
		return ps.GetNetworkConfigCalled(ctx)
	}

	return nil, nil
}

func (ps *ProxyStub) GetDefaultTransactionArguments(
	ctx context.Context,
	address core.AddressHandler,
	networkConfigs *data.NetworkConfig,
) (data.ArgCreateTransaction, error) {
	if ps.GetDefaultTransactionArgumentsCalled != nil {
		return ps.GetDefaultTransactionArgumentsCalled(ctx, address, networkConfigs)
	}

	return data.ArgCreateTransaction{}, nil
}
