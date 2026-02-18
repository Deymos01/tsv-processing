package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/Deymos01/tsv-processing/internal/config"
	"github.com/Deymos01/tsv-processing/internal/generator"
	"github.com/Deymos01/tsv-processing/internal/repository/postgres"
	httpserver "github.com/Deymos01/tsv-processing/internal/transport/http"
	"github.com/Deymos01/tsv-processing/internal/transport/http/handler"
	"github.com/Deymos01/tsv-processing/internal/usecase"
	"github.com/Deymos01/tsv-processing/internal/worker"
	"github.com/Deymos01/tsv-processing/pkg/logger"
)

func main() {
	// Configuration
	cfgPath := envOrDefault("CONFIG_PATH", "config.yaml")

	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Logger
	log, err := logger.New(cfg.App.Env)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	// Context
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Database
	dbPool, err := postgres.NewPool(ctx, cfg.Database)
	if err != nil {
		log.Fatal("failed to connect to database", zap.Error(err))
	}
	defer dbPool.Close()

	log.Info("database connection established")

	// Migrations
	if err = postgres.RunMigrations(cfg.Database.DSN(), "./migrations"); err != nil {
		log.Fatal("failed to run migrations", zap.Error(err))
	}

	log.Info("migrations applied")

	// Repositories
	messageRepo := postgres.NewMessageRepo(dbPool)
	fileRepo := postgres.NewFileRepo(dbPool)

	// Output generator
	rtfGen, err := generator.NewRTFGenerator(cfg.Files.OutputDir, log)
	if err != nil {
		log.Fatal("failed to init rtf generator", zap.Error(err))
	}

	// Use cases
	messageUC := usecase.NewMessageUseCase(messageRepo, log)
	fileUC := usecase.NewFileUseCase(fileRepo, messageRepo, rtfGen, log)

	// Worker pool
	processor := worker.NewProcessor(fileUC, log)
	pool := worker.NewPool(cfg.Worker.PoolSize, cfg.Worker.QueueSize, processor, log)
	scanner := worker.NewScanner(cfg.Files.InputDir, cfg.Worker.ScanInterval, pool, fileUC, log)

	go pool.Run(ctx)
	go scanner.Run(ctx)

	// HTTP server
	messageHandler := handler.NewMessageHandler(messageUC, log)
	srv := httpserver.NewServer(cfg.Server, messageHandler, log)

	// Start HTTP server in background.
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- srv.Run()
	}()

	log.Info("application started")

	// Graceful shutdown
	select {
	case <-ctx.Done():
		log.Info("shutdown signal received")
	case err = <-serverErr:
		if err != nil && !errors.Is(err, context.Canceled) {
			log.Error("http server error", zap.Error(err))
		}
	}

	// Give in-flight HTTP requests time to complete.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err = srv.Shutdown(shutdownCtx); err != nil {
		log.Error("http server shutdown error", zap.Error(err))
	}

	// Signal scanner to stop (cancel was already called via signal.NotifyContext).
	// Wait for in-flight jobs to drain.
	pool.Shutdown()

	log.Info("application stopped")
}

// envOrDefault returns the value of an environment variable or a fallback string.
func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
