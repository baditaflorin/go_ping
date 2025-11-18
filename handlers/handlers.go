package handlers

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"ping/middleware"
	"ping/observability"
)

// PongHandler is the main health check endpoint that returns "pong"
func PongHandler(w http.ResponseWriter, r *http.Request) {
	// Log with correlation ID from context
	middleware.LogWithCorrelationID(r.Context(), "Processing pong request")

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "pong")
}

// HealthHandler is a health check endpoint that can be used by load balancers
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	middleware.LogWithCorrelationID(r.Context(), "Processing health check request")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, `{"status":"healthy"}`)
}

// MetricsHandler exposes Prometheus metrics
func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	middleware.LogWithCorrelationID(r.Context(), "Processing metrics request")

	// Use Prometheus HTTP handler to serve metrics
	// This handler doesn't need instrumentation to avoid recursive metrics
	handler := promhttp.Handler()
	handler.ServeHTTP(w, r)
}

// PingWithContext is a handler that demonstrates correlation ID usage in business logic
func PingWithContext(w http.ResponseWriter, r *http.Request) {
	// Get correlation ID from context
	correlationID := observability.GetCorrelationID(r.Context())
	middleware.LogWithCorrelationID(r.Context(), "Processing ping request with context id=%s", correlationID)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "pong (id=%s)\n", correlationID)
}
