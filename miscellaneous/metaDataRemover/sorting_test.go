package main

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

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
	ret := createTxData(
		map[string][]*interval{
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
		})

	fmt.Println(ret)
}

func TestSplitIntervals(t *testing.T) {
	ret := splitIntervals("token", []*interval{
		{
			start: 1,
			end:   1,
		},
		{
			start: 3,
			end:   6,
		},
		{
			start: 8,
			end:   10,
		},
	})

	fmt.Println(ret)
}
