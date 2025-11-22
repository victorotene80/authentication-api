package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"authentication/internal/infrastructure/bootstrap"
	infraDB "authentication/internal/infrastructure/persistence"
	"authentication/shared/config"
	"authentication/shared/logging"

	"go.uber.org/zap"
)

func main() {
	// --------------------------
	// 1. Load Config
	// --------------------------
	cfg, err := config.Load()
	if err != nil {
		panic("❌ Failed to load config: " + err.Error())
	}

	// --------------------------
	// 2. Initialize Logger
	// --------------------------
	if err := logging.Initialize(cfg.App.Environment); err != nil {
		panic("❌ Failed to initialize logger: " + err.Error())
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := logging.Get().With(
		zap.String("service", "authentication"),
		zap.String("environment", cfg.App.Environment),
	)
	logger.Info(ctx, "Configuration loaded")

	db, err := infraDB.NewDatabase(cfg.Database)
	if err != nil {
		logger.Fatal(ctx, "Failed to connect to database", zap.Error(err))
	}
	defer infraDB.CloseDatabase(db)

	infraDB.CollectDBPoolMetrics(db)
	infraDB.CollectSystemMetrics()

	container, err := bootstrap.NewContainer(ctx, cfg, db)
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize DI container", zap.Error(err))
	}

	srv := container.HTTPServer()
	go func() {
		logger.Info(ctx, "HTTP server starting", zap.String("address", cfg.Server.GetServerAddr()))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal(ctx, "HTTP server failed", zap.Error(err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logger.Info(ctx, "Shutdown signal received")

	shutdownCtx, cancelShutdown := context.WithTimeout(ctx, 15*time.Second)
	defer cancelShutdown()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error(shutdownCtx, "HTTP server shutdown error", zap.Error(err))
	}

	// Stop background processors
	//container.OutboxProcessor.Stop()

	logger.Info(ctx, "Server stopped successfully")
}
