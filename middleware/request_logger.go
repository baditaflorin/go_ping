package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"ping/observability"
)

// ResponseWriter is a wrapper around http.ResponseWriter that captures the status code and size
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    int64
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write captures the response size
func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.written += int64(n)
	return n, err
}

// RequestInstrumentationMiddleware wraps an HTTP handler with:
// - Correlation ID extraction/generation
// - Request/response logging
// - Metrics recording (counters, histograms, gauges)
// - Correlation ID propagation via context
func RequestInstrumentationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get or create correlation ID from headers
		correlationID := r.Header.Get(observability.RequestIDHeader)
		if correlationID == "" {
			correlationID = r.Header.Get(observability.CorrelationIDHeader)
		}
		if correlationID == "" {
			correlationID = observability.GenerateCorrelationID()
		}

		// Add correlation ID to context
		ctx := observability.WithCorrelationID(r.Context(), correlationID)
		r = r.WithContext(ctx)

		// Add correlation ID to response headers so client can see it
		w.Header().Set(observability.ResponseCorrelationIDHeader, correlationID)

		// Initialize metrics
		metrics := observability.GetMetrics()
		startTime := time.Now()

		// Record request initiation
		defer metrics.RecordRequest()()

		// Wrap response writer to capture status and size
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // default
		}

		// Calculate request size
		requestSize := float64(r.ContentLength)
		if requestSize > 0 {
			metrics.ObserveRequestSize(requestSize)
		}

		// Log request start
		log.Printf("[%s] %s %s %s (id=%s)",
			r.Method,
			r.URL.Path,
			r.RemoteAddr,
			r.UserAgent(),
			correlationID)

		// Call next handler
		next.ServeHTTP(rw, r)

		// Record metrics
		duration := time.Since(startTime).Seconds()
		metrics.ObserveDuration(metrics.RequestDuration, duration)
		metrics.ObserveResponseSize(float64(rw.written))

		// Log request completion
		log.Printf("[%s] %s -> %d (duration=%.3fs, responseSize=%d, id=%s)",
			r.Method,
			r.URL.Path,
			rw.statusCode,
			duration,
			rw.written,
			correlationID)

		// Record HTTP errors
		if rw.statusCode >= 500 {
			metrics.HTTPErrorCounter.Inc()
		}
	})
}

// ContextLogMiddleware logs operations with the correlation ID from context
// This is useful for operations that receive context but need to log with correlation ID
func LogWithCorrelationID(ctx context.Context, message string, args ...interface{}) {
	correlationID := observability.GetCorrelationID(ctx)
	if correlationID != "" {
		prefix := fmt.Sprintf("[%s]", correlationID)
		log.Printf("%s %s", prefix, fmt.Sprintf(message, args...))
	} else {
		log.Printf(message, args...)
	}
}
