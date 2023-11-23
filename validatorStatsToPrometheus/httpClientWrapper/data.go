package httpClientWrapper

// ValidatorStatistics defines the validator statistics api response
type ValidatorStatistics struct {
	TempRating                         float32 `json:"tempRating"`
	NumLeaderSuccess                   uint32  `json:"numLeaderSuccess"`
	NumLeaderFailure                   uint32  `json:"numLeaderFailure"`
	NumValidatorSuccess                uint32  `json:"numValidatorSuccess"`
	NumValidatorFailure                uint32  `json:"numValidatorFailure"`
	NumValidatorIgnoredSignatures      uint32  `json:"numValidatorIgnoredSignatures"`
	Rating                             float32 `json:"rating"`
	RatingModifier                     float32 `json:"ratingModifier"`
	TotalNumLeaderSuccess              uint32  `json:"totalNumLeaderSuccess"`
	TotalNumLeaderFailure              uint32  `json:"totalNumLeaderFailure"`
	TotalNumValidatorSuccess           uint32  `json:"totalNumValidatorSuccess"`
	TotalNumValidatorFailure           uint32  `json:"totalNumValidatorFailure"`
	TotalNumValidatorIgnoredSignatures uint32  `json:"totalNumValidatorIgnoredSignatures"`
	ShardId                            uint32  `json:"shardId"`
	ValidatorStatus                    string  `json:"validatorStatus,omitempty"`
}

// ValidatorStatisticsApiResponse defines the response received when calling /validator/statistics endpoint
type ValidatorStatisticsApiResponse struct {
	Data struct {
		Statistics map[string]*ValidatorStatistics `json:"statistics"`
	} `json:"data"`
	Error string `json:"error"`
	Code  string `json:"code"`
}
