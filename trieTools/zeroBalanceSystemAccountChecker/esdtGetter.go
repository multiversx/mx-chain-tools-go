package main

import (
	"fmt"
	"github.com/tidwall/gjson"
	"io"
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
		log.Debug("fetching esdts from cache", "address", address)
		return tokens, nil
	}
	log.Debug("fetching esdts from proxy", "address", address)
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
			"response body", eg.getBody(resp),
			"num retrials", ctRetrials)

		ctRetrials++
	}

	return nil, fmt.Errorf("could not get adress's tokens = %s after num of retrials = %d", address, maxIndexerRetrials)
}

func (eg *esdtsGetter) saveInCache(address string, resp *http.Response) {
	body := eg.getBody(resp)
	esdts := gjson.Get(body, "data.esdts").Map()
	log.Debug("saving in cache", "address", address, "num esdts", len(esdts))
	for esdt := range esdts {
		esdtEntry := make(map[string]struct{})
		esdtEntry[esdt] = struct{}{}

		eg.cache[address] = esdtEntry
	}
}

func (eg *esdtsGetter) getBody(response *http.Response) string {
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		log.Error("could not ready bytes from body", "error", err)
		return ""
	}

	bodyStr := string(bodyBytes)
	bodyErr := gjson.Get(bodyStr, "error").String()
	if len(bodyErr) != 0 {
		log.Error("got error in body response when getting esdt tokens", "proxy url", eg.proxyURL, "error", bodyErr)
	}

	return string(bodyBytes)
}
