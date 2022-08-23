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

	for token, intervals := range ret {
		log.Info("found", "tokenID", token, "num of nonces", len(tokens[token]), "num of intervals", len(intervals))
	}

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
