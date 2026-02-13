package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/williandandrade/address-validation-service/internal/api"
	"github.com/williandandrade/address-validation-service/internal/api/handler"
	"github.com/williandandrade/address-validation-service/internal/api/middleware"
	"github.com/williandandrade/address-validation-service/internal/infra/config"
	"github.com/williandandrade/address-validation-service/internal/usecase"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Setup logger
	logger := setupLogger(cfg)
	slog.SetDefault(logger)

	// Log startup
	logger.Info("starting address validation service",
		slog.String("env", cfg.Env),
		slog.String("port", cfg.Port),
	)

	// Create usecases
	validateUsecase := usecase.NewValidateUsecase()

	// Create handlers
	healthHandler := handler.NewHealthHandler()
	validateHandler := handler.NewValidateHandler(validateUsecase)

	// Create router
	router := api.NewRouter(healthHandler, validateHandler)

	// Build handler with middleware (applied in reverse order)
	httpHandler := router.Handler()
	httpHandler = middleware.Logging(logger)(httpHandler)

	// Create server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      httpHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("server listening", slog.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", slog.String("error", err.Error()))
		cancel() // Ensure context is canceled
		os.Exit(1)
	}

	logger.Info("server stopped")
}

// setupLogger creates a structured logger based on configuration.
func setupLogger(cfg *config.Config) *slog.Logger {
	var level slog.Level
	switch cfg.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if cfg.LogFormat == "json" || cfg.IsProduction() {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

func init() {
	// Ensure we fail fast if PORT is explicitly set but invalid
	if port := os.Getenv("PORT"); port != "" {
		if _, err := fmt.Sscanf(port, "%d", new(int)); err != nil {
			fmt.Fprintf(os.Stderr, "invalid PORT: %s\n", port)
			os.Exit(1)
		}
	}
}
