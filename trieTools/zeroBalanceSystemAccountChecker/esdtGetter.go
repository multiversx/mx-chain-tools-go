package main

import (
	"fmt"
	"github.com/tidwall/gjson"
	"net/http"
)

type esdtsGetter struct {
	proxyURL string
	cache    map[string]map[string]struct{}
}

func newESDTsGetter(proxyURL string) *esdtsGetter {
	return &esdtsGetter{
		proxyURL: proxyURL,
		cache:    make(map[string]map[string]struct{}),
	}
}

func (eg *esdtsGetter) getTokens(address string) (map[string]struct{}, error) {
	tokens, exist := eg.cache[address]
	if exist {
		return tokens, nil
	}

	return eg.fetchTokensFromProxy(address)
}

func (eg *esdtsGetter) fetchTokensFromProxy(address string) (map[string]struct{}, error) {
	ctRetrials := 0

	for ctRetrials < maxIndexerRetrials {
		url := fmt.Sprintf("%s/address/%s/esdt", eg.proxyURL, address)
		resp, err := http.Get(url)
		if err == nil {
			eg.saveInCache(address, resp)
			return eg.cache[address], nil
		}

		log.Warn("could tokens",
			"address", address,
			"error", err,
			"response body", getBody(resp),
			"num retrials", ctRetrials)

		ctRetrials++
	}

	return nil, fmt.Errorf("could not get adress's tokens = %s after num of retrials = %d", address, maxIndexerRetrials)
}

func (eg *esdtsGetter) saveInCache(address string, resp *http.Response) {
	body := getBody(resp)
	esdts := gjson.Get(body, "data.esdts").Map()
	for esdt := range esdts {
		esdtEntry := make(map[string]struct{})
		esdtEntry[esdt] = struct{}{}

		eg.cache[address] = esdtEntry
	}
}
