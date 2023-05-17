package txSender

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-tools-go/pendingGuardianRemover"
	"github.com/multiversx/mx-chain-tools-go/pendingGuardianRemover/mock"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/stretchr/testify/require"
)

var expectedErr = errors.New("expected error")

func TestNewTxSender(t *testing.T) {
	t.Parallel()

	t.Run("nil httpClient should error", func(t *testing.T) {
		t.Parallel()

		sender, err := NewTxSender(nil, time.Second, make(map[uint64]*data.Transaction))
		require.Equal(t, pendingGuardianRemover.ErrNilHttpClient, err)
		require.Nil(t, sender)
	})
	t.Run("invalid delay should error", func(t *testing.T) {
		t.Parallel()

		invalidDelay := time.Second - time.Nanosecond
		sender, err := NewTxSender(&mock.HttpClientWrapperStub{}, invalidDelay, make(map[uint64]*data.Transaction))
		require.True(t, errors.Is(err, pendingGuardianRemover.ErrInvalidValue))
		require.Contains(t, err.Error(), "delayBetweenSends")
		require.Nil(t, sender)
	})
	t.Run("empty map should error", func(t *testing.T) {
		t.Parallel()

		sender, err := NewTxSender(&mock.HttpClientWrapperStub{}, time.Second, make(map[uint64]*data.Transaction))
		require.True(t, errors.Is(err, pendingGuardianRemover.ErrInvalidValue))
		require.Contains(t, err.Error(), "transactions map")
		require.Nil(t, sender)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		sender, err := NewTxSender(
			&mock.HttpClientWrapperStub{},
			time.Second,
			map[uint64]*data.Transaction{
				0: {},
			})
		require.NoError(t, err)
		require.NotNil(t, sender)
		require.Nil(t, sender.Close())
	})
	t.Run("missing pending guardian should not send tx", func(t *testing.T) {
		t.Parallel()

		sender, err := NewTxSender(
			&mock.HttpClientWrapperStub{
				GetAccountCalled: func(ctx context.Context, address string) (*data.Account, error) {
					require.Fail(t, "should have not been called")
					return nil, nil
				},
			},
			time.Second,
			map[uint64]*data.Transaction{
				0: {},
			})
		require.NoError(t, err)
		require.NotNil(t, sender)
		time.Sleep(time.Second + time.Millisecond*100) // allow one loop
		require.Nil(t, sender.Close())
	})
	t.Run("missing guardians should not send tx", func(t *testing.T) {
		t.Parallel()

		sender, err := NewTxSender(
			&mock.HttpClientWrapperStub{
				GetGuardianDataCalled: func(ctx context.Context, address string) (*api.GuardianData, error) {
					return nil, nil
				},
				GetAccountCalled: func(ctx context.Context, address string) (*data.Account, error) {
					require.Fail(t, "should have not been called")
					return nil, nil
				},
			},
			time.Second,
			map[uint64]*data.Transaction{
				0: {},
			})
		require.NoError(t, err)
		require.NotNil(t, sender)
		time.Sleep(time.Second + time.Millisecond*100) // allow one loop
		require.Nil(t, sender.Close())
	})
	t.Run("error on get nonce should not send tx", func(t *testing.T) {
		t.Parallel()

		sender, err := NewTxSender(
			&mock.HttpClientWrapperStub{
				GetGuardianDataCalled: func(ctx context.Context, address string) (*api.GuardianData, error) {
					return nil, expectedErr // for coverage, error should proceed with tx send
				},
				GetAccountCalled: func(ctx context.Context, address string) (*data.Account, error) {
					return nil, expectedErr
				},
				SendTransactionCalled: func(ctx context.Context, txBuff []byte) (string, error) {
					require.Fail(t, "should have not been called")
					return "", nil
				},
			},
			time.Second,
			map[uint64]*data.Transaction{
				0: {},
			})
		require.NoError(t, err)
		require.NotNil(t, sender)
		time.Sleep(time.Second + time.Millisecond*100) // allow one loop
		require.Nil(t, sender.Close())
	})
	t.Run("pending guardian but nonce not in map should not send tx", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		sender, err := NewTxSender(
			&mock.HttpClientWrapperStub{
				GetGuardianDataCalled: func(ctx context.Context, address string) (*api.GuardianData, error) {
					return &api.GuardianData{
						ActiveGuardian:  &api.Guardian{},
						PendingGuardian: &api.Guardian{},
						Guarded:         true,
					}, nil
				},
				GetAccountCalled: func(ctx context.Context, address string) (*data.Account, error) {
					wasCalled = true
					return &data.Account{
						Nonce: 1, // missing from map
					}, nil
				},
				SendTransactionCalled: func(ctx context.Context, txBuff []byte) (string, error) {
					require.Fail(t, "should have not been called")
					return "", nil
				},
			},
			time.Second,
			map[uint64]*data.Transaction{
				0: {},
			})
		require.NoError(t, err)
		require.NotNil(t, sender)
		time.Sleep(time.Second + time.Millisecond*100) // allow one loop
		require.True(t, wasCalled)
		require.Nil(t, sender.Close())
	})
	t.Run("different senders in map should not send tx", func(t *testing.T) {
		t.Parallel()

		calledCounter := uint64(0)
		sendCalledCounter := 0
		sender, err := NewTxSender(
			&mock.HttpClientWrapperStub{
				GetGuardianDataCalled: func(ctx context.Context, address string) (*api.GuardianData, error) {
					return &api.GuardianData{
						ActiveGuardian:  &api.Guardian{},
						PendingGuardian: &api.Guardian{},
						Guarded:         true,
					}, nil
				},
				GetAccountCalled: func(ctx context.Context, address string) (*data.Account, error) {
					calledCounter++
					return &data.Account{
						Nonce: calledCounter, // make sure it is called for both entries
					}, nil
				},
				SendTransactionCalled: func(ctx context.Context, txBuff []byte) (string, error) {
					sendCalledCounter++
					return "", nil
				},
			},
			time.Second,
			map[uint64]*data.Transaction{
				1: {
					SndAddr: "sender_1",
				},
				2: {
					SndAddr: "sender_2",
				},
			})
		require.NoError(t, err)
		require.NotNil(t, sender)
		time.Sleep(time.Second*2 + time.Millisecond*100) // allow two loops
		require.Equal(t, uint64(2), calledCounter)
		require.Equal(t, 1, sendCalledCounter)
		require.Nil(t, sender.Close())
	})
	t.Run("error on get guardian data should send tx", func(t *testing.T) {
		t.Parallel()

		providedNonce := uint64(100)
		providedTx := &data.Transaction{
			Nonce: providedNonce,
		}
		wasCalled := false
		sender, err := NewTxSender(
			&mock.HttpClientWrapperStub{
				GetGuardianDataCalled: func(ctx context.Context, address string) (*api.GuardianData, error) {
					return nil, expectedErr // error should proceed with tx send
				},
				GetAccountCalled: func(ctx context.Context, address string) (*data.Account, error) {
					return &data.Account{
						Nonce: providedNonce,
					}, nil
				},
				SendTransactionCalled: func(ctx context.Context, txBuff []byte) (string, error) {
					wasCalled = true
					return "", nil
				},
			},
			time.Second,
			map[uint64]*data.Transaction{
				providedNonce: providedTx,
			})
		require.NoError(t, err)
		require.NotNil(t, sender)
		time.Sleep(time.Second + time.Millisecond*100) // allow one loop
		require.True(t, wasCalled)
		require.Nil(t, sender.Close())
	})
}
