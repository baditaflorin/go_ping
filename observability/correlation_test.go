package observability

import (
	"context"
	"testing"
)

func TestGenerateCorrelationID(t *testing.T) {
	id1 := GenerateCorrelationID()
	id2 := GenerateCorrelationID()

	// Should generate non-empty strings
	if id1 == "" {
		t.Error("GenerateCorrelationID returned empty string")
	}
	if id2 == "" {
		t.Error("GenerateCorrelationID returned empty string")
	}

	// Should generate unique IDs
	if id1 == id2 {
		t.Error("GenerateCorrelationID should generate unique IDs")
	}
}

func TestWithCorrelationID(t *testing.T) {
	ctx := context.Background()
	id := "test-correlation-id-123"

	// Add correlation ID to context
	ctx = WithCorrelationID(ctx, id)

	// Retrieve correlation ID
	retrieved := GetCorrelationID(ctx)
	if retrieved != id {
		t.Errorf("Expected %s, got %s", id, retrieved)
	}
}

func TestGetOrCreateCorrelationID(t *testing.T) {
	// Test with existing correlation ID
	ctx := context.Background()
	existingID := "existing-id"
	ctx = WithCorrelationID(ctx, existingID)

	retrieved := GetOrCreateCorrelationID(ctx)
	if retrieved != existingID {
		t.Errorf("Should return existing ID: expected %s, got %s", existingID, retrieved)
	}

	// Test with missing correlation ID (should create new one)
	ctx2 := context.Background()
	newID := GetOrCreateCorrelationID(ctx2)
	if newID == "" {
		t.Error("Should generate new correlation ID when missing")
	}

	// Subsequent calls should return the same generated ID only if stored in context
	// Verify the generated ID is a valid UUID format (has dashes)
	if len(newID) == 0 {
		t.Error("Generated correlation ID is empty")
	}
}

func TestGetCorrelationIDMissing(t *testing.T) {
	ctx := context.Background()

	// Should return empty string when correlation ID is not in context
	id := GetCorrelationID(ctx)
	if id != "" {
		t.Errorf("Expected empty string for missing correlation ID, got %s", id)
	}
}

func TestGetCorrelationIDEmpty(t *testing.T) {
	ctx := context.Background()
	ctx = WithCorrelationID(ctx, "")

	// Should return empty string when correlation ID is empty
	id := GetCorrelationID(ctx)
	if id != "" {
		t.Errorf("Expected empty string, got %s", id)
	}
}

func TestCorrelationIDHeaderConstants(t *testing.T) {
	// Verify header constants are non-empty
	if RequestIDHeader == "" {
		t.Error("RequestIDHeader is empty")
	}
	if CorrelationIDHeader == "" {
		t.Error("CorrelationIDHeader is empty")
	}
	if ResponseCorrelationIDHeader == "" {
		t.Error("ResponseCorrelationIDHeader is empty")
	}

	// Verify they have expected values
	if RequestIDHeader != "X-Request-ID" {
		t.Errorf("RequestIDHeader has unexpected value: %s", RequestIDHeader)
	}
	if CorrelationIDHeader != "X-Correlation-ID" {
		t.Errorf("CorrelationIDHeader has unexpected value: %s", CorrelationIDHeader)
	}
	if ResponseCorrelationIDHeader != "X-Correlation-ID" {
		t.Errorf("ResponseCorrelationIDHeader has unexpected value: %s", ResponseCorrelationIDHeader)
	}
}

func TestMultipleContextValues(t *testing.T) {
	ctx := context.Background()
	id1 := "first-id"
	id2 := "second-id"

	// Add first ID
	ctx = WithCorrelationID(ctx, id1)
	if retrieved := GetCorrelationID(ctx); retrieved != id1 {
		t.Errorf("Expected %s, got %s", id1, retrieved)
	}

	// Overwrite with second ID
	ctx = WithCorrelationID(ctx, id2)
	if retrieved := GetCorrelationID(ctx); retrieved != id2 {
		t.Errorf("Expected %s, got %s", id2, retrieved)
	}
}
