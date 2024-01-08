package main

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-tools-go/netComparator/core/domain"
)

func TestGetDifference(t *testing.T) {
	testCases := []struct {
		name     string
		txHash   string
		t1       domain.Transaction
		t2       domain.Transaction
		expected wrappedDifferences
	}{
		{
			"same transaction",
			"1",
			domain.Transaction{
				Nonce:    1,
				Value:    "randomValue",
				Sender:   "randomSender",
				Receiver: "randomReceiver",
				GasPrice: 50000,
				GasLimit: 60000,
			},
			domain.Transaction{
				Nonce:    1,
				Value:    "randomValue",
				Sender:   "randomSender",
				Receiver: "randomReceiver",
				GasPrice: 50000,
				GasLimit: 60000,
			},
			wrappedDifferences{"1", nil, ""},
		},

		{
			"different transaction fields",
			"2",
			domain.Transaction{
				Nonce:    1,
				Value:    "randomValue",
				Sender:   "randomSender",
				Receiver: "randomReceiver",
				GasPrice: 50000,
				GasLimit: 60000,
			},
			domain.Transaction{
				Nonce:    13,
				Value:    "someRandomValue",
				Sender:   "someRandomSender",
				Receiver: "someRandomReceiver",
				GasPrice: 50001,
				GasLimit: 60001,
			},
			wrappedDifferences{"2", map[string][]any{
				"Nonce":    {uint64(1), uint64(13)},
				"Value":    {"randomValue", "someRandomValue"},
				"Sender":   {"randomSender", "someRandomSender"},
				"Receiver": {"randomReceiver", "someRandomReceiver"},
				"GasPrice": {uint64(50000), uint64(50001)},
				"GasLimit": {uint64(60000), uint64(60001)},
			}, ""},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			difference := getDifference(tt.txHash, tt.t1, tt.t2)
			require.Equal(t, difference, tt.expected)
		})
	}
}

func TestRetryCalculator(t *testing.T) {
	testCases := []struct {
		name     string
		n        int
		expected uint
	}{
		{
			"100 transactions",
			100,
			minimumNumberOfRetries,
		},
		{
			"500 transactions",
			500,
			minimumNumberOfRetries,
		},
		{
			"1000 transactions",
			1000,
			mediumNumberOfRetries,
		},
		{
			"2000 transactions",
			2000,
			mediumNumberOfRetries,
		},
		{
			"5000 transactions",
			5000,
			maximumNumberOfRetries,
		},
		{
			"10000 transactions",
			10000,
			maximumNumberOfRetries,
		},
		{
			"10001 transactions",
			10001,
			maximumNumberOfRetries,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateRetryAttempts(tt.n)
			require.Equal(t, got, tt.expected)
		})
	}
}

func TestCalculateBatches(t *testing.T) {
	testCases := []struct {
		name     string
		n        uint
		expected []uint
	}{
		{
			"100 transactions",
			100,
			[]uint{50, 50},
		},
		{
			"125 transactions",
			125,
			[]uint{50, 50, 25},
		},
		{
			"150 transactions",
			150,
			[]uint{50, 50, 50},
		},
		{
			"1.013 transactions",
			1_013,
			append(append(make([]uint, 0), mockExpectedBatchResult(20)...), uint(13)),
		},
		{
			"20.000 transactions",
			20_000,
			mockExpectedBatchResult(400),
		},
		{
			"100.000 transactions",
			100_000,
			mockExpectedBatchResult(2000),
		},
		{
			"100.125 transactions",
			100_125,
			append(append(make([]uint, 0), mockExpectedBatchResult(2002)...), uint(25)),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			batches := make([]uint, 0)
			calculateBatches(tt.n, &batches)
			require.Equal(t, tt.expected, batches)
		})
	}
}

func mockExpectedBatchResult(n int) []uint {
	mockBatch := make([]uint, 0)
	for i := 0; i < n; i++ {
		mockBatch = append(mockBatch, maximumNumberOfTransactionsPerBatch)
	}

	return mockBatch
}
