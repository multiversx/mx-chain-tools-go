package collector

import (
	"context"

	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-tools-go/jsonToPrometheus"
	"github.com/multiversx/mx-chain-tools-go/jsonToPrometheus/httpClientWrapper"
	"github.com/prometheus/client_golang/prometheus"
)

var log = logger.GetOrCreate("collector")

type prometheusDescriptors struct {
	tempRating                         *prometheus.Desc
	numLeaderSuccess                   *prometheus.Desc
	numLeaderFailure                   *prometheus.Desc
	numValidatorSuccess                *prometheus.Desc
	numValidatorFailure                *prometheus.Desc
	numValidatorIgnoredSignatures      *prometheus.Desc
	rating                             *prometheus.Desc
	ratingModifier                     *prometheus.Desc
	totalNumLeaderSuccess              *prometheus.Desc
	totalNumLeaderFailure              *prometheus.Desc
	totalNumValidatorSuccess           *prometheus.Desc
	totalNumValidatorFailure           *prometheus.Desc
	totalNumValidatorIgnoredSignatures *prometheus.Desc
	shardId                            *prometheus.Desc
	validatorStatus                    *prometheus.Desc
}

type prometheusCollector struct {
	httpClient  jsonToPrometheus.HttpClientWrapper
	descriptors prometheusDescriptors
	blsKeys     []string
}

// NewPrometheusCollector returns a new instance of prometheus exporter
func NewPrometheusCollector(httpClient jsonToPrometheus.HttpClientWrapper, chain string, blsKeys []string) (*prometheusCollector, error) {
	if check.IfNil(httpClient) {
		return nil, jsonToPrometheus.ErrNilHTTPClientWrapper
	}
	if len(blsKeys) == 0 {
		return nil, jsonToPrometheus.ErrNoKeyProvided
	}

	constLabels := prometheus.Labels{"chain": chain}

	return &prometheusCollector{
		httpClient: httpClient,
		blsKeys:    blsKeys,
		descriptors: prometheusDescriptors{
			tempRating:                         prometheus.NewDesc("tempRating_metric", "Temporary rating", []string{"blsKey"}, constLabels),
			numLeaderSuccess:                   prometheus.NewDesc("numLeaderSuccess_metric", "Num leader success", []string{"blsKey"}, constLabels),
			numLeaderFailure:                   prometheus.NewDesc("numLeaderFailure_metric", "Num leader failure", []string{"blsKey"}, constLabels),
			numValidatorSuccess:                prometheus.NewDesc("numValidatorSuccess_metric", "Num validator success", []string{"blsKey"}, constLabels),
			numValidatorFailure:                prometheus.NewDesc("numValidatorFailure_metric", "Num validator failure", []string{"blsKey"}, constLabels),
			numValidatorIgnoredSignatures:      prometheus.NewDesc("numValidatorIgnoredSignatures_metric", "Num validator ignored signatures", []string{"blsKey"}, constLabels),
			rating:                             prometheus.NewDesc("rating_metric", "Rating", []string{"blsKey"}, constLabels),
			ratingModifier:                     prometheus.NewDesc("ratingModifier_metric", "Rating modifier", []string{"blsKey"}, constLabels),
			totalNumLeaderSuccess:              prometheus.NewDesc("totalNumLeaderSuccess_metric", "Total num leader success", []string{"blsKey"}, constLabels),
			totalNumLeaderFailure:              prometheus.NewDesc("totalNumLeaderFailure_metric", "Total num leader failure", []string{"blsKey"}, constLabels),
			totalNumValidatorSuccess:           prometheus.NewDesc("totalNumValidatorSuccess_metric", "Total num validator success", []string{"blsKey"}, constLabels),
			totalNumValidatorFailure:           prometheus.NewDesc("totalNumValidatorFailure_metric", "Total num validator failure", []string{"blsKey"}, constLabels),
			totalNumValidatorIgnoredSignatures: prometheus.NewDesc("totalNumValidatorIgnoredSignatures_metric", "Total num validator ignored signatures", []string{"blsKey"}, constLabels),
			shardId:                            prometheus.NewDesc("shardId_metric", "Shard ID", []string{"blsKey"}, constLabels),
			validatorStatus:                    prometheus.NewDesc("validatorStatus_metric", "Validator status", []string{"blsKey", "status"}, constLabels),
		},
	}, nil
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector to the provided channel and returns once
// the last descriptor has been sent. The sent descriptors fulfill the
// consistency and uniqueness requirements described in the Desc
// documentation.
func (collector *prometheusCollector) Describe(descriptorsChan chan<- *prometheus.Desc) {
	descriptorsChan <- collector.descriptors.tempRating
	descriptorsChan <- collector.descriptors.numLeaderSuccess
	descriptorsChan <- collector.descriptors.numLeaderFailure
	descriptorsChan <- collector.descriptors.numValidatorSuccess
	descriptorsChan <- collector.descriptors.numValidatorFailure
	descriptorsChan <- collector.descriptors.numValidatorIgnoredSignatures
	descriptorsChan <- collector.descriptors.rating
	descriptorsChan <- collector.descriptors.ratingModifier
	descriptorsChan <- collector.descriptors.totalNumLeaderSuccess
	descriptorsChan <- collector.descriptors.totalNumLeaderFailure
	descriptorsChan <- collector.descriptors.totalNumValidatorSuccess
	descriptorsChan <- collector.descriptors.totalNumValidatorFailure
	descriptorsChan <- collector.descriptors.totalNumValidatorIgnoredSignatures
	descriptorsChan <- collector.descriptors.shardId
	descriptorsChan <- collector.descriptors.validatorStatus
}

// Collect is called by the Prometheus registry when collecting
// metrics. The implementation sends each collected metric via the
// provided channel and returns once the last metric has been sent. The
// descriptor of each sent metric is one of those returned by Describe
// (unless the Collector is unchecked, see above). Returned metrics that
// share the same descriptor must differ in their variable label
// values.
func (collector *prometheusCollector) Collect(metricsChan chan<- prometheus.Metric) {
	validatorStatistics, err := collector.httpClient.GetValidatorStatistics(context.Background())
	if err != nil {
		log.Error("could not fetch validator statistics", "error", err.Error())
		return
	}

	for _, blsKey := range collector.blsKeys {
		statistics, found := validatorStatistics[blsKey]
		if !found {
			log.Error("no validator statistics found", "key", blsKey)
			continue
		}

		collector.collectMetricsForKey(metricsChan, statistics, blsKey)
	}
}

func (collector *prometheusCollector) collectMetricsForKey(
	metricsChan chan<- prometheus.Metric,
	statistics *httpClientWrapper.ValidatorStatistics,
	blsKey string,
) {
	tempRating := prometheus.MustNewConstMetric(collector.descriptors.tempRating, prometheus.GaugeValue, float64(statistics.TempRating), blsKey)
	metricsChan <- tempRating

	numLeaderSuccess := prometheus.MustNewConstMetric(collector.descriptors.numLeaderSuccess, prometheus.GaugeValue, float64(statistics.NumLeaderSuccess), blsKey)
	metricsChan <- numLeaderSuccess

	numLeaderFailure := prometheus.MustNewConstMetric(collector.descriptors.numLeaderFailure, prometheus.GaugeValue, float64(statistics.NumLeaderFailure), blsKey)
	metricsChan <- numLeaderFailure

	numValidatorSuccess := prometheus.MustNewConstMetric(collector.descriptors.numValidatorSuccess, prometheus.GaugeValue, float64(statistics.NumValidatorSuccess), blsKey)
	metricsChan <- numValidatorSuccess

	numValidatorFailure := prometheus.MustNewConstMetric(collector.descriptors.numValidatorFailure, prometheus.GaugeValue, float64(statistics.NumValidatorFailure), blsKey)
	metricsChan <- numValidatorFailure

	numValidatorIgnoredSignatures := prometheus.MustNewConstMetric(collector.descriptors.numValidatorIgnoredSignatures, prometheus.GaugeValue, float64(statistics.NumValidatorIgnoredSignatures), blsKey)
	metricsChan <- numValidatorIgnoredSignatures

	rating := prometheus.MustNewConstMetric(collector.descriptors.rating, prometheus.GaugeValue, float64(statistics.Rating), blsKey)
	metricsChan <- rating

	ratingModifier := prometheus.MustNewConstMetric(collector.descriptors.ratingModifier, prometheus.GaugeValue, float64(statistics.RatingModifier), blsKey)
	metricsChan <- ratingModifier

	totalNumLeaderSuccess := prometheus.MustNewConstMetric(collector.descriptors.totalNumLeaderSuccess, prometheus.GaugeValue, float64(statistics.TotalNumLeaderSuccess), blsKey)
	metricsChan <- totalNumLeaderSuccess

	totalNumLeaderFailure := prometheus.MustNewConstMetric(collector.descriptors.totalNumLeaderFailure, prometheus.GaugeValue, float64(statistics.TotalNumLeaderFailure), blsKey)
	metricsChan <- totalNumLeaderFailure

	totalNumValidatorSuccess := prometheus.MustNewConstMetric(collector.descriptors.totalNumValidatorSuccess, prometheus.GaugeValue, float64(statistics.TotalNumValidatorSuccess), blsKey)
	metricsChan <- totalNumValidatorSuccess

	totalNumValidatorFailure := prometheus.MustNewConstMetric(collector.descriptors.totalNumValidatorFailure, prometheus.GaugeValue, float64(statistics.TotalNumValidatorFailure), blsKey)
	metricsChan <- totalNumValidatorFailure

	totalNumValidatorIgnoredSignatures := prometheus.MustNewConstMetric(collector.descriptors.totalNumValidatorIgnoredSignatures, prometheus.GaugeValue, float64(statistics.TotalNumValidatorIgnoredSignatures), blsKey)
	metricsChan <- totalNumValidatorIgnoredSignatures

	shardId := prometheus.MustNewConstMetric(collector.descriptors.shardId, prometheus.GaugeValue, float64(statistics.ShardId), blsKey)
	metricsChan <- shardId

	validatorStatus := prometheus.MustNewConstMetric(collector.descriptors.validatorStatus, prometheus.GaugeValue, 0, blsKey, statistics.ValidatorStatus)
	metricsChan <- validatorStatus
}

// IsInterfaceNil returns true if there is no value under the interface
func (collector *prometheusCollector) IsInterfaceNil() bool {
	return collector == nil
}
