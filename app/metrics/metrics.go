package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HttpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "userservice_http_requests_total",
			Help: "Total number of HTTP requests handled by UserService",
		},
		[]string{"method", "path", "status"},
	)

	HttpRequestDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "userservice_http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	UserRequestsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "userservice_requests_total",
		Help: "Total number of business-level requests handled by UserService",
	})

	UserRequestOutcomesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "userservice_request_outcomes_total",
			Help: "Total number of UserService request outcomes",
		},
		[]string{"outcome"},
	)

	UserRequestDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "userservice_request_duration_seconds",
		Help:    "Duration of UserService request handling",
		Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2, 5},
	})

	PasswordResetRequestsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "userservice_password_reset_requests_total",
		Help: "Total number of password reset requests",
	})

	NotificationCallsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "userservice_notification_calls_total",
			Help: "Total calls from UserService to NotificationService",
		},
		[]string{"status"},
	)
)
