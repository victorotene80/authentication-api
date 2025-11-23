package metrics

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"authentication/shared/config"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/otlptranslator"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.uber.org/zap"
)

// MetricsServer manages the Prometheus metrics HTTP server
type MetricsServer struct {
	server   *http.Server
	exporter *prometheus.Exporter
	logger   *zap.Logger
}

// InitMetrics sets up Prometheus metrics with HTTP server
func InitMetrics(ctx context.Context, cfg config.MetricsConfig, res *resource.Resource) (*metric.MeterProvider, *MetricsServer, error) {
	if !cfg.Enabled {
		return nil, nil, nil
	}

	// Create Prometheus exporter
	exporter, err := prometheus.New(
		prometheus.WithTranslationStrategy(otlptranslator.NoTranslation),
		prometheus.WithoutScopeInfo(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Prometheus exporter: %w", err)
	}

	// Create meter provider
	mp := metric.NewMeterProvider(
		metric.WithReader(exporter),
		metric.WithResource(res),
	)

	// Create temporary logger for metrics server
	logger, _ := zap.NewProduction()

	// Setup HTTP mux
	mux := http.NewServeMux()

	// Prometheus scrape endpoint
	mux.Handle(cfg.MetricsPath, promhttp.Handler())

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Readiness check
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("READY"))
	})

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           mux,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	metricsServer := &MetricsServer{
		server:   server,
		exporter: exporter,
		logger:   logger,
	}

	return mp, metricsServer, nil
}

// Start begins serving Prometheus metrics
func (ms *MetricsServer) Start() error {
	if ms == nil || ms.server == nil {
		return nil
	}

	errCh := make(chan error, 1)

	go func() {
		ms.logger.Info("starting metrics server",
			zap.String("addr", ms.server.Addr),
		)

		if err := ms.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			ms.logger.Error("metrics server error", zap.Error(err))
			errCh <- err
		}
	}()

	// Wait for startup or immediate failure
	select {
	case err := <-errCh:
		return fmt.Errorf("failed to start metrics server: %w", err)
	case <-time.After(150 * time.Millisecond):
		ms.logger.Info("metrics server started successfully",
			zap.String("addr", ms.server.Addr),
		)
		return nil
	}
}

// Shutdown gracefully stops the metrics server
func (ms *MetricsServer) Shutdown(ctx context.Context) error {
	if ms == nil || ms.server == nil {
		return nil
	}

	ms.logger.Info("shutting down metrics server")

	if err := ms.server.Shutdown(ctx); err != nil {
		ms.logger.Error("error shutting down metrics server", zap.Error(err))
		return fmt.Errorf("metrics server shutdown error: %w", err)
	}

	ms.logger.Info("metrics server shut down successfully")
	return nil
}

// GetAddr returns the address the metrics server is listening on
func (ms *MetricsServer) GetAddr() string {
	if ms == nil || ms.server == nil {
		return ""
	}
	return ms.server.Addr
}
