package httpClientWrapper_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/multiversx/mx-chain-tools-go/jsonToPrometheus/httpClientWrapper"
	"github.com/multiversx/mx-sdk-go/testsCommon"
	"github.com/stretchr/testify/require"
)

func TestNewHttpClientWrapper(t *testing.T) {
	t.Parallel()

	t.Run("nil http client should error", func(t *testing.T) {
		t.Parallel()

		wrapper, err := httpClientWrapper.NewHttpClientWrapper(nil)
		require.Equal(t, httpClientWrapper.ErrNilHttpClient, err)
		require.Nil(t, wrapper)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		wrapper, err := httpClientWrapper.NewHttpClientWrapper(&testsCommon.HTTPClientWrapperStub{})
		require.NoError(t, err)
		require.NotNil(t, wrapper)
	})
}

func TestHttpClientWrapper_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	wrapper, _ := httpClientWrapper.NewHttpClientWrapper(nil)
	require.True(t, wrapper.IsInterfaceNil())

	wrapper, _ = httpClientWrapper.NewHttpClientWrapper(&testsCommon.HTTPClientWrapperStub{})
	require.False(t, wrapper.IsInterfaceNil())
}

func TestHttpClientWrapper_GetValidatorStatistics(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	t.Run("http client error should error", func(t *testing.T) {
		t.Parallel()

		wrapper, err := httpClientWrapper.NewHttpClientWrapper(&testsCommon.HTTPClientWrapperStub{
			GetHTTPCalled: func(ctx context.Context, endpoint string) ([]byte, int, error) {
				return nil, 0, expectedErr
			},
		})
		require.NoError(t, err)

		statistics, err := wrapper.GetValidatorStatistics(context.Background())
		require.True(t, errors.Is(err, expectedErr))
		require.Nil(t, statistics)
	})
	t.Run("http client status not ok should error", func(t *testing.T) {
		t.Parallel()

		wrapper, err := httpClientWrapper.NewHttpClientWrapper(&testsCommon.HTTPClientWrapperStub{
			GetHTTPCalled: func(ctx context.Context, endpoint string) ([]byte, int, error) {
				return nil, http.StatusBadRequest, nil
			},
		})
		require.NoError(t, err)

		statistics, err := wrapper.GetValidatorStatistics(context.Background())
		require.True(t, errors.Is(err, httpClientWrapper.ErrHTTPStatusCodeIsNotOK))
		require.Nil(t, statistics)
	})
	t.Run("http client returns empty buff should error", func(t *testing.T) {
		t.Parallel()

		wrapper, err := httpClientWrapper.NewHttpClientWrapper(&testsCommon.HTTPClientWrapperStub{
			GetHTTPCalled: func(ctx context.Context, endpoint string) ([]byte, int, error) {
				return nil, http.StatusOK, nil
			},
		})
		require.NoError(t, err)

		statistics, err := wrapper.GetValidatorStatistics(context.Background())
		require.True(t, errors.Is(err, httpClientWrapper.ErrEmptyData))
		require.Nil(t, statistics)
	})
	t.Run("unmarshal failure should error", func(t *testing.T) {
		t.Parallel()

		wrapper, err := httpClientWrapper.NewHttpClientWrapper(&testsCommon.HTTPClientWrapperStub{
			GetHTTPCalled: func(ctx context.Context, endpoint string) ([]byte, int, error) {
				return []byte("this is not a valid response"), http.StatusOK, nil
			},
		})
		require.NoError(t, err)

		statistics, err := wrapper.GetValidatorStatistics(context.Background())
		require.Error(t, err)
		require.Nil(t, statistics)
	})
	t.Run("empty statistics should error", func(t *testing.T) {
		t.Parallel()

		response := httpClientWrapper.ValidatorStatisticsApiResponse{
			Data: struct {
				Statistics map[string]*httpClientWrapper.ValidatorStatistics `json:"statistics"`
			}{
				Statistics: nil,
			},
		}
		buff, _ := json.Marshal(response)
		wrapper, err := httpClientWrapper.NewHttpClientWrapper(&testsCommon.HTTPClientWrapperStub{
			GetHTTPCalled: func(ctx context.Context, endpoint string) ([]byte, int, error) {
				return buff, http.StatusOK, nil
			},
		})
		require.NoError(t, err)

		statistics, err := wrapper.GetValidatorStatistics(context.Background())
		require.True(t, errors.Is(err, httpClientWrapper.ErrEmptyData))
		require.Nil(t, statistics)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedStatistics := map[string]*httpClientWrapper.ValidatorStatistics{
			"key1": {
				TempRating:                         1,
				NumLeaderSuccess:                   2,
				NumLeaderFailure:                   3,
				NumValidatorSuccess:                4,
				NumValidatorFailure:                5,
				NumValidatorIgnoredSignatures:      6,
				Rating:                             7,
				RatingModifier:                     8,
				TotalNumLeaderSuccess:              9,
				TotalNumLeaderFailure:              10,
				TotalNumValidatorSuccess:           11,
				TotalNumValidatorFailure:           12,
				TotalNumValidatorIgnoredSignatures: 13,
				ShardId:                            0,
				ValidatorStatus:                    "eligible",
			},
		}
		response := httpClientWrapper.ValidatorStatisticsApiResponse{
			Data: struct {
				Statistics map[string]*httpClientWrapper.ValidatorStatistics `json:"statistics"`
			}{
				Statistics: providedStatistics,
			},
		}
		buff, _ := json.Marshal(response)
		wrapper, err := httpClientWrapper.NewHttpClientWrapper(&testsCommon.HTTPClientWrapperStub{
			GetHTTPCalled: func(ctx context.Context, endpoint string) ([]byte, int, error) {
				return buff, http.StatusOK, nil
			},
		})
		require.NoError(t, err)

		statistics, err := wrapper.GetValidatorStatistics(context.Background())
		require.NoError(t, err)
		require.Equal(t, providedStatistics, statistics)
	})
}
