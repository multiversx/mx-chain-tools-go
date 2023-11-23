package httpClientWrapper

import "github.com/multiversx/mx-chain-go/state/accounts"

// ValidatorStatistics defines the validator statistics api response
type ValidatorStatistics = accounts.ValidatorApiResponse

// ValidatorStatisticsApiResponse defines the response received when calling /validator/statistics endpoint
type ValidatorStatisticsApiResponse struct {
	Data struct {
		Statistics map[string]*ValidatorStatistics `json:"statistics"`
	} `json:"data"`
	Error string `json:"error"`
	Code  string `json:"code"`
}
