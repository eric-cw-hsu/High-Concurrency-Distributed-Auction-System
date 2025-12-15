package metrics

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/shared/logger"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

// StandaloneMetricsServer provides a dedicated HTTP server for Prometheus metrics
type StandaloneMetricsServer struct {
	server    *http.Server
	collector *PrometheusCollector
}

// NewStandaloneMetricsServer creates a new dedicated metrics server
func NewStandaloneMetricsServer(port int, collector *PrometheusCollector) *StandaloneMetricsServer {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(collector.GetRegistry(), promhttp.HandlerOpts{}))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &StandaloneMetricsServer{
		server:    server,
		collector: collector,
	}
}

// Start starts the metrics server
func (s *StandaloneMetricsServer) Start() error {
	logger.Infof("Starting metrics server on %s", s.server.Addr)
	return s.server.ListenAndServe()
}

// Stop gracefully stops the metrics server
func (s *StandaloneMetricsServer) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logger.Info("Stopping metrics server")
	return s.server.Shutdown(ctx)
}

// StartAsync starts the metrics server in a goroutine with logging.
func (s *StandaloneMetricsServer) StartAsync() {
	go func() {
		if err := s.Start(); err != nil {
			if err == http.ErrServerClosed {
				logger.Info("Metrics server closed", nil)
				return
			}

			logger.Error("Metrics server failed", logrus.Fields{
				"error": err.Error(),
			})
		}
	}()
}

// EmbeddedMetricsEndpoints provides metrics endpoints that can be embedded in existing services
type EmbeddedMetricsEndpoints struct {
	collector *PrometheusCollector
	handler   http.Handler
}

// NewEmbeddedMetricsEndpoints creates metrics endpoints for embedding in existing services
func NewEmbeddedMetricsEndpoints(collector *PrometheusCollector) *EmbeddedMetricsEndpoints {
	return &EmbeddedMetricsEndpoints{
		collector: collector,
		handler:   promhttp.HandlerFor(collector.GetRegistry(), promhttp.HandlerOpts{}),
	}
}

// MetricsHandler returns the HTTP handler for metrics endpoint
func (e *EmbeddedMetricsEndpoints) MetricsHandler() http.Handler {
	return e.handler
}

// HealthHandler returns a simple health check handler
func (e *EmbeddedMetricsEndpoints) HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}

// RegisterWithGin registers metrics and health endpoints with Gin router
func (e *EmbeddedMetricsEndpoints) RegisterWithGin(router *gin.Engine, metricsPath, healthPath string) {
	if metricsPath == "" {
		metricsPath = "/metrics"
	}
	if healthPath == "" {
		healthPath = "/health"
	}

	// Register metrics endpoint
	router.GET(metricsPath, gin.WrapH(e.MetricsHandler()))

	// Register health endpoint
	router.GET(healthPath, func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})
}

// RegisterWithGinGroup registers metrics and health endpoints with Gin router group
func (e *EmbeddedMetricsEndpoints) RegisterWithGinGroup(group *gin.RouterGroup, metricsPath, healthPath string) {
	if metricsPath == "" {
		metricsPath = "/metrics"
	}
	if healthPath == "" {
		healthPath = "/health"
	}

	// Register metrics endpoint
	group.GET(metricsPath, gin.WrapH(e.MetricsHandler()))

	// Register health endpoint
	group.GET(healthPath, func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})
}
