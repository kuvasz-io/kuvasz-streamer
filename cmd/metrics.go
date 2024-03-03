package main

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	requestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "streamer_operations_total",
			Help: "Total number of INSERT/UPDATE/DELETE operations.",
		}, []string{"database", "sid", "table", "operation", "result"},
	)

	requestDuration = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "streamer_operations_seconds",
			Help:       "Duration of INSERT/UPDATE/DELETE operations.",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
			MaxAge:     1 * time.Minute,
		}, []string{"database", "sid", "table", "operation", "result"},
	)

	syncRowsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "streamer_sync_total_rows",
			Help: "Total number of rows synced.",
		}, []string{"database", "sid", "table"},
	)
	syncBytesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "streamer_sync_total_bytes",
			Help: "Total number of bytes synced.",
		}, []string{"database", "sid", "table"},
	)
	jobsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "streamer_jobs_total",
			Help: "Total number of INSERT/UPDATE/DELETE operations.",
		}, []string{"channel"},
	)
)
