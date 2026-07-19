package queries

import (
	"time"

	"github.com/zmaillard/whereami/metrics"
	"github.com/zmaillard/whereami/models"
)

// InstrumentQuery wraps a models.Querier with Prometheus metrics instrumentation.
// It automatically tracks:
// - Query execution duration (end-to-end)
// - Success/error counts
//
// Usage:
//
//	instrumented := queries.InstrumentQuery("county", database.GetCounty)
func InstrumentQuery(queryType string, querier models.Querier) models.Querier {
	return func(coordinates models.Coordinates) (models.Result, error) {
		start := time.Now()

		// Execute the actual query
		result, err := querier(coordinates)

		// Record metrics
		duration := time.Since(start).Seconds()
		metrics.RecordQueryDuration(queryType, duration)

		if err != nil {
			metrics.RecordQueryError(queryType)
		} else {
			metrics.RecordQuerySuccess(queryType)
		}

		return result, err
	}
}
