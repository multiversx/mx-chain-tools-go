package collector_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-tools-go/jsonToPrometheus"
	"github.com/multiversx/mx-chain-tools-go/jsonToPrometheus/collector"
	"github.com/multiversx/mx-chain-tools-go/jsonToPrometheus/httpClientWrapper"
	"github.com/multiversx/mx-chain-tools-go/jsonToPrometheus/mock"
	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

func TestNewPrometheusCollector(t *testing.T) {
	t.Parallel()

	t.Run("nil http client should error", func(t *testing.T) {
		t.Parallel()

		promCollector, err := collector.NewPrometheusCollector(nil, "", nil)
		require.Equal(t, jsonToPrometheus.ErrNilHTTPClientWrapper, err)
		require.Nil(t, promCollector)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		promCollector, err := collector.NewPrometheusCollector(&mock.HTTPClientWrapperMock{}, "", []string{"key"})
		require.NoError(t, err)
		require.NotNil(t, promCollector)
	})
}

func TestPrometheusCollector_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	promCollector, _ := collector.NewPrometheusCollector(nil, "", nil)
	require.True(t, promCollector.IsInterfaceNil())

	promCollector, _ = collector.NewPrometheusCollector(&mock.HTTPClientWrapperMock{}, "", []string{"key"})
	require.False(t, promCollector.IsInterfaceNil())
}

func TestPrometheusCollector_Describe(t *testing.T) {
	t.Parallel()

	expectedDescriptors := []string{
		"Desc{fqName: \"tempRating_metric\", help: \"Temporary rating\", constLabels: {chain=\"\"}, variableLabels: [blsKey]}",
		"Desc{fqName: \"numLeaderSuccess_metric\", help: \"Num leader success\", constLabels: {chain=\"\"}, variableLabels: [blsKey]}",
		"Desc{fqName: \"numLeaderFailure_metric\", help: \"Num leader failure\", constLabels: {chain=\"\"}, variableLabels: [blsKey]}",
		"Desc{fqName: \"numValidatorSuccess_metric\", help: \"Num validator success\", constLabels: {chain=\"\"}, variableLabels: [blsKey]}",
		"Desc{fqName: \"numValidatorFailure_metric\", help: \"Num validator failure\", constLabels: {chain=\"\"}, variableLabels: [blsKey]}",
		"Desc{fqName: \"numValidatorIgnoredSignatures_metric\", help: \"Num validator ignored signatures\", constLabels: {chain=\"\"}, variableLabels: [blsKey]}",
		"Desc{fqName: \"rating_metric\", help: \"Rating\", constLabels: {chain=\"\"}, variableLabels: [blsKey]}",
		"Desc{fqName: \"ratingModifier_metric\", help: \"Rating modifier\", constLabels: {chain=\"\"}, variableLabels: [blsKey]}",
		"Desc{fqName: \"totalNumLeaderSuccess_metric\", help: \"Total num leader success\", constLabels: {chain=\"\"}, variableLabels: [blsKey]}",
		"Desc{fqName: \"totalNumLeaderFailure_metric\", help: \"Total num leader failure\", constLabels: {chain=\"\"}, variableLabels: [blsKey]}",
		"Desc{fqName: \"totalNumValidatorSuccess_metric\", help: \"Total num validator success\", constLabels: {chain=\"\"}, variableLabels: [blsKey]}",
		"Desc{fqName: \"totalNumValidatorFailure_metric\", help: \"Total num validator failure\", constLabels: {chain=\"\"}, variableLabels: [blsKey]}",
		"Desc{fqName: \"totalNumValidatorIgnoredSignatures_metric\", help: \"Total num validator ignored signatures\", constLabels: {chain=\"\"}, variableLabels: [blsKey]}",
		"Desc{fqName: \"shardId_metric\", help: \"Shard ID\", constLabels: {chain=\"\"}, variableLabels: [blsKey]}",
		"Desc{fqName: \"validatorStatus_metric\", help: \"Validator status\", constLabels: {chain=\"\"}, variableLabels: [blsKey status]}",
	}

	promCollector, err := collector.NewPrometheusCollector(&mock.HTTPClientWrapperMock{}, "", []string{"key"})
	require.NoError(t, err)

	descriptorsChan := make(chan *prometheus.Desc)
	receivedDescriptors := make([]string, 0)
	mut := sync.RWMutex{}
	go func() {
		for {
			select {
			case descriptor := <-descriptorsChan:
				mut.Lock()
				receivedDescriptors = append(receivedDescriptors, descriptor.String())
				mut.Unlock()
			case <-time.After(time.Second):
				return
			}
		}
	}()

	promCollector.Describe(descriptorsChan)

	time.Sleep(time.Second)

	mut.RLock()
	require.True(t, len(receivedDescriptors) > 0)
	require.Equal(t, expectedDescriptors, receivedDescriptors)
	mut.RUnlock()
}

func TestPrometheusCollector_Collect(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	t.Run("http client returns error should early exit", func(t *testing.T) {
		t.Parallel()

		httpClient := &mock.HTTPClientWrapperMock{
			GetValidatorStatisticsCalled: func(ctx context.Context) (map[string]*httpClientWrapper.ValidatorStatistics, error) {
				return nil, expectedErr
			},
		}

		promCollector, err := collector.NewPrometheusCollector(httpClient, "", []string{"key"})
		require.NoError(t, err)

		metricsChan := make(chan prometheus.Metric)
		promCollector.Collect(metricsChan)
	})
	t.Run("should work for one key", func(t *testing.T) {
		t.Parallel()

		providedChain := "chain"
		providedExistingKey := "key1"
		expectedValues := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 0, 0}
		providedStatus := "eligible"
		httpClient := &mock.HTTPClientWrapperMock{
			GetValidatorStatisticsCalled: func(ctx context.Context) (map[string]*httpClientWrapper.ValidatorStatistics, error) {
				return map[string]*httpClientWrapper.ValidatorStatistics{
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
						ValidatorStatus:                    providedStatus,
					},
					"key3": {
						TempRating:                         14,
						NumLeaderSuccess:                   15,
						NumLeaderFailure:                   16,
						NumValidatorSuccess:                17,
						NumValidatorFailure:                18,
						NumValidatorIgnoredSignatures:      19,
						Rating:                             20,
						RatingModifier:                     21,
						TotalNumLeaderSuccess:              22,
						TotalNumLeaderFailure:              23,
						TotalNumValidatorSuccess:           24,
						TotalNumValidatorFailure:           25,
						TotalNumValidatorIgnoredSignatures: 26,
						ShardId:                            0,
						ValidatorStatus:                    providedStatus,
					},
				}, nil
			},
		}
		promCollector, err := collector.NewPrometheusCollector(httpClient, providedChain, []string{providedExistingKey, "key2"})
		require.NoError(t, err)

		metricsChan := make(chan prometheus.Metric)
		cnt := uint32(0)
		go func() {
			for {
				select {
				case metric := <-metricsChan:
					dtoMetric := &io_prometheus_client.Metric{}
					err = metric.Write(dtoMetric)
					require.NoError(t, err)

					require.Equal(t, providedExistingKey, *dtoMetric.Label[0].Value)
					require.Equal(t, providedChain, *dtoMetric.Label[1].Value)
					if len(dtoMetric.Label) > 2 {
						// validator status metric
						require.Equal(t, providedStatus, *dtoMetric.Label[2].Value)
					}
					require.Equal(t, expectedValues[cnt], *dtoMetric.Gauge.Value)
					atomic.AddUint32(&cnt, 1)
				case <-time.After(time.Second):
					return
				}
			}
		}()

		promCollector.Collect(metricsChan)

		time.Sleep(time.Second)

		require.True(t, atomic.LoadUint32(&cnt) == 15)
	})
	t.Run("should work for all keys", func(t *testing.T) {
		t.Parallel()

		providedChain := "chain"
		expectedValues := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 0, 0, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 0, 0}
		providedStatus := "eligible"
		httpClient := &mock.HTTPClientWrapperMock{
			GetValidatorStatisticsCalled: func(ctx context.Context) (map[string]*httpClientWrapper.ValidatorStatistics, error) {
				return map[string]*httpClientWrapper.ValidatorStatistics{
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
						ValidatorStatus:                    providedStatus,
					},
					"key2": {
						TempRating:                         14,
						NumLeaderSuccess:                   15,
						NumLeaderFailure:                   16,
						NumValidatorSuccess:                17,
						NumValidatorFailure:                18,
						NumValidatorIgnoredSignatures:      19,
						Rating:                             20,
						RatingModifier:                     21,
						TotalNumLeaderSuccess:              22,
						TotalNumLeaderFailure:              23,
						TotalNumValidatorSuccess:           24,
						TotalNumValidatorFailure:           25,
						TotalNumValidatorIgnoredSignatures: 26,
						ShardId:                            0,
						ValidatorStatus:                    providedStatus,
					},
				}, nil
			},
		}
		promCollector, err := collector.NewPrometheusCollector(httpClient, providedChain, []string{})
		require.NoError(t, err)

		metricsChan := make(chan prometheus.Metric)
		cnt := uint32(0)
		go func() {
			for {
				select {
				case metric := <-metricsChan:
					dtoMetric := &io_prometheus_client.Metric{}
					err = metric.Write(dtoMetric)
					require.NoError(t, err)

					require.Equal(t, providedChain, *dtoMetric.Label[1].Value)
					if len(dtoMetric.Label) > 2 {
						// validator status metric
						require.Equal(t, providedStatus, *dtoMetric.Label[2].Value)
					}
					require.Equal(t, expectedValues[cnt], *dtoMetric.Gauge.Value)
					atomic.AddUint32(&cnt, 1)
				case <-time.After(time.Second):
					return
				}
			}
		}()

		promCollector.Collect(metricsChan)

		time.Sleep(time.Second)

		require.True(t, atomic.LoadUint32(&cnt) == 30)
	})
}
