package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ping/handlers"
	"ping/middleware"
	"ping/observability"
)

func main() {
	// Initialize metrics
	metrics := observability.InitMetrics()
	log.Println("✓ Metrics initialized")

	// Create HTTP mux
	mux := http.NewServeMux()

	// Register handlers with instrumentation middleware
	mux.HandleFunc("/", handlers.PongHandler)
	mux.HandleFunc("/metrics", handlers.MetricsHandler)
	mux.HandleFunc("/health", handlers.HealthHandler)

	// Wrap mux with middleware
	instrumentedMux := middleware.RequestInstrumentationMiddleware(mux)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      instrumentedMux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Channel for graceful shutdown
	done := make(chan struct{})
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		log.Printf("⇨ listening on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Log startup info
	log.Printf("✓ Pong service started (version: 1.0.0)")
	log.Printf("✓ Metrics available at http://localhost:%s/metrics", port)
	log.Printf("✓ Correlation ID headers: %s, %s", observability.RequestIDHeader, observability.CorrelationIDHeader)

	// Wait for shutdown signal
	<-sigChan
	log.Println("⇨ Shutdown signal received, shutting down gracefully...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

	close(done)
	log.Println("✓ Server stopped")

	// Log final metrics info
	_ = metrics // Use metrics to avoid unused variable warning
}
