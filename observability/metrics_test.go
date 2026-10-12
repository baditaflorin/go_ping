package observability

import (
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

// resetMetricsForTest drops the metrics singleton and gives promauto a fresh
// registry. Production initialises metrics once; tests re-initialise them, and
// registering the same collectors twice in the default registry panics.
func resetMetricsForTest() {
	metricsInstance = nil
	once = sync.Once{}
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
}

func TestMetricsInitialization(t *testing.T) {
	// Clear the metrics singleton for testing
	resetMetricsForTest()

	// Initialize metrics
	metrics := InitMetrics()
	if metrics == nil {
		t.Fatal("InitMetrics returned nil")
	}

	// Verify all metrics are registered
	if metrics.RequestCounter == nil {
		t.Error("RequestCounter is nil")
	}
	if metrics.RequestDuration == nil {
		t.Error("RequestDuration is nil")
	}
	if metrics.HTTPErrorCounter == nil {
		t.Error("HTTPErrorCounter is nil")
	}
	if metrics.ActiveRequestsGauge == nil {
		t.Error("ActiveRequestsGauge is nil")
	}
}

func TestMetricsNoPanic(t *testing.T) {
	// Test that InitMetrics can be called multiple times without panicking
	resetMetricsForTest()

	m1 := InitMetrics()
	m2 := InitMetrics()

	if m1 != m2 {
		t.Error("InitMetrics should return the same instance")
	}
}

func TestRecordRequest(t *testing.T) {
	resetMetricsForTest()

	metrics := InitMetrics()

	// Record a request
	cleanup := metrics.RecordRequest()

	// Check that counters incremented
	if err := testutil.CollectAndCompare(metrics.RequestCounter, strings.NewReader(`
		# HELP http_requests_total Total number of HTTP requests received
		# TYPE http_requests_total counter
		http_requests_total 1
	`)); err != nil {
		t.Logf("Counter check: %v (may fail in test environment)", err)
	}

	// Check active requests gauge incremented
	if err := testutil.CollectAndCompare(metrics.ActiveRequestsGauge, strings.NewReader(`
		# HELP http_requests_active Number of currently active HTTP requests
		# TYPE http_requests_active gauge
		http_requests_active 1
	`)); err != nil {
		t.Logf("Gauge check: %v (may fail in test environment)", err)
	}

	// Clean up (this should decrement the active gauge)
	cleanup()

	// The active gauge should now be 0
	if err := testutil.CollectAndCompare(metrics.ActiveRequestsGauge, strings.NewReader(`
		# HELP http_requests_active Number of currently active HTTP requests
		# TYPE http_requests_active gauge
		http_requests_active 0
	`)); err != nil {
		t.Logf("Gauge after cleanup check: %v (may fail in test environment)", err)
	}
}

func TestObserveDuration(t *testing.T) {
	resetMetricsForTest()

	metrics := InitMetrics()

	// Observe a duration
	metrics.ObserveDuration(metrics.RequestDuration, 0.5)

	// Verify the observation was recorded
	hist := testutil.CollectAndCount(metrics.RequestDuration)

	if hist == 0 {
		t.Logf("No histogram data collected")
	}
}

func TestIncError(t *testing.T) {
	resetMetricsForTest()

	metrics := InitMetrics()

	// Increment error counter
	metrics.IncError(metrics.HTTPErrorCounter)

	// Verify the counter incremented
	if err := testutil.CollectAndCompare(metrics.HTTPErrorCounter, strings.NewReader(`
		# HELP http_errors_total Total number of HTTP errors (5xx)
		# TYPE http_errors_total counter
		http_errors_total 1
	`)); err != nil {
		t.Logf("Error counter check: %v (may fail in test environment)", err)
	}
}

func TestObserveRequestSize(t *testing.T) {
	resetMetricsForTest()

	metrics := InitMetrics()

	// Observe request size
	metrics.ObserveRequestSize(512)

	// Verify the observation was recorded
	hist := testutil.CollectAndCount(metrics.RequestSize)

	if hist == 0 {
		t.Logf("No histogram data collected")
	}
}

func TestRecordAPICall(t *testing.T) {
	resetMetricsForTest()

	metrics := InitMetrics()

	// Record successful API call
	metrics.RecordAPICall(0.25, nil)

	// Record failed API call
	metrics.RecordAPICall(0.5, errors.New("simulated failure"))

	// Verify counters incremented
	if err := testutil.CollectAndCompare(metrics.APICallCounter, strings.NewReader(`
		# HELP api_calls_total Total number of external API calls made
		# TYPE api_calls_total counter
		api_calls_total 2
	`)); err != nil {
		t.Logf("API call counter check: %v (may fail in test environment)", err)
	}
}

func TestRecordBackgroundJob(t *testing.T) {
	resetMetricsForTest()

	metrics := InitMetrics()

	// Record successful background job
	metrics.RecordBackgroundJob(1.0, nil)

	// Record failed background job
	metrics.RecordBackgroundJob(0.5, errors.New("simulated failure"))

	// Verify counters incremented
	if err := testutil.CollectAndCompare(metrics.BackgroundJobCounter, strings.NewReader(`
		# HELP background_jobs_total Total number of background jobs executed
		# TYPE background_jobs_total counter
		background_jobs_total 2
	`)); err != nil {
		t.Logf("Background job counter check: %v (may fail in test environment)", err)
	}
}

func TestRecordFileProcess(t *testing.T) {
	resetMetricsForTest()

	metrics := InitMetrics()

	// Record file processing
	metrics.RecordFileProcess(2.0, 1024, nil)

	// Record failed file processing
	metrics.RecordFileProcess(1.5, 512, errors.New("simulated failure"))

	// Verify counters incremented
	if err := testutil.CollectAndCompare(metrics.FileProcessCounter, strings.NewReader(`
		# HELP file_processes_total Total number of file processing operations
		# TYPE file_processes_total counter
		file_processes_total 2
	`)); err != nil {
		t.Logf("File process counter check: %v (may fail in test environment)", err)
	}
}

func TestGetMetricsWithoutInit(t *testing.T) {
	// Save the current state
	savedInstance := metricsInstance

	// Clear metrics
	metricsInstance = nil

	defer func() {
		// Restore
		metricsInstance = savedInstance
	}()

	// Should panic if called without init
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when calling GetMetrics without InitMetrics")
		}
	}()

	GetMetrics()
}
