package metrics

import (
	"context"
	"fmt"
	"time"

	"authentication/shared/config"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ObservabilityManager manages all observability components
type ObservabilityManager struct {
	tracerProvider  *sdktrace.TracerProvider
	meterProvider   *sdkmetric.MeterProvider
	metricsServer   *MetricsServer
	logger          *zap.Logger
	metricsRecorder *MetricsRecorder
	logstashWriter  *LogstashWriter
	resource        *resource.Resource
	config          *config.Config
}

// NewObservabilityManager initializes all observability components
func NewObservabilityManager(ctx context.Context, cfg *config.Config) (*ObservabilityManager, error) {
	om := &ObservabilityManager{
		config: cfg,
	}

	var err error

	// Step 1: Create shared resource for traces and metrics
	om.resource, err = createResource(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Step 2: Initialize base logger (without Logstash initially)
	om.logger, err = createBaseLogger(cfg.Logging)
	if err != nil {
		return nil, fmt.Errorf("failed to create base logger: %w", err)
	}

	// Step 3: Initialize Tracer (OTel → Jaeger)
	if cfg.Tracing.Enabled {
		om.tracerProvider, err = InitTracer(ctx, cfg.Tracing, om.resource)
		if err != nil {
			om.cleanup(ctx)
			return nil, fmt.Errorf("failed to init tracer: %w", err)
		}
		otel.SetTracerProvider(om.tracerProvider)
		om.logger.Info("tracer initialized",
			zap.String("endpoint", cfg.Tracing.Endpoint),
			zap.Float64("sample_ratio", cfg.Tracing.SampleRatio),
		)
	}

	// Step 4: Initialize Metrics (OTel → Prometheus)
	if cfg.Metrics.Enabled {
		om.meterProvider, om.metricsServer, err = InitMetrics(ctx, cfg.Metrics, om.resource)
		if err != nil {
			om.cleanup(ctx)
			return nil, fmt.Errorf("failed to init metrics: %w", err)
		}
		otel.SetMeterProvider(om.meterProvider)

		// Start metrics server
		if err := om.metricsServer.Start(); err != nil {
			om.cleanup(ctx)
			return nil, fmt.Errorf("failed to start metrics server: %w", err)
		}

		om.logger.Info("metrics initialized",
			zap.Int("port", cfg.Metrics.Port),
			zap.String("path", cfg.Metrics.MetricsPath),
		)

		// Initialize metrics recorder
		meter := om.meterProvider.Meter(cfg.Metrics.ServiceName)
		om.metricsRecorder, err = NewMetricsRecorder(meter)
		if err != nil {
			om.cleanup(ctx)
			return nil, fmt.Errorf("failed to init metrics recorder: %w", err)
		}
	}

	// Step 5: Initialize Logstash writer (OTel → Logstash → Elasticsearch)
	if cfg.Logging.LogstashEnabled {
		logstashCore, writer, err := CreateLogstashCore(cfg.Logging, om.logger)
		if err != nil {
			om.cleanup(ctx)
			return nil, fmt.Errorf("failed to create Logstash core: %w", err)
		}

		om.logstashWriter = writer
		// Upgrade logger to include Logstash output
		om.logger = om.logger.WithOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
			return zapcore.NewTee(c, logstashCore)
		}))

		om.logger.Info("Logstash integration initialized",
			zap.String("addr", cfg.Logging.GetLogstashAddr()),
			zap.String("protocol", cfg.Logging.LogstashProtocol),
			zap.Bool("connected", writer.IsConnected()),
		)
	}

	om.logger.Info("observability manager initialized successfully",
		zap.Bool("tracing_enabled", cfg.Tracing.Enabled),
		zap.Bool("metrics_enabled", cfg.Metrics.Enabled),
		zap.Bool("logstash_enabled", cfg.Logging.LogstashEnabled),
	)

	return om, nil
}

// createResource creates a shared OpenTelemetry resource
func createResource(ctx context.Context, cfg *config.Config) (*resource.Resource, error) {
	resourceCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return resource.New(resourceCtx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.App.Name),
			semconv.ServiceVersion(cfg.App.Version),
			semconv.DeploymentEnvironment(cfg.App.Environment),
			attribute.String("service.name", cfg.App.Name),
			attribute.String("service.version", cfg.App.Version),
			attribute.String("environment", cfg.App.Environment),
		),
	)
}

// createBaseLogger creates the initial zap logger
func createBaseLogger(cfg config.LoggingConfig) (*zap.Logger, error) {
	var zapConfig zap.Config

	if cfg.Format == "json" {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
	}

	// Set log level
	level := zap.InfoLevel
	if err := level.UnmarshalText([]byte(cfg.Level)); err == nil {
		zapConfig.Level = zap.NewAtomicLevelAt(level)
	}

	// Configure output paths
	if cfg.Output == "file" {
		zapConfig.OutputPaths = []string{cfg.FilePath}
		zapConfig.ErrorOutputPaths = []string{cfg.FilePath}
	} else {
		zapConfig.OutputPaths = []string{"stdout"}
		zapConfig.ErrorOutputPaths = []string{"stderr"}
	}

	return zapConfig.Build()
}

// GetLogger returns the configured logger
func (om *ObservabilityManager) GetLogger() *zap.Logger {
	return om.logger
}

// GetMetricsRecorder returns the metrics recorder
func (om *ObservabilityManager) GetMetricsRecorder() *MetricsRecorder {
	return om.metricsRecorder
}

// GetTracerProvider returns the tracer provider
func (om *ObservabilityManager) GetTracerProvider() *sdktrace.TracerProvider {
	return om.tracerProvider
}

// GetMeterProvider returns the meter provider
func (om *ObservabilityManager) GetMeterProvider() *sdkmetric.MeterProvider {
	return om.meterProvider
}

// Shutdown gracefully stops all observability components
func (om *ObservabilityManager) Shutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	var errs []error

	om.logger.Info("shutting down observability manager")

	// Shutdown tracer
	if om.tracerProvider != nil {
		if err := om.tracerProvider.Shutdown(shutdownCtx); err != nil {
			om.logger.Error("failed to shutdown tracer", zap.Error(err))
			errs = append(errs, fmt.Errorf("tracer shutdown: %w", err))
		} else {
			om.logger.Info("tracer shut down successfully")
		}
	}

	// Shutdown meter
	if om.meterProvider != nil {
		if err := om.meterProvider.Shutdown(shutdownCtx); err != nil {
			om.logger.Error("failed to shutdown meter", zap.Error(err))
			errs = append(errs, fmt.Errorf("meter shutdown: %w", err))
		} else {
			om.logger.Info("meter shut down successfully")
		}
	}

	// Shutdown metrics server
	if om.metricsServer != nil {
		if err := om.metricsServer.Shutdown(shutdownCtx); err != nil {
			om.logger.Error("failed to shutdown metrics server", zap.Error(err))
			errs = append(errs, fmt.Errorf("metrics server shutdown: %w", err))
		} else {
			om.logger.Info("metrics server shut down successfully")
		}
	}

	// Close Logstash writer
	if om.logstashWriter != nil {
		if err := om.logstashWriter.Close(); err != nil {
			om.logger.Error("failed to close Logstash writer", zap.Error(err))
			errs = append(errs, fmt.Errorf("logstash writer close: %w", err))
		} else {
			om.logger.Info("Logstash writer closed successfully")
		}
	}

	// Sync logger (flush any buffered logs)
	if om.logger != nil {
		if err := om.logger.Sync(); err != nil {
			// Ignore sync errors on stdout/stderr (common in containers)
			errs = append(errs, fmt.Errorf("logger sync: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}

	return nil
}

// cleanup is an internal helper for initialization failures
func (om *ObservabilityManager) cleanup(ctx context.Context) {
	cleanupCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if om.tracerProvider != nil {
		om.tracerProvider.Shutdown(cleanupCtx)
	}
	if om.meterProvider != nil {
		om.meterProvider.Shutdown(cleanupCtx)
	}
	if om.metricsServer != nil {
		om.metricsServer.Shutdown(cleanupCtx)
	}
	if om.logstashWriter != nil {
		om.logstashWriter.Close()
	}
	if om.logger != nil {
		om.logger.Sync()
	}
}
