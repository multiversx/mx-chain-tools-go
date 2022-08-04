package main

import (
	"context"
	"testing"

	"github.com/ElrondNetwork/elrond-tools-go/hdKeysGenerator/common"
	"github.com/stretchr/testify/require"
)

const (
	// See https://raw.githubusercontent.com/ElrondNetwork/elrond-sdk-testwallets/main/users/mnemonic.txt
	DummyMnemonic = "moral volcano peasant pass circle pen over picture flat shop clap goat never lyrics gather prepare woman film husband gravity behind test tiger improve"
)

func TestInlineSortKeysByIndexes(t *testing.T) {
	keys := []common.GeneratedKey{
		{
			AccountIndex: 0,
			AddressIndex: 4,
		},
		{
			AccountIndex: 0,
			AddressIndex: 3,
		},
		{
			AccountIndex: 4,
			AddressIndex: 1,
		},
		{
			AccountIndex: 3,
			AddressIndex: 3,
		},
	}

	sortedKeys := []common.GeneratedKey{
		{
			AccountIndex: 0,
			AddressIndex: 3,
		},
		{
			AccountIndex: 0,
			AddressIndex: 4,
		},
		{
			AccountIndex: 3,
			AddressIndex: 3,
		},
		{
			AccountIndex: 4,
			AddressIndex: 1,
		},
	}

	inlineSortKeysByIndexes(keys)

	require.Equal(t, sortedKeys, keys)
}

func TestGenerateKeysInParallel_GeneratesKeyAsInSequence(t *testing.T) {
	numKeys := 9
	startIndex := 1
	args := argsGenerateKeysInParallel{
		numKeys:    numKeys,
		startIndex: startIndex,
		numTasks:   4,
	}

	noConstraints, _ := newConstraints(3, common.OptionalUint32{}, common.OptionalUint32{})
	keys, err := generateKeysInParallel(context.Background(), args, DummyMnemonic, noConstraints)

	require.Nil(t, err)
	require.Equal(t, numKeys, len(keys))

	// See https://github.com/ElrondNetwork/elrond-sdk-testwallets/blob/main/users
	require.Equal(t, 1, keys[0].AddressIndex)
	require.Equal(t, "erd1spyavw0956vq68xj8y4tenjpq2wd5a9p2c6j8gsz7ztyrnpxrruqzu66jx", keys[0].Address)
	require.Equal(t, 2, keys[1].AddressIndex)
	require.Equal(t, "erd1k2s324ww2g0yj38qn2ch2jwctdy8mnfxep94q9arncc6xecg3xaq6mjse8", keys[1].Address)
	require.Equal(t, 3, keys[2].AddressIndex)
	require.Equal(t, "erd1kyaqzaprcdnv4luvanah0gfxzzsnpaygsy6pytrexll2urtd05ts9vegu7", keys[2].Address)
	require.Equal(t, 4, keys[3].AddressIndex)
	require.Equal(t, "erd18tudnj2z8vjh0339yu3vrkgzz2jpz8mjq0uhgnmklnap6z33qqeszq2yn4", keys[3].Address)
	require.Equal(t, 5, keys[4].AddressIndex)
	require.Equal(t, "erd1kdl46yctawygtwg2k462307dmz2v55c605737dp3zkxh04sct7asqylhyv", keys[4].Address)
	require.Equal(t, 6, keys[5].AddressIndex)
	require.Equal(t, "erd1r69gk66fmedhhcg24g2c5kn2f2a5k4kvpr6jfw67dn2lyydd8cfswy6ede", keys[5].Address)
	require.Equal(t, 7, keys[6].AddressIndex)
	require.Equal(t, "erd1dc3yzxxeq69wvf583gw0h67td226gu2ahpk3k50qdgzzym8npltq7ndgha", keys[6].Address)
	require.Equal(t, 8, keys[7].AddressIndex)
	require.Equal(t, "erd13x29rvmp4qlgn4emgztd8jgvyzdj0p6vn37tqxas3v9mfhq4dy7shalqrx", keys[7].Address)
	require.Equal(t, 9, keys[8].AddressIndex)
	require.Equal(t, "erd1fggp5ru0jhcjrp5rjqyqrnvhr3sz3v2e0fm3ktknvlg7mcyan54qzccnan", keys[8].Address)
}
