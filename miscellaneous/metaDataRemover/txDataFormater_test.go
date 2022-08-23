package main

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCreateTxsData(t *testing.T) {
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

	// Tokens/tx = 1
	txsData, err := createTxsData(tokensIntervals, 1)
	require.Nil(t, err)
	expectedTxsData := [][]byte{
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@00@00"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@01@01"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@02@02"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@04@04"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@03@03"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@05@05"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@06@06"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@07@07"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@08@08"),
		[]byte("ESDTDeleteMetadata@746f6b656e32@01@01@01"),
		[]byte("ESDTDeleteMetadata@746f6b656e32@01@02@02"),
		[]byte("ESDTDeleteMetadata@746f6b656e32@01@03@03"),
		[]byte("ESDTDeleteMetadata@746f6b656e32@01@04@04"),
		[]byte("ESDTDeleteMetadata@746f6b656e32@01@05@05"),
		[]byte("ESDTDeleteMetadata@746f6b656e33@01@00@00"),
		[]byte("ESDTDeleteMetadata@746f6b656e33@01@01@01"),
		[]byte("ESDTDeleteMetadata@746f6b656e33@01@06@06"),
		[]byte("ESDTDeleteMetadata@746f6b656e33@01@02@02"),
		[]byte("ESDTDeleteMetadata@746f6b656e33@01@07@07"),
		[]byte("ESDTDeleteMetadata@746f6b656e33@01@03@03"),
		[]byte("ESDTDeleteMetadata@746f6b656e33@01@04@04"),
	}
	require.Equal(t, expectedTxsData, txsData)

	// Tokens/tx = 2
	txsData, err = createTxsData(tokensIntervals, 2)
	require.Nil(t, err)
	expectedTxsData = [][]byte{
		[]byte("ESDTDeleteMetadata@746f6b656e31@02@00@00@01@01"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@02@03"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@04@05"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@06@07"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@08@08@746f6b656e32@01@01@01"),
		[]byte("ESDTDeleteMetadata@746f6b656e32@01@02@03"),
		[]byte("ESDTDeleteMetadata@746f6b656e32@01@04@05"),
		[]byte("ESDTDeleteMetadata@746f6b656e33@02@00@00@01@01"),
		[]byte("ESDTDeleteMetadata@746f6b656e33@01@06@07"),
		[]byte("ESDTDeleteMetadata@746f6b656e33@01@02@03"),
		[]byte("ESDTDeleteMetadata@746f6b656e33@01@04@04"),
	}
	require.Equal(t, expectedTxsData, txsData)

	// Tokens/tx = 3
	txsData, err = createTxsData(tokensIntervals, 3)
	require.Nil(t, err)
	expectedTxsData = [][]byte{
		[]byte("ESDTDeleteMetadata@746f6b656e31@03@00@00@01@01@02@02"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@04@06"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@02@03@03@07@08"),
		[]byte("ESDTDeleteMetadata@746f6b656e32@01@01@03"),
		[]byte("ESDTDeleteMetadata@746f6b656e32@01@04@05@746f6b656e33@01@00@00"),
		[]byte("ESDTDeleteMetadata@746f6b656e33@01@01@03"),
		[]byte("ESDTDeleteMetadata@746f6b656e33@02@06@07@04@04"),
	}
	require.Equal(t, expectedTxsData, txsData)

	// Tokens/tx = 4
	txsData, err = createTxsData(tokensIntervals, 4)
	require.Nil(t, err)
	expectedTxsData = [][]byte{
		[]byte("ESDTDeleteMetadata@746f6b656e31@03@00@00@01@01@02@03"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@04@07"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@08@08@746f6b656e32@01@01@03"),
		[]byte("ESDTDeleteMetadata@746f6b656e32@01@04@05@746f6b656e33@02@00@00@01@01"),
		[]byte("ESDTDeleteMetadata@746f6b656e33@02@06@07@02@03"),
		[]byte("ESDTDeleteMetadata@746f6b656e33@01@04@04"),
	}
	require.Equal(t, expectedTxsData, txsData)

	// Tokens/tx = 5
	txsData, err = createTxsData(tokensIntervals, 5)
	require.Nil(t, err)
	expectedTxsData = [][]byte{
		[]byte("ESDTDeleteMetadata@746f6b656e31@04@00@00@01@01@02@03@04@04"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@05@08@746f6b656e32@01@01@01"),
		[]byte("ESDTDeleteMetadata@746f6b656e32@01@02@05@746f6b656e33@01@00@00"),
		[]byte("ESDTDeleteMetadata@746f6b656e33@02@01@04@06@06"),
		[]byte("ESDTDeleteMetadata@746f6b656e33@01@07@07"),
	}
	require.Equal(t, expectedTxsData, txsData)

	// Tokens/tx = 6
	txsData, err = createTxsData(tokensIntervals, 6)
	require.Nil(t, err)
	expectedTxsData = [][]byte{
		[]byte("ESDTDeleteMetadata@746f6b656e31@04@00@00@01@01@02@03@04@05"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@06@08@746f6b656e32@01@01@03"),
		[]byte("ESDTDeleteMetadata@746f6b656e32@01@04@05@746f6b656e33@02@00@00@01@03"),
		[]byte("ESDTDeleteMetadata@746f6b656e33@02@06@07@04@04"),
	}
	require.Equal(t, expectedTxsData, txsData)

	// Tokens/tx = 7
	txsData, err = createTxsData(tokensIntervals, 7)
	require.Nil(t, err)
	expectedTxsData = [][]byte{
		[]byte("ESDTDeleteMetadata@746f6b656e31@04@00@00@01@01@02@03@04@06"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@07@08@746f6b656e32@01@01@05"),
		[]byte("ESDTDeleteMetadata@746f6b656e33@03@00@00@01@04@06@07"),
	}
	require.Equal(t, expectedTxsData, txsData)

	// Tokens/tx = 8
	txsData, err = createTxsData(tokensIntervals, 8)
	require.Nil(t, err)
	expectedTxsData = [][]byte{
		[]byte("ESDTDeleteMetadata@746f6b656e31@04@00@00@01@01@02@03@04@07"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@08@08@746f6b656e32@01@01@05@746f6b656e33@02@00@00@01@01"),
		[]byte("ESDTDeleteMetadata@746f6b656e33@02@06@07@02@04"),
	}
	require.Equal(t, expectedTxsData, txsData)

	// Tokens/tx = 20
	txsData, err = createTxsData(tokensIntervals, 20)
	require.Nil(t, err)
	expectedTxsData = [][]byte{
		[]byte("ESDTDeleteMetadata@746f6b656e31@04@00@00@01@01@02@03@04@08@746f6b656e32@01@01@05@746f6b656e33@03@00@00@01@04@06@06"),
		[]byte("ESDTDeleteMetadata@746f6b656e33@01@07@07"),
	}
	require.Equal(t, expectedTxsData, txsData)

	// Tokens/tx >= 21
	expectedTxsData = [][]byte{
		[]byte("ESDTDeleteMetadata@746f6b656e31@04@00@00@01@01@02@03@04@08@746f6b656e32@01@01@05@746f6b656e33@03@00@00@01@04@06@07"),
	}
	for i := uint64(21); i < 100; i++ {
		txsData, err = createTxsData(tokensIntervals, i)
		require.Nil(t, err)
		require.Equal(t, expectedTxsData, txsData)
	}
}
