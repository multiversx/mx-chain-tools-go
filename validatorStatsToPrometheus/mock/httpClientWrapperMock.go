package mock

import (
	"context"

	"github.com/multiversx/mx-chain-tools-go/jsonToPrometheus/httpClientWrapper"
)

// HTTPClientWrapperMock -
type HTTPClientWrapperMock struct {
	GetValidatorStatisticsCalled func(ctx context.Context) (map[string]*httpClientWrapper.ValidatorStatistics, error)
}

// GetValidatorStatistics -
func (mock *HTTPClientWrapperMock) GetValidatorStatistics(ctx context.Context) (map[string]*httpClientWrapper.ValidatorStatistics, error) {
	if mock.GetValidatorStatisticsCalled != nil {
		return mock.GetValidatorStatisticsCalled(ctx)
	}
	return make(map[string]*httpClientWrapper.ValidatorStatistics), nil
}

// IsInterfaceNil -
func (mock *HTTPClientWrapperMock) IsInterfaceNil() bool {
	return mock == nil
}
