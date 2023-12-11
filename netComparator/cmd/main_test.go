package main

import (
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetDifference(t *testing.T) {
	testCases := []struct {
		name     string
		txHash   string
		t1       data.TransactionOnNetwork
		t2       data.TransactionOnNetwork
		expected wrappedDifferences
	}{
		// Same transaction fields
		{
			"same transaction",
			"1",
			data.TransactionOnNetwork{
				Nonce:    1,
				Value:    "randomValue",
				Sender:   "randomSender",
				Receiver: "randomReceiver",
				GasPrice: 50000,
				GasLimit: 60000,
			},
			data.TransactionOnNetwork{
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
			data.TransactionOnNetwork{
				Nonce:    1,
				Value:    "randomValue",
				Sender:   "randomSender",
				Receiver: "randomReceiver",
				GasPrice: 50000,
				GasLimit: 60000,
			},
			data.TransactionOnNetwork{
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
