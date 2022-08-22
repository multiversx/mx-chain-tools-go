package main

import (
	"context"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	"github.com/ElrondNetwork/elrond-tools-go/miscellaneous/metaDataRemover/mocks"
	"github.com/stretchr/testify/require"
	"testing"
)

func requireSameSliceDifferentOrder(t *testing.T, s1, s2 [][]byte) {
	require.Equal(t, len(s1), len(s2))

	for _, elemInS1 := range s1 {
		require.Contains(t, s2, elemInS1)
	}
}

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

func TestSendTxs(t *testing.T) {
	networkConfig := &data.NetworkConfig{
		ChainID:     "chainID",
		MinGasLimit: 2000,
		MinGasPrice: 1000,
	}
	proxy := &mocks.ProxyStub{
		GetNetworkConfigCalled: func(ctx context.Context) (*data.NetworkConfig, error) {
			return networkConfig, nil
		},

		GetDefaultTransactionArgumentsCalled: func(ctx context.Context, address core.AddressHandler, networkConfigs *data.NetworkConfig) (data.ArgCreateTransaction, error) {
			require.Equal(t, networkConfig, networkConfigs)

			return data.ArgCreateTransaction{
				Nonce:    4,
				SndAddr:  "erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th",
				GasPrice: networkConfig.MinGasPrice,
				GasLimit: networkConfig.MinGasLimit,
				ChainID:  networkConfig.ChainID,
			}, nil
		},
	}

	txIndex := 0
	bulkSize := 2
	currTx := &data.Transaction{}
	expectedTxsData := [][]byte{
		[]byte("ESDTDeleteMetadata@746f6b656e31@02@00@00@01@01"), // token1: 2 intervals: [0,0];[1,1]
		[]byte("ESDTDeleteMetadata@746f6b656e31@02@02@03@04@08"), // token1: 2 intervals: [2,3];[4,8]
		[]byte("ESDTDeleteMetadata@746f6b656e32@01@01@05"),       // token2: 1 interval:  [1,5]
		[]byte("ESDTDeleteMetadata@746f6b656e33@02@00@00@01@04"), // token3: 2 intervals: [0,0];[1,4]
		[]byte("ESDTDeleteMetadata@746f6b656e33@01@05@06"),       // token3: 1 interval:  [5,6]
	}
	expectedArgTx := data.ArgCreateTransaction{
		Nonce:    4,
		Value:    "0",
		SndAddr:  "erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th",
		RcvAddr:  "erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th",
		GasPrice: networkConfig.MinGasPrice,
		GasLimit: networkConfig.MinGasLimit,
		ChainID:  networkConfig.ChainID,
	}
	txInteractor := &mocks.TransactionInteractorStub{
		ApplySignatureAndGenerateTxCalled: func(skBytes []byte, arg data.ArgCreateTransaction) (*data.Transaction, error) {
			expectedArgTx.Nonce++
			expectedArgTx.Data = expectedTxsData[txIndex]
			require.Equal(t, expectedArgTx, arg)

			currTx = &data.Transaction{
				Nonce:    expectedArgTx.Nonce,
				Value:    "0",
				RcvAddr:  "erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th",
				SndAddr:  "erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th",
				GasPrice: arg.GasPrice,
				GasLimit: arg.GasLimit,
				Data:     arg.Data,
				ChainID:  arg.ChainID,
			}

			return currTx, nil
		},
		AddTransactionCalled: func(tx *data.Transaction) {
			require.Equal(t, currTx, tx)
			txIndex++
		},
		SendTransactionsAsBunchCalled: func(ctx context.Context, bunchSize int) ([]string, error) {
			switch txIndex {
			case 2:
				require.Equal(t, bulkSize, bunchSize)
				return []string{"hash1", "hash2"}, nil
			case 4:
				require.Equal(t, bulkSize, bunchSize)
				return []string{"hash3", "hash4"}, nil
			case 5:
				require.Equal(t, 1, bunchSize)
				return []string{"hash5"}, nil
			default:
				require.Fail(t, "should not have sent another bulk; bulk size = 2; we have 5 txs to send; first 2 bulks have 2 txs; remaining bulk has 1 tx")
				return nil, nil
			}
		},
	}

	err := sendTxs("alice.pem", proxy, txInteractor, expectedTxsData, bulkSize)
	require.Nil(t, err)
	require.Equal(t, len(expectedTxsData), txIndex)
}

func TestCreateTxData2(t *testing.T) {
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
	txsData, err := createTxsData2(tokensIntervals, 1)
	require.Nil(t, err)
	expectedTxsData := [][]byte{
		[]byte("ESDTDeleteMetadata@token1@01@00@00"),
		[]byte("ESDTDeleteMetadata@token1@01@01@01"),
		[]byte("ESDTDeleteMetadata@token1@01@02@02"),
		[]byte("ESDTDeleteMetadata@token1@01@04@04"),
		[]byte("ESDTDeleteMetadata@token1@01@03@03"),
		[]byte("ESDTDeleteMetadata@token1@01@05@05"),
		[]byte("ESDTDeleteMetadata@token1@01@06@06"),
		[]byte("ESDTDeleteMetadata@token1@01@07@07"),
		[]byte("ESDTDeleteMetadata@token1@01@08@08"),
		[]byte("ESDTDeleteMetadata@token2@01@01@01"),
		[]byte("ESDTDeleteMetadata@token2@01@02@02"),
		[]byte("ESDTDeleteMetadata@token2@01@03@03"),
		[]byte("ESDTDeleteMetadata@token2@01@04@04"),
		[]byte("ESDTDeleteMetadata@token2@01@05@05"),
		[]byte("ESDTDeleteMetadata@token3@01@00@00"),
		[]byte("ESDTDeleteMetadata@token3@01@01@01"),
		[]byte("ESDTDeleteMetadata@token3@01@06@06"),
		[]byte("ESDTDeleteMetadata@token3@01@02@02"),
		[]byte("ESDTDeleteMetadata@token3@01@07@07"),
		[]byte("ESDTDeleteMetadata@token3@01@03@03"),
		[]byte("ESDTDeleteMetadata@token3@01@04@04"),
	}
	require.Equal(t, expectedTxsData, txsData)

	// Tokens/tx = 2
	txsData, err = createTxsData2(tokensIntervals, 2)
	require.Nil(t, err)
	expectedTxsData = [][]byte{
		[]byte("ESDTDeleteMetadata@token1@02@00@00@01@01"),
		[]byte("ESDTDeleteMetadata@token1@01@02@03"),
		[]byte("ESDTDeleteMetadata@token1@01@04@05"),
		[]byte("ESDTDeleteMetadata@token1@01@06@07"),
		[]byte("ESDTDeleteMetadata@token1@01@08@08@token2@01@01@01"),
		[]byte("ESDTDeleteMetadata@token2@01@02@03"),
		[]byte("ESDTDeleteMetadata@token2@01@04@05"),
		[]byte("ESDTDeleteMetadata@token3@02@00@00@01@01"),
		[]byte("ESDTDeleteMetadata@token3@01@06@07"),
		[]byte("ESDTDeleteMetadata@token3@01@02@03"),
		[]byte("ESDTDeleteMetadata@token3@01@04@04"),
	}
	require.Equal(t, expectedTxsData, txsData)

	// Tokens/tx = 3
	txsData, err = createTxsData2(tokensIntervals, 3)
	require.Nil(t, err)
	expectedTxsData = [][]byte{
		[]byte("ESDTDeleteMetadata@token1@03@00@00@01@01@02@02"),
		[]byte("ESDTDeleteMetadata@token1@01@04@06"),
		[]byte("ESDTDeleteMetadata@token1@02@03@03@07@08"),
		[]byte("ESDTDeleteMetadata@token2@01@01@03"),
		[]byte("ESDTDeleteMetadata@token2@01@04@05@token3@01@00@00"),
		[]byte("ESDTDeleteMetadata@token3@01@01@03"),
		[]byte("ESDTDeleteMetadata@token3@02@06@07@04@04"),
	}
	require.Equal(t, expectedTxsData, txsData)

	// Tokens/tx = 4
	txsData, err = createTxsData2(tokensIntervals, 4)
	require.Nil(t, err)
	expectedTxsData = [][]byte{
		[]byte("ESDTDeleteMetadata@token1@03@00@00@01@01@02@03"),
		[]byte("ESDTDeleteMetadata@token1@01@04@07"),
		[]byte("ESDTDeleteMetadata@token1@01@08@08@token2@01@01@03"),
		[]byte("ESDTDeleteMetadata@token2@01@04@05@token3@02@00@00@01@01"),
		[]byte("ESDTDeleteMetadata@token3@02@06@07@02@03"),
		[]byte("ESDTDeleteMetadata@token3@01@04@04"),
	}
	require.Equal(t, expectedTxsData, txsData)

	// Tokens/tx = 5
	txsData, err = createTxsData2(tokensIntervals, 5)
	require.Nil(t, err)
	expectedTxsData = [][]byte{
		[]byte("ESDTDeleteMetadata@token1@04@00@00@01@01@02@03@04@04"),
		[]byte("ESDTDeleteMetadata@token1@01@05@08@token2@01@01@01"),
		[]byte("ESDTDeleteMetadata@token2@01@02@05@token3@01@00@00"),
		[]byte("ESDTDeleteMetadata@token3@02@01@04@06@06"),
		[]byte("ESDTDeleteMetadata@token3@01@07@07"),
	}
	require.Equal(t, expectedTxsData, txsData)

	// Tokens/tx = 6
	txsData, err = createTxsData2(tokensIntervals, 6)
	require.Nil(t, err)
	expectedTxsData = [][]byte{
		[]byte("ESDTDeleteMetadata@token1@04@00@00@01@01@02@03@04@05"),
		[]byte("ESDTDeleteMetadata@token1@01@06@08@token2@01@01@03"),
		[]byte("ESDTDeleteMetadata@token2@01@04@05@token3@02@00@00@01@03"),
		[]byte("ESDTDeleteMetadata@token3@02@06@07@04@04"),
	}
	require.Equal(t, expectedTxsData, txsData)

	// Tokens/tx = 7
	txsData, err = createTxsData2(tokensIntervals, 7)
	require.Nil(t, err)
	expectedTxsData = [][]byte{
		[]byte("ESDTDeleteMetadata@token1@04@00@00@01@01@02@03@04@06"),
		[]byte("ESDTDeleteMetadata@token1@01@07@08@token2@01@01@05"),
		[]byte("ESDTDeleteMetadata@token3@03@00@00@01@04@06@07"),
	}
	require.Equal(t, expectedTxsData, txsData)

	// Tokens/tx = 8
	txsData, err = createTxsData2(tokensIntervals, 8)
	require.Nil(t, err)
	expectedTxsData = [][]byte{
		[]byte("ESDTDeleteMetadata@token1@04@00@00@01@01@02@03@04@07"),
		[]byte("ESDTDeleteMetadata@token1@01@08@08@token2@01@01@05@token3@02@00@00@01@01"),
		[]byte("ESDTDeleteMetadata@token3@02@06@07@02@04"),
	}
	require.Equal(t, expectedTxsData, txsData)

	// Tokens/tx = 20
	txsData, err = createTxsData2(tokensIntervals, 20)
	require.Nil(t, err)
	expectedTxsData = [][]byte{
		[]byte("ESDTDeleteMetadata@token1@04@00@00@01@01@02@03@04@08@token2@01@01@05@token3@03@00@00@01@04@06@06"),
		[]byte("ESDTDeleteMetadata@token3@01@07@07"),
	}
	require.Equal(t, expectedTxsData, txsData)

	// Tokens/tx >= 21
	expectedTxsData = [][]byte{
		[]byte("ESDTDeleteMetadata@token1@04@00@00@01@01@02@03@04@08@token2@01@01@05@token3@03@00@00@01@04@06@07"),
	}
	for i := 21; i < 100; i++ {
		txsData, err = createTxsData2(tokensIntervals, i)
		require.Nil(t, err)
		require.Equal(t, expectedTxsData, txsData)
	}
}
