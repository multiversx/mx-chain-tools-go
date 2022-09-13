package main

import (
	"context"
	"encoding/hex"
	"testing"

	"github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	"github.com/ElrondNetwork/elrond-tools-go/tokensRemover/metaDataRemover/mocks"
	"github.com/stretchr/testify/require"
)

func TestNewTxCreator(t *testing.T) {
	t.Parallel()

	t.Run("nil proxy, should err", func(t *testing.T) {
		t.Parallel()

		txc, err := newTxCreator(nil, &mocks.TransactionInteractorStub{})
		require.Nil(t, txc)
		require.Equal(t, errNilProxy, err)
	})

	t.Run("nil tx interactor, should err", func(t *testing.T) {
		t.Parallel()

		txc, err := newTxCreator(&mocks.ProxyStub{}, nil)
		require.Nil(t, txc)
		require.Equal(t, errNilTxInteractor, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		txc, err := newTxCreator(&mocks.ProxyStub{}, &mocks.TransactionInteractorStub{})
		require.Nil(t, err)
		require.NotNil(t, txc)
	})
}

func TestTxCreator_CreateTxs(t *testing.T) {
	t.Parallel()

	addr, err := data.NewAddressFromBech32String("erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th")
	require.Nil(t, err)
	sk, err := hex.DecodeString("413f42575f7f26fad3317a778771212fdb80245850981e48b58a4f25e344e8f9")
	require.Nil(t, err)
	pemData := &skAddress{
		secretKey: sk,
		address:   addr,
	}

	txData1 := []byte("txData1")
	txData2 := []byte("txData2")
	txsData := [][]byte{txData1, txData2}

	nonce := uint64(4)
	gasLimit := uint64(44444)

	txs := []*data.Transaction{
		{
			Nonce:     nonce,
			Signature: "signature1",
		},
		{
			Nonce:     nonce + 1,
			Signature: "signature2",
		},
	}

	networkCfg := &data.NetworkConfig{
		ChainID:     "1",
		MinGasPrice: 100,
	}

	proxy := &mocks.ProxyStub{
		GetNetworkConfigCalled: func(ctx context.Context) (*data.NetworkConfig, error) {
			return networkCfg, nil
		},

		GetDefaultTransactionArgumentsCalled: func(ctx context.Context, address core.AddressHandler, networkConfigs *data.NetworkConfig) (data.ArgCreateTransaction, error) {
			require.Equal(t, networkCfg, networkConfigs)
			require.Equal(t, addr, address)

			return data.ArgCreateTransaction{
				Nonce:    nonce,
				SndAddr:  addr.AddressAsBech32String(),
				ChainID:  networkCfg.ChainID,
				GasPrice: networkCfg.MinGasPrice,
			}, nil
		},
	}

	txIdx := 0
	txInteractor := &mocks.TransactionInteractorStub{
		ApplySignatureAndGenerateTxCalled: func(skBytes []byte, arg data.ArgCreateTransaction) (*data.Transaction, error) {
			require.Equal(t, sk, skBytes)
			require.Equal(t, data.ArgCreateTransaction{
				Nonce:    nonce,
				Value:    "0",
				Data:     txsData[txIdx],
				ChainID:  networkCfg.ChainID,
				GasPrice: networkCfg.MinGasPrice,
				GasLimit: gasLimit,
				SndAddr:  addr.AddressAsBech32String(),
				RcvAddr:  addr.AddressAsBech32String()}, arg)

			defer func() {
				nonce++
				txIdx++
			}()

			return txs[txIdx], nil
		},
	}

	txc, _ := newTxCreator(proxy, txInteractor)
	signedTxs, err := txc.createTxs(pemData, txsData, gasLimit)
	require.Nil(t, err)
	require.Equal(t, signedTxs, txs)
}
