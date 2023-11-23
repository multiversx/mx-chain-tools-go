package validatorStatsToPrometheus

import (
	"context"

	"github.com/multiversx/mx-chain-tools-go/jsonToPrometheus/httpClientWrapper"
)

// HttpClientWrapper defines the behavior of wrapper over HttpClient
type HttpClientWrapper interface {
	GetValidatorStatistics(ctx context.Context) (map[string]*httpClientWrapper.ValidatorStatistics, error)
	IsInterfaceNil() bool
}
