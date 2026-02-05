package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/souravkumar/distributed-rate-limiter/internal/platform/config"
)

// New creates a new Redis client
func New(cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		PoolSize: cfg.Redis.PoolSize,
	})

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	log.Info().Str("addr", cfg.Redis.Addr).Msg("redis connected")
	return client, nil
}

// Close closes the Redis client
func Close(client *redis.Client) error {
	return client.Close()
}
