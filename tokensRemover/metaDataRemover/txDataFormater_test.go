package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func toTxData(t *testing.T, tokens []*tokenWithInterval, numNoncesPerTx uint64) [][]byte {
	tokensInBulks := groupTokenIntervalsInBulks(tokens, numNoncesPerTx)
	txsData, err := createTxsData(tokensInBulks)
	require.Nil(t, err)

	return txsData
}

func TestCreateTxsData(t *testing.T) {
	tokensIntervals := []*tokenWithInterval{
		{
			tokenID: "token1",
			interval: &interval{
				start: 4,
				end:   8,
			},
		},
		{
			tokenID: "token2",
			interval: &interval{
				start: 1,
				end:   5,
			},
		},
		{
			tokenID: "token3",
			interval: &interval{
				start: 1,
				end:   4,
			},
		},
		{
			tokenID: "token1",
			interval: &interval{
				start: 2,
				end:   3,
			},
		},
		{
			tokenID: "token3",
			interval: &interval{
				start: 6,
				end:   7,
			},
		},
		{
			tokenID: "token1",
			interval: &interval{
				start: 0,
				end:   0,
			},
		},
		{
			tokenID: "token1",
			interval: &interval{
				start: 1,
				end:   1,
			},
		},
		{
			tokenID: "token3",
			interval: &interval{
				start: 0,
				end:   0,
			},
		},
	}

	// Nonces/tx = 1
	txsData := toTxData(t, tokensIntervals, 1)
	expectedTxsData := [][]byte{
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@04@04"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@05@05"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@06@06"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@07@07"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@08@08"),
		[]byte("ESDTDeleteMetadata@746f6b656e32@01@01@01"),
		[]byte("ESDTDeleteMetadata@746f6b656e32@01@02@02"),
		[]byte("ESDTDeleteMetadata@746f6b656e32@01@03@03"),
		[]byte("ESDTDeleteMetadata@746f6b656e32@01@04@04"),
		[]byte("ESDTDeleteMetadata@746f6b656e32@01@05@05"),
		[]byte("ESDTDeleteMetadata@746f6b656e33@01@01@01"),
		[]byte("ESDTDeleteMetadata@746f6b656e33@01@02@02"),
		[]byte("ESDTDeleteMetadata@746f6b656e33@01@03@03"),
		[]byte("ESDTDeleteMetadata@746f6b656e33@01@04@04"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@02@02"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@03@03"),
		[]byte("ESDTDeleteMetadata@746f6b656e33@01@06@06"),
		[]byte("ESDTDeleteMetadata@746f6b656e33@01@07@07"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@00@00"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@01@01"),
		[]byte("ESDTDeleteMetadata@746f6b656e33@01@00@00"),
	}
	require.Equal(t, expectedTxsData, txsData)

	// Nonces/tx = 2
	txsData = toTxData(t, tokensIntervals, 2)
	expectedTxsData = [][]byte{
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@04@05"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@06@07"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@08@08@746f6b656e32@01@01@01"),
		[]byte("ESDTDeleteMetadata@746f6b656e32@01@02@03"),
		[]byte("ESDTDeleteMetadata@746f6b656e32@01@04@05"),
		[]byte("ESDTDeleteMetadata@746f6b656e33@01@01@02"),
		[]byte("ESDTDeleteMetadata@746f6b656e33@01@03@04"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@02@03"),
		[]byte("ESDTDeleteMetadata@746f6b656e33@01@06@07"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@02@00@00@01@01"),
		[]byte("ESDTDeleteMetadata@746f6b656e33@01@00@00"),
	}
	require.Equal(t, expectedTxsData, txsData)

	// Nonces/tx = 3
	txsData = toTxData(t, tokensIntervals, 3)
	expectedTxsData = [][]byte{
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@04@06"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@07@08@746f6b656e32@01@01@01"),
		[]byte("ESDTDeleteMetadata@746f6b656e32@01@02@04"),
		[]byte("ESDTDeleteMetadata@746f6b656e32@01@05@05@746f6b656e33@01@01@02"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@02@02@746f6b656e33@01@03@04"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@03@03@746f6b656e33@01@06@07"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@02@00@00@01@01@746f6b656e33@01@00@00"),
	}
	require.Equal(t, expectedTxsData, txsData)

	// Nonces/tx = 4
	txsData = toTxData(t, tokensIntervals, 4)
	expectedTxsData = [][]byte{
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@04@07"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@08@08@746f6b656e32@01@01@03"),
		[]byte("ESDTDeleteMetadata@746f6b656e32@01@04@05@746f6b656e33@01@01@02"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@02@03@746f6b656e33@01@03@04"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@02@00@00@01@01@746f6b656e33@01@06@07"),
		[]byte("ESDTDeleteMetadata@746f6b656e33@01@00@00"),
	}
	require.Equal(t, expectedTxsData, txsData)

	// Nonces/tx = 5
	txsData = toTxData(t, tokensIntervals, 5)
	expectedTxsData = [][]byte{
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@04@08"),
		[]byte("ESDTDeleteMetadata@746f6b656e32@01@01@05"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@02@02@746f6b656e33@01@01@04"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@03@03@03@00@00@01@01@746f6b656e33@01@06@07"),
		[]byte("ESDTDeleteMetadata@746f6b656e33@01@00@00"),
	}
	require.Equal(t, expectedTxsData, txsData)

	// Nonces/tx = 6
	txsData = toTxData(t, tokensIntervals, 6)
	expectedTxsData = [][]byte{
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@04@08@746f6b656e32@01@01@01"),
		[]byte("ESDTDeleteMetadata@746f6b656e32@01@02@05@746f6b656e33@01@01@02"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@02@03@746f6b656e33@02@03@04@06@07"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@02@00@00@01@01@746f6b656e33@01@00@00"),
	}
	require.Equal(t, expectedTxsData, txsData)

	// Nonces/tx = 7
	txsData = toTxData(t, tokensIntervals, 7)
	expectedTxsData = [][]byte{
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@04@08@746f6b656e32@01@01@02"),
		[]byte("ESDTDeleteMetadata@746f6b656e32@01@03@05@746f6b656e33@01@01@04"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@03@02@03@00@00@01@01@746f6b656e33@02@06@07@00@00"),
	}
	require.Equal(t, expectedTxsData, txsData)

	// Nonces/tx = 8
	txsData = toTxData(t, tokensIntervals, 8)
	expectedTxsData = [][]byte{
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@04@08@746f6b656e32@01@01@03"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@01@02@03@746f6b656e32@01@04@05@746f6b656e33@01@01@04"),
		[]byte("ESDTDeleteMetadata@746f6b656e31@02@00@00@01@01@746f6b656e33@02@06@07@00@00"),
	}
	require.Equal(t, expectedTxsData, txsData)

	// Nonces/tx = 20
	txsData = toTxData(t, tokensIntervals, 20)
	expectedTxsData = [][]byte{
		[]byte("ESDTDeleteMetadata@746f6b656e31@04@04@08@02@03@00@00@01@01@746f6b656e32@01@01@05@746f6b656e33@02@01@04@06@07"),
		[]byte("ESDTDeleteMetadata@746f6b656e33@01@00@00"),
	}
	require.Equal(t, expectedTxsData, txsData)

	// Nonces/tx >= 21
	expectedTxsData = [][]byte{
		[]byte("ESDTDeleteMetadata@746f6b656e31@04@04@08@02@03@00@00@01@01@746f6b656e32@01@01@05@746f6b656e33@03@01@04@06@07@00@00"),
	}
	for i := uint64(21); i < 100; i++ {
		txsData = toTxData(t, tokensIntervals, i)
		require.Equal(t, expectedTxsData, txsData)
	}

}
