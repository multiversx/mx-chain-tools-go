package main

import (
	"fmt"
	"math/big"
	"sort"
	"strings"
)

func sortTokensIDByNonce(tokens map[string]struct{}) (map[string][]uint64, error) {
	ret := make(map[string][]uint64)
	for token := range tokens {
		splits := strings.Split(token, "-")
		if len(splits) != 3 {
			return nil, fmt.Errorf("found invalid format in token = %s; expected [ticker-randSequence-nonce]", token)
		}

		tokenID := splits[0] + "-" + splits[1] // ticker-randSequence
		nonceBI := big.NewInt(0)
		nonceBI.SetString(splits[2], 16)

		ret[tokenID] = append(ret[tokenID], nonceBI.Uint64())
	}

	log.Info("found", "num of tokensID", len(ret))
	for _, nonces := range ret {
		sort.SliceStable(nonces, func(i, j int) bool {
			return nonces[i] < nonces[j]
		})
	}

	return ret, nil
}

func groupTokensByIntervals(tokens map[string][]uint64) map[string][]*interval {
	ret := make(map[string][]*interval)

	for token, nonces := range tokens {
		numNonces := len(nonces)
		for idx := 0; idx < numNonces; idx++ {
			nonce := nonces[idx]
			if idx+1 >= numNonces {
				ret[token] = append(ret[token], &interval{
					start: nonce,
					end:   nonce,
				})
				break
			}

			currInterval := &interval{start: nonce}
			numConsecutiveNonces := uint64(0)
			for idx < numNonces-1 {
				currNonce := nonces[idx]
				nextNonce := nonces[idx+1]
				if nextNonce-currNonce > 1 {
					break
				}

				numConsecutiveNonces++
				idx++
			}

			currInterval.end = currInterval.start + numConsecutiveNonces
			ret[token] = append(ret[token], currInterval)
		}
	}

	//for token, intervals := range ret {
	//	log.Info("found", "tokenID", token, "num of nonces", len(tokens[token]), "num of intervals", len(intervals))
	//}

	return ret
}

type tokenWithInterval struct {
	tokenID  string
	interval *interval
}

func sortTokensByMaxConsecutiveNonces(tokens map[string][]*interval) []*tokenWithInterval {
	ret := make([]*tokenWithInterval, 0)
	for token, intervals := range tokens {
		for _, currInterval := range intervals {
			ret = append(ret, &tokenWithInterval{
				tokenID:  token,
				interval: currInterval,
			})

		}
	}

	sort.SliceStable(ret, func(i, j int) bool {
		consecutiveNonces1 := ret[i].interval.end - ret[i].interval.start + 1
		consecutiveNonces2 := ret[j].interval.end - ret[j].interval.start + 1

		if consecutiveNonces1 == consecutiveNonces2 {
			return ret[i].tokenID < ret[j].tokenID
		}

		return consecutiveNonces1 > consecutiveNonces2
	})

	return ret
}

func sortTokensInBulks(tokens []*tokenWithInterval, intervalBulkSize uint64) [][]*tokenData {
	intervalsInBulk := make([][]*tokenData, 0, intervalBulkSize)

	currBulk := make(map[string][]*interval, 0)
	numNoncesInBulk := uint64(0)

	tokensCopy := make([]*tokenWithInterval, len(tokens))
	copy(tokensCopy, tokens)

	index := 0
	for index < len(tokensCopy) {
		currTokenData := tokensCopy[index]
		currInterval := currTokenData.interval
		currTokenID := currTokenData.tokenID

		noncesInInterval := currInterval.end - currInterval.start + 1
		availableSlots := intervalBulkSize - numNoncesInBulk
		if availableSlots >= noncesInInterval {
			currBulk[currTokenID] = append(currBulk[currTokenID], currInterval)
			numNoncesInBulk += noncesInInterval
		} else {
			first, second := splitInterval(currInterval, availableSlots)

			tokensCopy = insert(tokensCopy, index+1, &tokenWithInterval{tokenID: currTokenID, interval: second})
			currBulk[currTokenID] = append(currBulk[currTokenID], first)
			numNoncesInBulk += availableSlots
		}

		bulkFull := numNoncesInBulk == intervalBulkSize
		lastInterval := index == len(tokensCopy)-1

		if bulkFull || lastInterval {
			intervalsInBulk = append(intervalsInBulk, tokensMapToOrderedArray(currBulk))

			currBulk = make(map[string][]*interval, 0)
			numNoncesInBulk = 0
		}

		index++
	}

	return intervalsInBulk
}

func insert(tokens []*tokenWithInterval, index int, token *tokenWithInterval) []*tokenWithInterval {
	if len(tokens) <= index {
		return append(tokens, token)
	}
	tokens = append(tokens[:index+1], tokens[index:]...)
	tokens[index] = token
	return tokens
}

func tokensMapToOrderedArray(tokens map[string][]*interval) []*tokenData {
	ret := make([]*tokenData, 0, len(tokens))

	for token, intervals := range tokens {
		ret = append(ret, &tokenData{
			tokenID:   token,
			intervals: intervals,
		})
	}

	sort.SliceStable(ret, func(i, j int) bool {
		return ret[i].tokenID < ret[j].tokenID
	})

	return ret
}