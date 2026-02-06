package main

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/souravkumar/distributed-rate-limiter/internal/auth"
	"github.com/souravkumar/distributed-rate-limiter/internal/platform/config"
	"github.com/souravkumar/distributed-rate-limiter/internal/platform/logger"
	"github.com/souravkumar/distributed-rate-limiter/internal/platform/postgres"
	"github.com/souravkumar/distributed-rate-limiter/internal/platform/redis"
	"github.com/souravkumar/distributed-rate-limiter/internal/platform/server"
	"github.com/souravkumar/distributed-rate-limiter/internal/profile"
	"github.com/souravkumar/distributed-rate-limiter/internal/ratelimit"
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
	db, err := postgres.New(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer postgres.Close(db)

	// Initialize Redis
	redisClient, err := redis.New(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to redis")
	}
	defer redis.Close(redisClient)

	// Initialize auth feature
	authRepo := auth.NewRepository(db)
	authService := auth.NewService(authRepo, cfg, log.Logger)
	authHandler := auth.NewHandler(authService, log.Logger)

	// Initialize profile feature
	profileRepo := profile.NewRepository(db)
	profileService := profile.NewService(profileRepo, log.Logger)
	profileHandler := profile.NewHandler(profileService, log.Logger)

	// Initialize rate limit feature
	limiter := ratelimit.NewTokenBucket(redisClient)
	rateLimitService := ratelimit.NewService(limiter, profileRepo, log.Logger)
	rateLimitHandler := ratelimit.NewHandler(rateLimitService, log.Logger)

	// Setup router
	router := server.NewRouter(&server.RouterConfig{
		AuthHandler:      authHandler,
		AuthMiddleware:   server.AuthMiddleware(authRepo),
		ProfileHandler:   profileHandler,
		RateLimitHandler: rateLimitHandler,
	})

	// Create and start server
	srv := server.New(cfg, router)
	if err := srv.Start(); err != nil {
		log.Fatal().Err(err).Msg("server failed")
	}
}
