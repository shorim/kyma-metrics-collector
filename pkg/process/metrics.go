package process

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	kmccache "github.com/kyma-project/kyma-metrics-collector/pkg/cache"
	skrcommons "github.com/kyma-project/kyma-metrics-collector/pkg/skr/commons"
)

const (
	namespace          = "kmc"
	subsystem          = "process"
	shootNameLabel     = "shoot_name"
	instanceIdLabel    = "instance_id"
	runtimeIdLabel     = "runtime_id"
	subAccountLabel    = "sub_account_id"
	globalAccountLabel = "global_account_id"
	successLabel       = "success"
	withOldMetricLabel = "with_old_metric"
	trackableLabel     = "trackable"
)

var (
	itemsInCache = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "items_in_cache",
			Help:      "Number of items in the cache.",
		}, nil)

	subAccountProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "sub_account_total",
			Help:      "Number of processings per subaccount, including successful and failed.",
		},
		[]string{successLabel, shootNameLabel, instanceIdLabel, runtimeIdLabel, subAccountLabel, globalAccountLabel},
	)
	subAccountProcessedTimeStamp = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "sub_account_processed_timestamp_seconds",
			Help:      "Unix timestamp (in seconds) of last successful processing of subaccount.",
		},
		[]string{withOldMetricLabel, shootNameLabel, instanceIdLabel, runtimeIdLabel, subAccountLabel, globalAccountLabel},
	)
	oldMetricsPublishedGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "old_metric_published",
			Help:      "Number of consecutive re-sends of old metrics to EDP per cluster. It will reset to 0 when new metric data is published.",
		},
		[]string{shootNameLabel, instanceIdLabel, runtimeIdLabel, subAccountLabel, globalAccountLabel},
	)
	kebFetchedClusters = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "fetched_clusters_total",
			Help:      "All clusters fetched from KEB, including trackable and not trackable.",
		},
		[]string{trackableLabel, shootNameLabel, instanceIdLabel, runtimeIdLabel, subAccountLabel, globalAccountLabel},
	)
)

func recordItemsInCache(count float64) {
	itemsInCache.WithLabelValues().Set(count)
}

func recordKEBFetchedClusters(trackable bool, shootName, instanceID, runtimeID, subAccountID, globalAccountID string) {
	// the order if the values should be same as defined in the metric declaration.
	kebFetchedClusters.WithLabelValues(
		strconv.FormatBool(trackable),
		shootName,
		instanceID,
		runtimeID,
		subAccountID,
		globalAccountID,
	).Inc()
}

// deleteMetrics deletes all the metrics for the provided shoot.
// Returns true if some metrics are deleted, returns false if no metrics are deleted for that subAccount.
func deleteMetrics(shootInfo kmccache.Record) bool {
	matchLabels := prometheus.Labels{
		"shoot_name":        shootInfo.ShootName,
		"instance_id":       shootInfo.InstanceID,
		"runtime_id":        shootInfo.RuntimeID,
		"sub_account_id":    shootInfo.SubAccountID,
		"global_account_id": shootInfo.GlobalAccountID,
	}

	count := 0 // total numbers of metrics deleted
	count += subAccountProcessed.DeletePartialMatch(matchLabels)
	count += subAccountProcessedTimeStamp.DeletePartialMatch(matchLabels)
	count += oldMetricsPublishedGauge.DeletePartialMatch(matchLabels)

	// delete metrics for SKR queries.
	return skrcommons.DeleteMetrics(shootInfo) && count > 0
}

func recordSubAccountProcessed(success bool, shootInfo kmccache.Record) {
	// the order if the values should be same as defined in the metric declaration.
	subAccountProcessed.WithLabelValues(
		strconv.FormatBool(success),
		shootInfo.ShootName,
		shootInfo.InstanceID,
		shootInfo.RuntimeID,
		shootInfo.SubAccountID,
		shootInfo.GlobalAccountID,
	).Inc()
}

func recordSubAccountProcessedTimeStamp(withOldMetric bool, shootInfo kmccache.Record) {
	// the order if the values should be same as defined in the metric declaration.
	subAccountProcessedTimeStamp.WithLabelValues(
		strconv.FormatBool(withOldMetric),
		shootInfo.ShootName,
		shootInfo.InstanceID,
		shootInfo.RuntimeID,
		shootInfo.SubAccountID,
		shootInfo.GlobalAccountID,
	).SetToCurrentTime()
}

func recordOldMetricsPublishedGauge(shootInfo kmccache.Record) {
	// the order if the values should be same as defined in the metric declaration.
	oldMetricsPublishedGauge.WithLabelValues(
		shootInfo.ShootName,
		shootInfo.InstanceID,
		shootInfo.RuntimeID,
		shootInfo.SubAccountID,
		shootInfo.GlobalAccountID,
	).Inc()
}

func resetOldMetricsPublishedGauge(shootInfo kmccache.Record) {
	// the order if the values should be same as defined in the metric declaration.
	oldMetricsPublishedGauge.WithLabelValues(
		shootInfo.ShootName,
		shootInfo.InstanceID,
		shootInfo.RuntimeID,
		shootInfo.SubAccountID,
		shootInfo.GlobalAccountID,
	).Set(0.0)
}
