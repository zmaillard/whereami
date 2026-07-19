package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Query execution metrics - tracks end-to-end query execution time
	queryDuration *prometheus.HistogramVec

	// Query total counter - counts queries by type and status
	queryTotal *prometheus.CounterVec

	// External API call duration - tracks individual API call latency
	externalAPICallDuration *prometheus.HistogramVec

	// External API call total - counts API calls by service and status
	externalAPICallTotal *prometheus.CounterVec
)

// Init initializes and registers all Prometheus metrics with the default registry.
// This should be called once during application startup, before the metrics server starts.
func Init() {
	queryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "whereami",
			Name:      "query_duration_seconds",
			Help:      "Duration of query execution by query type (end-to-end including DB and API calls)",
			Buckets:   prometheus.DefBuckets, // [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10]
		},
		[]string{"query_type"},
	)

	queryTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "whereami",
			Name:      "query_total",
			Help:      "Total number of queries executed by type and status",
		},
		[]string{"query_type", "status"}, // status: success, error
	)

	externalAPICallDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "whereami",
			Name:      "external_api_call_duration_seconds",
			Help:      "Duration of external API calls by service",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"service"}, // service: usgs_elevation, nws_weather, noaa_tides
	)

	externalAPICallTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "whereami",
			Name:      "external_api_call_total",
			Help:      "Total number of external API calls by service and status",
		},
		[]string{"service", "status"}, // status: success, error
	)
}

// RecordQueryDuration records the duration of a query execution
func RecordQueryDuration(queryType string, duration float64) {
	queryDuration.WithLabelValues(queryType).Observe(duration)
}

// RecordQuerySuccess increments the success counter for a query type
func RecordQuerySuccess(queryType string) {
	queryTotal.WithLabelValues(queryType, "success").Inc()
}

// RecordQueryError increments the error counter for a query type
func RecordQueryError(queryType string) {
	queryTotal.WithLabelValues(queryType, "error").Inc()
}

// RecordAPICallDuration records the duration of an external API call
func RecordAPICallDuration(service string, duration float64) {
	externalAPICallDuration.WithLabelValues(service).Observe(duration)
}

// RecordAPICallSuccess increments the success counter for an external API service
func RecordAPICallSuccess(service string) {
	externalAPICallTotal.WithLabelValues(service, "success").Inc()
}

// RecordAPICallError increments the error counter for an external API service
func RecordAPICallError(service string) {
	externalAPICallTotal.WithLabelValues(service, "error").Inc()
}
