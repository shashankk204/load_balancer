
package core

import (
    "github.com/prometheus/client_golang/prometheus"
)

var (
    // ===== ROUTE-LEVEL METRICS =====
    
    // Request volume and patterns
    RouteRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "lb_route_requests_total",
            Help: "Total number of requests received per route and HTTP method",
        },
        []string{"route", "method", "status_code"},
    )

    // Request latency distribution
    RouteRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "lb_route_request_duration_seconds",
            Help:    "Request latency distribution per route",
            Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
        },
        []string{"route", "method"},
    )

    // Active concurrent requests per route
    RouteActiveRequests = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "lb_route_active_requests",
            Help: "Current number of active requests being processed per route",
        },
        []string{"route"},
    )

    // Route error rates by type
    RouteErrorsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "lb_route_errors_total",
            Help: "Total number of errors per route and error type",
        },
        []string{"route", "error_type"}, // error_type: backend_down, timeout, connection_refused, etc.
    )

    // Request size distribution
    RouteRequestSize = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "lb_route_request_size_bytes",
            Help:    "Size of HTTP requests per route",
            Buckets: []float64{100, 1000, 10000, 100000, 1000000},
        },
        []string{"route"},
    )

    // Response size distribution
    RouteResponseSize = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "lb_route_response_size_bytes",
            Help:    "Size of HTTP responses per route",
            Buckets: []float64{100, 1000, 10000, 100000, 1000000, 10000000},
        },
        []string{"route"},
    )

    // Load balancing strategy effectiveness
    RouteStrategyChanges = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "lb_route_strategy_changes_total",
            Help: "Number of times load balancing strategy was changed per route",
        },
        []string{"route", "from_strategy", "to_strategy"},
    )

    // ===== BACKEND-LEVEL METRICS =====

    // Backend health and availability
    BackendHealthStatus = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "lb_backend_health_status",
            Help: "Backend health status (1 = healthy, 0 = unhealthy)",
        },
        []string{"route", "backend", "backend_host"},
    )

    // Backend request distribution
    BackendRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "lb_backend_requests_total",
            Help: "Total number of requests sent to each backend",
        },
        []string{"route", "backend", "backend_host", "status_code"},
    )

    // Backend response time performance
    BackendRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "lb_backend_request_duration_seconds",
            Help:    "Request latency distribution per backend",
            Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
        },
        []string{"route", "backend", "backend_host"},
    )

    // Backend active connections
    BackendActiveConnections = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "lb_backend_active_connections",
            Help: "Current number of active connections to each backend",
        },
        []string{"route", "backend", "backend_host"},
    )

  

    // Backend failure tracking
    BackendFailuresTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "lb_backend_failures_total",
            Help: "Total number of backend failures by failure type",
        },
        []string{"route", "backend", "backend_host", "failure_type"}, // timeout, connection_refused, 5xx, etc.
    )

    // Backend selection frequency (for strategy analysis)
    BackendSelectionTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "lb_backend_selection_total",
            Help: "Number of times each backend was selected by load balancing strategy",
        },
        []string{"route", "backend", "backend_host", "strategy"},
    )

    // Backend health check metrics
    BackendHealthCheckDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "lb_backend_health_check_duration_seconds",
            Help:    "Duration of health checks per backend",
            Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
        },
        []string{"route", "backend", "backend_host"},
    )

    BackendHealthCheckFailures = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "lb_backend_health_check_failures_total",
            Help: "Total number of health check failures per backend",
        },
        []string{"route", "backend", "backend_host"},
    )

    // Backend load metrics
    BackendLoadScore = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "lb_backend_load_score",
            Help: "Current load score of each backend (used by least-loaded strategy)",
        },
        []string{"route", "backend", "backend_host"},
    )
)

func InitMetrics() {
    prometheus.MustRegister(
        // Route-level metrics
        RouteRequestsTotal,
        RouteRequestDuration,
        RouteActiveRequests,
        RouteErrorsTotal,
        RouteRequestSize,
        RouteResponseSize,
        RouteStrategyChanges,
        
        // Backend-level metrics
        BackendHealthStatus,
        BackendRequestsTotal,
        BackendRequestDuration,
        BackendActiveConnections,
        BackendFailuresTotal,
        BackendSelectionTotal,
        BackendHealthCheckDuration,
        BackendHealthCheckFailures,
        BackendLoadScore,
    )
}