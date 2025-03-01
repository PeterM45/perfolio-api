package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/PeterM45/perfolio-api/cmd/api/app"
	"github.com/PeterM45/perfolio-api/internal/common/config"
	"github.com/PeterM45/perfolio-api/pkg/logger"
)

func main() {
	// Parse command line flags
	configFile := flag.String("config", "./configs/config.yaml", "Path to config file")
	flag.Parse()

	// Initialize logger
	log := logger.NewZapLogger("info")
	defer log.Sync()

	// Load configuration
	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Create application
	application, err := app.New(cfg, log)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create application")
	}

	// Start server in a goroutine
	go func() {
		log.Info().Int("port", cfg.Server.Port).Msg("Starting Perfolio API server")
		if err := application.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	s := <-quit

	log.Info().Str("signal", s.String()).Msg("Shutting down server...")

	// Create context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := application.Stop(ctx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
		os.Exit(1)
	}

	log.Info().Msg("Server gracefully stopped")
}
