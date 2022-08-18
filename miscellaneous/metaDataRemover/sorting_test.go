package main

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func requireSameSliceDifferentOrder(t *testing.T, s1, s2 [][]byte) {
	require.Equal(t, len(s1), len(s2))

	for _, elemInS1 := range s1 {
		require.Contains(t, s2, elemInS1)
	}
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

func TestCreateTxData(t *testing.T) {
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
				start: 5,
				end:   6,
			},
		},
	}

	txsData, err := createTxsData(tokensIntervals, 2)
	require.Nil(t, err)

	expectedTxsData := [][]byte{
		[]byte("ESDTDeleteMetadata@746f6b656e31@02@00@00@01@01"), // token1: 2 intervals: [0,0];[1,1]
		[]byte("ESDTDeleteMetadata@746f6b656e31@02@02@03@04@08"), // token1: 2 intervals: [2,3];[4,8]
		[]byte("ESDTDeleteMetadata@746f6b656e32@01@01@05"),       // token2: 1 interval:  [1,5]
		[]byte("ESDTDeleteMetadata@746f6b656e33@02@00@00@01@04"), // token3: 2 intervals: [0,0];[1,4]
		[]byte("ESDTDeleteMetadata@746f6b656e33@01@05@06"),       // token3: 1 interval:  [5,6]
	}
	requireSameSliceDifferentOrder(t, txsData, expectedTxsData)
}
