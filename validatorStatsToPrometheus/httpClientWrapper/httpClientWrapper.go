package httpClientWrapper

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/multiversx/mx-chain-core-go/core/check"
)

const (
	validatorStatisticsEndpoint = "validator/statistics"
)

type httpClientWrapper struct {
	httpClient HttpClient
}

// NewHttpClientWrapper returns a new instance of httpClientWrapper
func NewHttpClientWrapper(httpClient HttpClient) (*httpClientWrapper, error) {
	if check.IfNil(httpClient) {
		return nil, ErrNilHttpClient
	}

	return &httpClientWrapper{
		httpClient: httpClient,
	}, nil
}

// GetValidatorStatistics makes a http request and returns the validator statistics
func (hcw *httpClientWrapper) GetValidatorStatistics(ctx context.Context) (map[string]*ValidatorStatistics, error) {
	buff, err := hcw.getData(ctx, validatorStatisticsEndpoint)
	if err != nil {
		return nil, err
	}

	var resp ValidatorStatisticsApiResponse
	err = json.Unmarshal(buff, &resp)
	if err != nil {
		return nil, err
	}
	if resp.Data.Statistics == nil {
		return nil, fmt.Errorf("%w while getting validator statistics", ErrEmptyData)
	}

	return resp.Data.Statistics, nil
}

func (hcw *httpClientWrapper) getData(ctx context.Context, endpoint string) ([]byte, error) {
	buff, code, err := hcw.httpClient.GetHTTP(ctx, endpoint)
	if err != nil || code != http.StatusOK {
		return nil, createHTTPStatusError(code, err)
	}
	if len(buff) == 0 {
		return nil, fmt.Errorf("%w while calling %s, code %d", ErrEmptyData, endpoint, code)
	}

	return buff, nil
}

func createHTTPStatusError(httpStatusCode int, err error) error {
	if err == nil {
		err = ErrHTTPStatusCodeIsNotOK
	}

	return fmt.Errorf("%w, returned http status: %d, %s",
		err, httpStatusCode, http.StatusText(httpStatusCode))
}

// IsInterfaceNil returns true if there is no value under the interface
func (hcw *httpClientWrapper) IsInterfaceNil() bool {
	return hcw == nil
}
