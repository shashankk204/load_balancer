package core

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	RequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "lb_requests_total",
			Help: "Total number of requests received per route",
		},
		[]string{"route"},
	)

	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "lb_request_duration_seconds",
			Help:    "Request latency per route",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"route"},
	)

	BackendErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "lb_backend_errors_total",
			Help: "Total number of failed backend requests per route",
		},
		[]string{"route"},
	)
	 BackendAlive = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "lb_backend_alive",
            Help: "Backend health status (1 = alive, 0 = down).",
        },
        []string{"route", "backend"},
    )

    // Total requests handled per backend
    BackendRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "lb_backend_requests_total",
            Help: "Total number of requests handled by each backend.",
        },
        []string{"route", "backend"},
    )

    // Average latency (exposed as a gauge so you can export AvgLatency())
    BackendAvgLatency = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "lb_backend_avg_latency_ms",
            Help: "Average latency of requests handled by each backend (in ms).",
        },
        []string{"route", "backend"},
    )

    // Active requests (live connections)
    BackendActive = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "lb_backend_active_requests",
            Help: "Current number of active requests per backend.",
        },
        []string{"route", "backend"},
    )
)

func InitMetrics() {
	prometheus.MustRegister(RequestsTotal, RequestDuration, BackendErrors,
		BackendAlive,
        BackendRequestsTotal,
        BackendAvgLatency,
        BackendActive,)
}
