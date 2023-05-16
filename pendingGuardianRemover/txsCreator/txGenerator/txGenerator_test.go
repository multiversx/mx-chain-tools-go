package txGenerator

import (
	"context"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-tools-go/pendingGuardianRemover"
	"github.com/multiversx/mx-chain-tools-go/pendingGuardianRemover/mock"
	"github.com/multiversx/mx-chain-tools-go/pendingGuardianRemover/txsCreator/config"
	"github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/multiversx/mx-sdk-go/testsCommon"
	"github.com/stretchr/testify/require"
)

var (
	expectedErr = errors.New("expected error")

	userCryptoHolder = &testsCommon.CryptoComponentsHolderStub{
		GetPublicKeyCalled: func() crypto.PublicKey {
			return &testsCommon.PublicKeyStub{
				ToByteArrayCalled: func() ([]byte, error) {
					return []byte("provided user"), nil
				},
			}
		},
		GetBech32Called: func() string {
			return "provided user"
		},
	}
	guardianCryptoHolder = &testsCommon.CryptoComponentsHolderStub{
		GetPublicKeyCalled: func() crypto.PublicKey {
			return &testsCommon.PublicKeyStub{
				ToByteArrayCalled: func() ([]byte, error) {
					return []byte("provided guardian"), nil
				},
			}
		},
		GetBech32Called: func() string {
			return "provided guardian"
		},
	}
)

func TestNewTxGenerator(t *testing.T) {
	t.Parallel()

	t.Run("nil txBuilder", func(t *testing.T) {
		t.Parallel()

		gen, err := NewTxGenerator(
			nil,
			&testsCommon.CryptoComponentsHolderStub{},
			&testsCommon.CryptoComponentsHolderStub{},
			&mock.HttpClientWrapperStub{},
			config.TxGeneratorConfig{},
		)
		require.Equal(t, pendingGuardianRemover.ErrNilTXBuilder, err)
		require.Nil(t, gen)
	})
	t.Run("nil userCryptoHolder", func(t *testing.T) {
		t.Parallel()

		gen, err := NewTxGenerator(
			&mock.TxBuilderStub{},
			nil,
			&testsCommon.CryptoComponentsHolderStub{},
			&mock.HttpClientWrapperStub{},
			config.TxGeneratorConfig{},
		)
		require.True(t, errors.Is(err, pendingGuardianRemover.ErrNilCryptoHolder))
		require.Contains(t, err.Error(), "user")
		require.Nil(t, gen)
	})
	t.Run("nil guardianCryptoHolder", func(t *testing.T) {
		t.Parallel()

		gen, err := NewTxGenerator(
			&mock.TxBuilderStub{},
			&testsCommon.CryptoComponentsHolderStub{},
			nil,
			&mock.HttpClientWrapperStub{},
			config.TxGeneratorConfig{},
		)
		require.True(t, errors.Is(err, pendingGuardianRemover.ErrNilCryptoHolder))
		require.Contains(t, err.Error(), "guardian")
		require.Nil(t, gen)
	})
	t.Run("nil httpClient", func(t *testing.T) {
		t.Parallel()

		gen, err := NewTxGenerator(
			&mock.TxBuilderStub{},
			&testsCommon.CryptoComponentsHolderStub{},
			&testsCommon.CryptoComponentsHolderStub{},
			nil,
			config.TxGeneratorConfig{},
		)
		require.Equal(t, pendingGuardianRemover.ErrNilHttpClient, err)
		require.Nil(t, gen)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		gen, err := NewTxGenerator(
			&mock.TxBuilderStub{},
			&testsCommon.CryptoComponentsHolderStub{},
			&testsCommon.CryptoComponentsHolderStub{},
			&mock.HttpClientWrapperStub{},
			config.TxGeneratorConfig{},
		)
		require.NoError(t, err)
		require.NotNil(t, gen)
	})
}

func TestTxGenerator_GenerateTxs(t *testing.T) {
	t.Parallel()

	t.Run("GetAccount error should error", func(t *testing.T) {
		t.Parallel()

		gen, _ := NewTxGenerator(
			&mock.TxBuilderStub{},
			&testsCommon.CryptoComponentsHolderStub{},
			&testsCommon.CryptoComponentsHolderStub{},
			&mock.HttpClientWrapperStub{
				GetAccountCalled: func(ctx context.Context, address string) (*data.Account, error) {
					return nil, expectedErr
				},
			},
			config.TxGeneratorConfig{},
		)
		require.NotNil(t, gen)

		txs, err := gen.GenerateTxs()
		require.Equal(t, expectedErr, err)
		require.Nil(t, txs)
	})
	t.Run("ApplyUserSignatureAndGenerateTx error should error", func(t *testing.T) {
		t.Parallel()

		gen, _ := NewTxGenerator(
			&mock.TxBuilderStub{
				ApplyUserSignatureAndGenerateTxCalled: func(cryptoHolder core.CryptoComponentsHolder, arg data.ArgCreateTransaction) (*data.Transaction, error) {
					return nil, expectedErr
				},
			},
			userCryptoHolder,
			guardianCryptoHolder,
			&mock.HttpClientWrapperStub{},
			config.TxGeneratorConfig{
				NumOfTransactions: 10,
			},
		)
		require.NotNil(t, gen)

		txs, err := gen.GenerateTxs()
		require.Equal(t, expectedErr, err)
		require.Nil(t, txs)
	})
	t.Run("ApplyGuardianSignature error should error", func(t *testing.T) {
		t.Parallel()

		gen, _ := NewTxGenerator(
			&mock.TxBuilderStub{
				ApplyGuardianSignatureCalled: func(cryptoHolderGuardian core.CryptoComponentsHolder, tx *data.Transaction) error {
					return expectedErr
				},
			},
			userCryptoHolder,
			guardianCryptoHolder,
			&mock.HttpClientWrapperStub{},
			config.TxGeneratorConfig{
				NumOfTransactions: 10,
			},
		)
		require.NotNil(t, gen)

		txs, err := gen.GenerateTxs()
		require.Equal(t, expectedErr, err)
		require.Nil(t, txs)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		userBytes, _ := guardianCryptoHolder.GetPublicKey().ToByteArray()
		expectedTxData := []byte("SetGuardian@" + hex.EncodeToString(userBytes) + "@" + hex.EncodeToString([]byte("ServiceID")))
		providedTx := &data.Transaction{
			RcvAddr:      userCryptoHolder.GetBech32(),
			SndAddr:      userCryptoHolder.GetBech32(),
			Data:         expectedTxData,
			Signature:    "provided sig",
			GuardianAddr: guardianCryptoHolder.GetBech32(),
		}
		providedGuardianSig := "provided guardian sig"
		providedNumTxs := uint64(10)
		gen, _ := NewTxGenerator(
			&mock.TxBuilderStub{
				ApplyUserSignatureAndGenerateTxCalled: func(cryptoHolder core.CryptoComponentsHolder, arg data.ArgCreateTransaction) (*data.Transaction, error) {
					return providedTx, nil
				},
				ApplyGuardianSignatureCalled: func(cryptoHolderGuardian core.CryptoComponentsHolder, tx *data.Transaction) error {
					tx.GuardianSignature = providedGuardianSig
					return nil
				},
			},
			userCryptoHolder,
			guardianCryptoHolder,
			&mock.HttpClientWrapperStub{},
			config.TxGeneratorConfig{
				NumOfTransactions: providedNumTxs,
			},
		)
		require.NotNil(t, gen)

		txs, err := gen.GenerateTxs()
		require.NoError(t, err)
		require.Equal(t, providedNumTxs, uint64(len(txs)))

		for _, tx := range txs {
			require.Equal(t, expectedTxData, tx.Data)
			require.Equal(t, providedTx.Signature, tx.Signature)
			require.Equal(t, providedGuardianSig, tx.GuardianSignature)
		}
	})
}
