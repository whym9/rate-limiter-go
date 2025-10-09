package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// custom counter for your rate limiter decisions
	RateLimitRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_requests_total",
			Help: "Rate-limit decisions.",
		},
		[]string{"decision"}, // allow|deny
	)

	// optional: generic request duration histogram if not using echo-contrib
	HTTPReqDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method", "status"},
	)
)
