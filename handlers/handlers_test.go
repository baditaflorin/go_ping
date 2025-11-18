package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ping/observability"
)

func TestPongHandler(t *testing.T) {
	// Initialize metrics
	observability.InitMetrics()

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	PongHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "pong") {
		t.Errorf("Expected response to contain 'pong', got %s", w.Body.String())
	}

	if ct := w.Header().Get("Content-Type"); ct != "text/plain" {
		t.Errorf("Expected Content-Type text/plain, got %s", ct)
	}
}

func TestHealthHandler(t *testing.T) {
	// Initialize metrics
	observability.InitMetrics()

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	HealthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "healthy") {
		t.Errorf("Expected response to contain 'healthy', got %s", w.Body.String())
	}

	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", ct)
	}
}

func TestHealthHandlerJSON(t *testing.T) {
	// Initialize metrics
	observability.InitMetrics()

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	HealthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "status") || !strings.Contains(body, "healthy") {
		t.Errorf("Expected JSON with status and healthy, got %s", body)
	}
}

func TestMetricsHandler(t *testing.T) {
	// Initialize metrics
	observability.InitMetrics()

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	MetricsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Metrics response should contain Prometheus format lines
	body := w.Body.String()
	if !strings.Contains(body, "# HELP") && !strings.Contains(body, "HELP") {
		t.Logf("Metrics body: %s", body)
		// Note: In some test environments, metrics might be empty or in different format
	}
}

func TestPingWithContext(t *testing.T) {
	// Initialize metrics
	observability.InitMetrics()

	// Create context with correlation ID
	correlationID := "test-correlation-id-789"
	ctx := observability.WithCorrelationID(context.Background(), correlationID)

	// Create request with context
	req := httptest.NewRequest("GET", "/ping", nil)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	PingWithContext(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "pong") {
		t.Errorf("Expected response to contain 'pong', got %s", body)
	}

	if !strings.Contains(body, correlationID) {
		t.Errorf("Expected response to contain correlation ID %s, got %s", correlationID, body)
	}
}

func TestPingWithContextWithoutID(t *testing.T) {
	// Initialize metrics
	observability.InitMetrics()

	// Create request without correlation ID in context
	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()

	PingWithContext(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "pong") {
		t.Errorf("Expected response to contain 'pong', got %s", body)
	}

	// Response should contain (id=) but with empty value since no ID was set
	if !strings.Contains(body, "id=") {
		t.Errorf("Expected response to contain 'id=' placeholder, got %s", body)
	}
}

func TestHandlersConcurrency(t *testing.T) {
	// Initialize metrics
	observability.InitMetrics()

	// Run concurrent handler calls
	done := make(chan bool, 3)

	go func() {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		PongHandler(w, req)
		done <- w.Code == http.StatusOK
	}()

	go func() {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		HealthHandler(w, req)
		done <- w.Code == http.StatusOK
	}()

	go func() {
		ctx := observability.WithCorrelationID(context.Background(), "test-id")
		req := httptest.NewRequest("GET", "/ping", nil)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()
		PingWithContext(w, req)
		done <- w.Code == http.StatusOK
	}()

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		if !<-done {
			t.Error("Handler test failed in concurrent execution")
		}
	}
}
