package main

import (
	"errors"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/tidwall/gjson"
)

const (
	maxRequestsRetrial = 10
	multipleSearchBulk = 10000
)

type nftBalancesGetter interface {
	getBalance(address, token string) (string, error)
}

type elasticMultiSearchClient interface {
	GetMultiple(index string, requests []string) ([]byte, error)
}

type crossExtraTokensChecker struct {
	nftBalancesGetter nftBalancesGetter
	elasticClient     elasticMultiSearchClient
}

func newExtraTokensCrossChecker(client elasticMultiSearchClient, nftBalancesGetter nftBalancesGetter) (crossTokenChecker, error) {
	if client == nil {
		return nil, errors.New("nil elastic client provided")
	}
	if nftBalancesGetter == nil {
		return nil, errors.New("nil nft balances getter provided")
	}

	return &crossExtraTokensChecker{
		nftBalancesGetter: nftBalancesGetter,
		elasticClient:     client,
	}, nil
}

func (ctc *crossExtraTokensChecker) crossCheckExtraTokens(tokens map[string]struct{}) ([]string, error) {
	numTokens := len(tokens)
	log.Info("starting to cross-check", "num of tokens", numTokens)

	bulkSize := core.MinInt(multipleSearchBulk, numTokens)
	tokensThatStillExist := make([]string, 0)
	requests := make([]string, 0, bulkSize)
	currTokenIdx := 0
	ctRequests := 0
	for token := range tokens {
		currTokenIdx++
		requests = append(requests, createRequest(token))

		notEnoughRequests := len(requests) < bulkSize
		notLastBulk := currTokenIdx != numTokens
		if notEnoughRequests && notLastBulk {
			continue
		}

		respBytes, err := ctc.elasticClient.GetMultiple("accountsesdt", requests)
		if err != nil {
			log.Error("elasticClient.GetMultiple(accountsesdt, requests)",
				"error", err,
				"requests", requests)
			return nil, err
		}

		responses := gjson.Get(string(respBytes), "responses").Array()
		crossCheckFailedTokens, err := ctc.checkIndexerResponse(requests, responses)
		if err != nil {
			return nil, err
		}

		go printProgress(numTokens, currTokenIdx)

		ctRequests += len(requests)
		requests = make([]string, 0, bulkSize)
		tokensThatStillExist = append(tokensThatStillExist, crossCheckFailedTokens...)
	}

	log.Info("finished cross-checking",
		"total num of tokens", numTokens,
		"total num of tokens cross-checked", currTokenIdx,
		"total num of tokens requests in indexer", ctRequests)

	if numTokens != currTokenIdx || numTokens != ctRequests {
		return nil, errors.New("failed to cross check all tokens, check logs")
	}

	return tokensThatStillExist, nil
}

func (ctc *crossExtraTokensChecker) checkIndexerResponse(requests []string, responses []gjson.Result) ([]string, error) {
	tokensThatStillExist := make([]string, 0)
	for idxRequestedToken, res := range responses {
		hits := res.Get("hits.hits").Array()
		if len(hits) != 0 {
			token := gjson.Get(requests[idxRequestedToken], "query.match.identifier.query").String()
			log.Debug("found token in indexer with hits/accounts",
				"token", token,
				"num hits/accounts", len(hits))

			checkFailed, err := ctc.crossCheckToken(hits, token)
			if err != nil {
				return nil, err
			}

			if checkFailed {
				tokensThatStillExist = append(tokensThatStillExist, token)
			}
		}
		idxRequestedToken++
	}

	return tokensThatStillExist, nil
}

func (ctc *crossExtraTokensChecker) crossCheckToken(hits []gjson.Result, token string) (bool, error) {
	checkFailed := false
	for _, hit := range hits {
		address := hit.Get("_source.address").String()
		balance, err := ctc.nftBalancesGetter.getBalance(address, token)
		if err != nil {
			return false, err
		}

		log.Debug("checking gateway if token still exists in trie",
			"token", token,
			"address", address)

		if balance != "0" {
			checkFailed = true
			log.Error("cross-check failed; found token which is still in other address",
				"token", token,
				"balance", balance,
				"address", address)
			break
		}

		log.Warn("possible indexer problem",
			"token", token,
			"hit in address", address,
			"found in trie", false)
	}

	return checkFailed, nil
}

func createRequest(token string) string {
	return `{"query" : {"match" : { "identifier": {"query":"` + token + `","operator":"and"}}}}`
}

func printProgress(numTokens, numTokensCrossChecked int) {
	log.Info("status",
		"num cross checked tokens", numTokensCrossChecked,
		"remaining num of tokens to check", numTokens-numTokensCrossChecked,
		"progress(%)", (100*numTokensCrossChecked)/numTokens) // this should not panic with div by zero, since func is only called if numTokens > 0
}
