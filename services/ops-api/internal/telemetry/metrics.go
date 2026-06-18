package telemetry

import (
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// UCPFeedGenerationDuration tracks how long the UCP workflow takes to extract and upload
	UCPFeedGenerationDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "synq_ucp_feed_generation_duration_seconds",
		Help:    "Duration of UCP feed generation workflow",
		Buckets: prometheus.DefBuckets,
	}, []string{"tenant_id", "status"})

	// PIMProductsCreatedTotal tracks the number of products created via PIM
	PIMProductsCreatedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "synq_pim_products_created_total",
		Help: "Total number of products created",
	}, []string{"tenant_id"})

	// PIMProductsUpdatedTotal tracks the number of products updated via PIM
	PIMProductsUpdatedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "synq_pim_products_updated_total",
		Help: "Total number of products updated",
	}, []string{"tenant_id"})

	// PIMProductsDeletedTotal tracks the number of products deleted via PIM
	PIMProductsDeletedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "synq_pim_products_deleted_total",
		Help: "Total number of products deleted",
	}, []string{"tenant_id"})

	// UnifiedSyncSuccessTotal tracks successful pushes to Unified.to
	UnifiedSyncSuccessTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "synq_unified_sync_success_total",
		Help: "Total successful unified syncs",
	}, []string{"tenant_id", "connection_id", "action"})

	// UnifiedSyncFailedTotal tracks failed pushes to Unified.to
	UnifiedSyncFailedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "synq_unified_sync_failed_total",
		Help: "Total failed unified syncs",
	}, []string{"tenant_id", "connection_id", "action"})

	// UnifiedSyncDLQTotal tracks syncs moved to the DLQ
	UnifiedSyncDLQTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "synq_unified_sync_dlq_total",
		Help: "Total syncs moved to DLQ",
	}, []string{"tenant_id", "connection_id", "action"})
)

// SetupMetrics registers the /metrics endpoint on the provided router.
func SetupMetrics(r chi.Router) {
	r.Handle("/metrics", promhttp.Handler())
}
