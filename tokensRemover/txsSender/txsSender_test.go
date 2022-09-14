package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	"github.com/ElrondNetwork/elrond-tools-go/tokensRemover/metaDataRemover/mocks"
	"github.com/stretchr/testify/require"
)

func TestTxsSender_SendTxs(t *testing.T) {
	getAccountCt := 0
	nonce := uint64(4)
	currTxIndex := 0
	txs := []*data.Transaction{
		{
			Nonce:   nonce,
			SndAddr: "erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th",
		},
		{
			Nonce:   nonce + 1,
			SndAddr: "erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th",
		},
	}
	proxy := &mocks.ProxyStub{
		GetNetworkConfigCalled: func(ctx context.Context) (*data.NetworkConfig, error) {
			return &data.NetworkConfig{
				RoundDuration: 100,
			}, nil
		},
		GetAccountCalled: func(ctx context.Context, address core.AddressHandler) (*data.Account, error) {
			defer func() {
				getAccountCt++
				nonce++
			}()

			switch getAccountCt {
			case 0, 1:
				return &data.Account{Nonce: nonce}, nil
			default:
				require.Fail(t, "should not request account anymore")
			}

			return nil, nil
		},
		SendTransactionCalled: func(ctx context.Context, tx *data.Transaction) (string, error) {
			require.Equal(t, txs[currTxIndex], tx)

			currTxIndex++
			return fmt.Sprintf("txhash%d", currTxIndex), nil
		},
	}

	ts := txsSender{
		proxy: proxy,
	}

	start := time.Now()
	err := ts.send(txs)
	elapsed := time.Since(start)

	require.Nil(t, err)
	require.True(t, elapsed < time.Second)
	require.Equal(t, 2, getAccountCt)
	require.Equal(t, 2, currTxIndex)
	require.Equal(t, uint64(6), nonce)
}
