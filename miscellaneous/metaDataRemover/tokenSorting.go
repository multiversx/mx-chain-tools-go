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
		return consecutiveNonces1 > consecutiveNonces2
	})

	consecutiveNoncesOverThreshold := uint64(0)
	consecutiveNoncesUnderThreshold := uint64(0)
	for _, r := range ret {
		if r.interval.end-r.interval.start+1 >= 50 {
			consecutiveNoncesOverThreshold += r.interval.end - r.interval.start + 1
		} else {
			consecutiveNoncesUnderThreshold += r.interval.end - r.interval.start + 1
		}
	}
	totalNonces := consecutiveNoncesOverThreshold + consecutiveNoncesUnderThreshold
	//for _, r := range ret {
	//	log.Info("found", "tokenID", r.tokenID, "consecutive nonces", r.interval.end - r.interval.start + 1)
	//}

	log.Info("found",
		"consecutiveNoncesOverThreshold", consecutiveNoncesOverThreshold,
		"consecutiveNoncesUnderThreshold", consecutiveNoncesUnderThreshold,
		"% * consecutiveNoncesOverThreshold of total nonces", (100*consecutiveNoncesOverThreshold)/totalNonces)
	return ret
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
