package main

import (
	"fmt"
	"github.com/tidwall/gjson"
	"io"
	"math/big"
	"net/http"
	"strings"
)

type nftBalanceGetter struct {
	proxyURL string
}

func newNFTBalanceGetter(proxyURL string) *nftBalanceGetter {
	return &nftBalanceGetter{
		proxyURL: proxyURL,
	}
}

func (eg *nftBalanceGetter) getBalance(address, token string) (string, error) {
	tokenID, nonce := getTokenIDAndNonce(token)
	return eg.fetchNFTBalanceWithRetrial(address, tokenID, nonce)
}

func getTokenIDAndNonce(token string) (string, uint64) {
	tokens := strings.Split(token, "-")
	tokenID := tokens[0] + "-" + tokens[1]

	nonce := big.NewInt(0)
	nonce.SetString(tokens[2], 16)

	return tokenID, nonce.Uint64()
}

func (eg *nftBalanceGetter) fetchNFTBalanceWithRetrial(address, tokenID string, nonce uint64) (string, error) {
	ctRetrials := 0

	for ctRetrials < maxRequestsRetrial {
		url := fmt.Sprintf("%s/address/%s/nft/%s/nonce/%d", eg.proxyURL, address, tokenID, nonce)
		resp, errHttp := http.Get(url)
		body, errBody := eg.getBody(resp)
		if errHttp == nil && errBody == nil {
			return gjson.Get(body, "data.tokenData.balance").String(), nil
		}

		log.Warn("could not get balance; retrying...",
			"address", address,
			"tokenID", tokenID,
			"token nonce", nonce,
			"error http", errHttp,
			"error body", errBody,
			"num retrials", ctRetrials)

		ctRetrials++
	}

	return "", fmt.Errorf("could not get adress's tokens = %s after num of retrials = %d", address, maxRequestsRetrial)
}

func (eg *nftBalanceGetter) getBody(response *http.Response) (string, error) {
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("could not ready bytes from body; error: %w", err)
	}

	bodyStr := string(bodyBytes)
	bodyErr := gjson.Get(bodyStr, "error").String()
	if len(bodyErr) != 0 {
		return "", fmt.Errorf("got error in body response when getting esdt tokens, proxy url: %s, body error: %s", eg.proxyURL, bodyErr)
	}

	return string(bodyBytes), nil
}
