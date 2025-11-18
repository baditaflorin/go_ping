package observability

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus collectors for the application.
// This struct is the central registry for all metrics.
type Metrics struct {
	// HTTP Request Metrics
	RequestCounter      prometheus.Counter
	RequestDuration     prometheus.Histogram
	RequestSize         prometheus.Histogram
	ResponseSize        prometheus.Histogram
	HTTPErrorCounter    prometheus.Counter
	ActiveRequestsGauge prometheus.Gauge

	// Background Job Metrics
	BackgroundJobCounter    prometheus.Counter
	BackgroundJobDuration   prometheus.Histogram
	BackgroundJobErrorCount prometheus.Counter

	// External API Call Metrics
	APICallCounter      prometheus.Counter
	APICallDuration     prometheus.Histogram
	APICallErrorCounter prometheus.Counter

	// File/CSV/TSV Processing Metrics
	FileProcessCounter      prometheus.Counter
	FileProcessDuration     prometheus.Histogram
	FileProcessBytesCounter prometheus.Counter
	FileProcessErrorCounter prometheus.Counter
}

var (
	metricsInstance *Metrics
	once             sync.Once
)

// InitMetrics initializes and registers all Prometheus metrics.
// This should be called once at application startup.
// It uses sync.Once to ensure metrics are only registered once.
func InitMetrics() *Metrics {
	once.Do(func() {
		metricsInstance = &Metrics{
			// HTTP Request Metrics
			RequestCounter: promauto.NewCounter(prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests received",
			}),
			RequestDuration: promauto.NewHistogram(prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request latency in seconds",
				Buckets: prometheus.DefBuckets,
			}),
			RequestSize: promauto.NewHistogram(prometheus.HistogramOpts{
				Name:    "http_request_size_bytes",
				Help:    "HTTP request size in bytes",
				Buckets: []float64{100, 500, 1000, 5000, 10000, 50000, 100000},
			}),
			ResponseSize: promauto.NewHistogram(prometheus.HistogramOpts{
				Name:    "http_response_size_bytes",
				Help:    "HTTP response size in bytes",
				Buckets: []float64{100, 500, 1000, 5000, 10000, 50000, 100000},
			}),
			HTTPErrorCounter: promauto.NewCounter(prometheus.CounterOpts{
				Name: "http_errors_total",
				Help: "Total number of HTTP errors (5xx)",
			}),
			ActiveRequestsGauge: promauto.NewGauge(prometheus.GaugeOpts{
				Name: "http_requests_active",
				Help: "Number of currently active HTTP requests",
			}),

			// Background Job Metrics
			BackgroundJobCounter: promauto.NewCounter(prometheus.CounterOpts{
				Name: "background_jobs_total",
				Help: "Total number of background jobs executed",
			}),
			BackgroundJobDuration: promauto.NewHistogram(prometheus.HistogramOpts{
				Name:    "background_job_duration_seconds",
				Help:    "Background job execution time in seconds",
				Buckets: prometheus.DefBuckets,
			}),
			BackgroundJobErrorCount: promauto.NewCounter(prometheus.CounterOpts{
				Name: "background_job_errors_total",
				Help: "Total number of background job errors",
			}),

			// External API Call Metrics
			APICallCounter: promauto.NewCounter(prometheus.CounterOpts{
				Name: "api_calls_total",
				Help: "Total number of external API calls made",
			}),
			APICallDuration: promauto.NewHistogram(prometheus.HistogramOpts{
				Name:    "api_call_duration_seconds",
				Help:    "External API call latency in seconds",
				Buckets: prometheus.DefBuckets,
			}),
			APICallErrorCounter: promauto.NewCounter(prometheus.CounterOpts{
				Name: "api_call_errors_total",
				Help: "Total number of external API call errors",
			}),

			// File/CSV/TSV Processing Metrics
			FileProcessCounter: promauto.NewCounter(prometheus.CounterOpts{
				Name: "file_processes_total",
				Help: "Total number of file processing operations",
			}),
			FileProcessDuration: promauto.NewHistogram(prometheus.HistogramOpts{
				Name:    "file_process_duration_seconds",
				Help:    "File processing duration in seconds",
				Buckets: prometheus.DefBuckets,
			}),
			FileProcessBytesCounter: promauto.NewCounter(prometheus.CounterOpts{
				Name: "file_process_bytes_total",
				Help: "Total bytes processed",
			}),
			FileProcessErrorCounter: promauto.NewCounter(prometheus.CounterOpts{
				Name: "file_process_errors_total",
				Help: "Total number of file processing errors",
			}),
		}
	})
	return metricsInstance
}

// GetMetrics returns the initialized Metrics instance.
// InitMetrics must be called before calling this function.
func GetMetrics() *Metrics {
	if metricsInstance == nil {
		panic("metrics not initialized: call InitMetrics() first")
	}
	return metricsInstance
}

// RecordRequest increments the request counter and returns a function to observe duration.
// Usage:
//   defer metrics.RecordRequest()()
func (m *Metrics) RecordRequest() func() {
	m.RequestCounter.Inc()
	m.ActiveRequestsGauge.Inc()
	return func() {
		m.ActiveRequestsGauge.Dec()
	}
}

// ObserveDuration observes the duration of an operation in seconds.
func (m *Metrics) ObserveDuration(histogram prometheus.Histogram, duration float64) {
	histogram.Observe(duration)
}

// IncError increments the error counter.
func (m *Metrics) IncError(counter prometheus.Counter) {
	counter.Inc()
}

// ObserveRequestSize observes the size of an HTTP request.
func (m *Metrics) ObserveRequestSize(size float64) {
	m.RequestSize.Observe(size)
}

// ObserveResponseSize observes the size of an HTTP response.
func (m *Metrics) ObserveResponseSize(size float64) {
	m.ResponseSize.Observe(size)
}

// RecordAPICall records an external API call with optional error.
func (m *Metrics) RecordAPICall(duration float64, err error) {
	m.APICallCounter.Inc()
	m.APICallDuration.Observe(duration)
	if err != nil {
		m.APICallErrorCounter.Inc()
	}
}

// RecordBackgroundJob records a background job execution with optional error.
func (m *Metrics) RecordBackgroundJob(duration float64, err error) {
	m.BackgroundJobCounter.Inc()
	m.BackgroundJobDuration.Observe(duration)
	if err != nil {
		m.BackgroundJobErrorCount.Inc()
	}
}

// RecordFileProcess records file processing with size and optional error.
func (m *Metrics) RecordFileProcess(duration float64, bytes float64, err error) {
	m.FileProcessCounter.Inc()
	m.FileProcessDuration.Observe(duration)
	m.FileProcessBytesCounter.Add(bytes)
	if err != nil {
		m.FileProcessErrorCounter.Inc()
	}
}
