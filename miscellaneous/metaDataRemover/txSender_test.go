package main

import (
	"context"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	"github.com/ElrondNetwork/elrond-tools-go/miscellaneous/metaDataRemover/mocks"
	"github.com/stretchr/testify/require"
	"testing"
)

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
