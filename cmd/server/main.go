package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/souravkumar/distributed-rate-limiter/internal/api"
	"github.com/souravkumar/distributed-rate-limiter/internal/config"
	"github.com/souravkumar/distributed-rate-limiter/internal/logger"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func main() {
	// Load configuration
	cfg, err := config.Load("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger.Init(cfg.Logging.Level, cfg.Logging.Format)
	log.Info().Msg("starting rate limiter service")

	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	// Initialize handlers
	handler := api.NewHandler()

	// Setup router
	router := api.SetupRouter(handler)

	// Create HTTP server
	srv := &http.Server{
		Addr:    cfg.GetAddress(),
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		log.Info().Str("address", cfg.GetAddress()).Msg("server starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("failed to start server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down server...")

	// Graceful shutdown with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("server forced to shutdown")
	} else {
		log.Info().Msg("server exited gracefully")
	}
}
