package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSortTokensIDByNonce(t *testing.T) {
	tokens := map[string]struct{}{
		"token1-rand1-0f": {},
		"token1-rand1-01": {},
		"token1-rand1-0a": {},
		"token1-rand1-0b": {},

		"token2-rand2-04": {},

		"token3-rand3-04": {},
		"token3-rand3-08": {},
	}

	sortedTokens, err := sortTokensIDByNonce(tokens)
	require.Nil(t, err)
	require.Equal(t, sortedTokens, map[string][]uint64{
		"token1-rand1": {1, 10, 11, 15},
		"token2-rand2": {4},
		"token3-rand3": {4, 8},
	})
}

func TestGroupTokensByIntervals(t *testing.T) {
	tokens := map[string][]uint64{
		"token1": {1, 2, 3, 8, 9, 10},
		"token2": {1},
		"token3": {3, 9},
		"token4": {11, 12},
		"token5": {10, 100, 101, 102, 111},
		"token6": {4, 5, 6, 7},
	}

	sortedTokens := groupTokensByIntervals(tokens)
	require.Equal(t, sortedTokens,
		map[string][]*interval{
			"token1": {
				{
					start: 1,
					end:   3,
				},
				{
					start: 8,
					end:   10,
				},
			},
			"token2": {
				{
					start: 1,
					end:   1,
				},
			},
			"token3": {
				{
					start: 3,
					end:   3,
				},
				{
					start: 9,
					end:   9,
				},
			},
			"token4": {
				{
					start: 11,
					end:   12,
				},
			},
			"token5": {
				{
					start: 10,
					end:   10,
				},
				{
					start: 100,
					end:   102,
				},
				{
					start: 111,
					end:   111,
				},
			},
			"token6": {
				{
					start: 4,
					end:   7,
				},
			},
		},
	)
}

func TestSortTokensInBulk(t *testing.T) {
	tokensIntervals := map[string][]*interval{
		"token1": {
			{
				start: 0,
				end:   0,
			},
			{
				start: 1,
				end:   1,
			},
			{
				start: 2,
				end:   3,
			},
			{
				start: 4,
				end:   8,
			},
		},
		"token2": {
			{
				start: 1,
				end:   5,
			},
		},
		"token3": {
			{
				start: 0,
				end:   0,
			},
			{
				start: 1,
				end:   4,
			},
			{
				start: 6,
				end:   7,
			},
		},
	}

	in := sortTokensByMaxConsecutiveNonces(tokensIntervals)
	out := sortTokensInBulks(in, 6)

	for idx, bulk := range out {
		fmt.Println("###### bulk idx", idx)

		for _, tkData := range bulk {
			for _, intrv := range tkData.intervals {
				fmt.Println("---- token id", tkData.tokenID, "start", intrv.start, "end", intrv.end)
			}
		}

	}

}
