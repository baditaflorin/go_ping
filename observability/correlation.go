package observability

import (
	"context"

	"github.com/google/uuid"
)

// CorrelationIDKey is the context key for storing correlation IDs
type CorrelationIDKey string

const (
	// CorrelationID is the key used to store the correlation ID in the request context
	CorrelationID CorrelationIDKey = "correlation-id"

	// RequestIDHeader is the HTTP header name for incoming request IDs
	RequestIDHeader = "X-Request-ID"

	// CorrelationIDHeader is the HTTP header name for correlation IDs (alternative)
	CorrelationIDHeader = "X-Correlation-ID"

	// ResponseCorrelationIDHeader is the HTTP header name for exposing correlation ID in responses
	ResponseCorrelationIDHeader = "X-Correlation-ID"
)

// GenerateCorrelationID creates a new UUID-based correlation ID
func GenerateCorrelationID() string {
	return uuid.New().String()
}

// GetOrCreateCorrelationID retrieves an existing correlation ID from the context
// or generates a new one if it doesn't exist
func GetOrCreateCorrelationID(ctx context.Context) string {
	if corrID, ok := ctx.Value(CorrelationID).(string); ok && corrID != "" {
		return corrID
	}
	return GenerateCorrelationID()
}

// WithCorrelationID adds a correlation ID to the context
func WithCorrelationID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, CorrelationID, id)
}

// GetCorrelationID retrieves the correlation ID from the context
// Returns empty string if not found
func GetCorrelationID(ctx context.Context) string {
	if corrID, ok := ctx.Value(CorrelationID).(string); ok {
		return corrID
	}
	return ""
}
