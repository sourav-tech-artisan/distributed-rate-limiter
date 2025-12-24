package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server         ServerConfig         `mapstructure:"server"`
	Redis          RedisConfig          `mapstructure:"redis"`
	RateLimiter    RateLimiterConfig    `mapstructure:"rate_limiter"`
	CircuitBreaker CircuitBreakerConfig `mapstructure:"circuit_breaker"`
	Logging        LoggingConfig        `mapstructure:"logging"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	PoolSize int    `mapstructure:"pool_size"`
	Timeout  string `mapstructure:"timeout"` // Duration as string (e.g., "1s", "100ms")
}

// GetTimeout returns the timeout as time.Duration
func (r *RedisConfig) GetTimeout() time.Duration {
	duration, err := time.ParseDuration(r.Timeout)
	if err != nil {
		return time.Second // Default to 1 second if parsing fails
	}
	return duration
}

// RateLimiterConfig holds rate limiter configuration
type RateLimiterConfig struct {
	DefaultAlgorithm string            `mapstructure:"default_algorithm"`
	DefaultLimit     int               `mapstructure:"default_limit"`
	DefaultWindow    string            `mapstructure:"default_window"` // Duration as string (e.g., "1m", "5s")
	TokenBucket      TokenBucketConfig `mapstructure:"token_bucket"`
}

// GetDefaultWindow returns the default window as time.Duration
func (r *RateLimiterConfig) GetDefaultWindow() time.Duration {
	duration, err := time.ParseDuration(r.DefaultWindow)
	if err != nil {
		return time.Minute // Default to 1 minute if parsing fails
	}
	return duration
}

// TokenBucketConfig holds token bucket specific configuration
type TokenBucketConfig struct {
	DefaultCapacity     int    `mapstructure:"default_capacity"`
	RefillCheckInterval string `mapstructure:"refill_check_interval"` // Duration as string (e.g., "1s")
}

// GetRefillCheckInterval returns the refill check interval as time.Duration
func (t *TokenBucketConfig) GetRefillCheckInterval() time.Duration {
	duration, err := time.ParseDuration(t.RefillCheckInterval)
	if err != nil {
		return time.Second // Default to 1 second if parsing fails
	}
	return duration
}

// CircuitBreakerConfig holds circuit breaker configuration
type CircuitBreakerConfig struct {
	Enabled             bool   `mapstructure:"enabled"`
	FailureThreshold    uint32 `mapstructure:"failure_threshold"`
	Timeout             string `mapstructure:"timeout"` // Duration as string (e.g., "10s")
	HalfOpenMaxRequests uint32 `mapstructure:"half_open_max_requests"`
	FailMode            string `mapstructure:"fail_mode"`
}

// GetTimeout returns the timeout as time.Duration
func (c *CircuitBreakerConfig) GetTimeout() time.Duration {
	duration, err := time.ParseDuration(c.Timeout)
	if err != nil {
		return 10 * time.Second // Default to 10 seconds if parsing fails
	}
	return duration
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// Load reads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Set config file
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
	}

	// Enable environment variables
	v.SetEnvPrefix("")
	v.AutomaticEnv()

	// Read config file (optional - env vars can override)
	if err := v.ReadInConfig(); err != nil {
		// Config file not found is okay if we have env vars
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)

	// Redis defaults
	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.pool_size", 10)
	v.SetDefault("redis.timeout", "1s")

	// Rate limiter defaults
	v.SetDefault("rate_limiter.default_algorithm", "token_bucket")
	v.SetDefault("rate_limiter.default_limit", 100)
	v.SetDefault("rate_limiter.default_window", "1m")
	v.SetDefault("rate_limiter.token_bucket.default_capacity", 100)
	v.SetDefault("rate_limiter.token_bucket.refill_check_interval", "1s")

	// Circuit breaker defaults
	v.SetDefault("circuit_breaker.enabled", true)
	v.SetDefault("circuit_breaker.failure_threshold", 5)
	v.SetDefault("circuit_breaker.timeout", "10s")
	v.SetDefault("circuit_breaker.half_open_max_requests", 3)
	v.SetDefault("circuit_breaker.fail_mode", "closed")

	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
}

// GetAddress returns the server address in host:port format
func (c *Config) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}
