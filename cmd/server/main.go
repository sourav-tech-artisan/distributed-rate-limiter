package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/souravkumar/distributed-rate-limiter/internal/cache"
	"github.com/souravkumar/distributed-rate-limiter/internal/config"
	"github.com/souravkumar/distributed-rate-limiter/internal/database"
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

	// Initialize database
	db, err := database.NewPostgres(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer database.Close(db)

	// Initialize Redis
	redisClient, err := cache.NewRedis(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to redis")
	}
	defer cache.Close(redisClient)

	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	// Setup minimal router
	router := gin.New()
	router.Use(gin.Recovery())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":   "healthy",
			"postgres": "connected",
			"redis":    "connected",
		})
	})

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
