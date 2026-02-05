package config

import (
	"fmt"
	"github.com/spf13/viper"
	"strings"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Server         ServerConfig         `mapstructure:"server"`
	Postgres       PostgresConfig       `mapstructure:"postgres"`
	Redis          RedisConfig          `mapstructure:"redis"`
	Cache          CacheConfig          `mapstructure:"cache"`
	Quotas         QuotasConfig         `mapstructure:"quotas"`
	CircuitBreaker CircuitBreakerConfig `mapstructure:"circuit_breaker"`
	Logging        LoggingConfig        `mapstructure:"logging"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// PostgresConfig holds PostgreSQL configuration
type PostgresConfig struct {
	Host           string `mapstructure:"host"`
	Port           int    `mapstructure:"port"`
	User           string `mapstructure:"user"`
	Password       string `mapstructure:"password"`
	Database       string `mapstructure:"database"`
	MaxConnections int    `mapstructure:"max_connections"`
}

// GetDSN returns the PostgreSQL connection string
func (p *PostgresConfig) GetDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		p.User, p.Password, p.Host, p.Port, p.Database)
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	PoolSize int    `mapstructure:"pool_size"`
	Timeout  string `mapstructure:"timeout"`
}

// GetTimeout returns the timeout as time.Duration
func (r *RedisConfig) GetTimeout() time.Duration {
	duration, err := time.ParseDuration(r.Timeout)
	if err != nil {
		return time.Second
	}
	return duration
}

// CacheConfig holds caching configuration
type CacheConfig struct {
	TenantTTL  string `mapstructure:"tenant_ttl"`
	ProfileTTL string `mapstructure:"profile_ttl"`
}

// GetTenantTTL returns tenant cache TTL as time.Duration
func (c *CacheConfig) GetTenantTTL() time.Duration {
	duration, err := time.ParseDuration(c.TenantTTL)
	if err != nil {
		return 5 * time.Minute
	}
	return duration
}

// GetProfileTTL returns profile cache TTL as time.Duration
func (c *CacheConfig) GetProfileTTL() time.Duration {
	duration, err := time.ParseDuration(c.ProfileTTL)
	if err != nil {
		return 5 * time.Minute
	}
	return duration
}

// QuotasConfig holds default quota settings for new tenants
type QuotasConfig struct {
	DefaultMaxProfiles       int `mapstructure:"default_max_profiles"`
	DefaultMaxRequestsPerDay int `mapstructure:"default_max_requests_per_day"`
}

// CircuitBreakerConfig holds circuit breaker configuration
type CircuitBreakerConfig struct {
	Enabled             bool   `mapstructure:"enabled"`
	FailureThreshold    uint32 `mapstructure:"failure_threshold"`
	Timeout             string `mapstructure:"timeout"`
	HalfOpenMaxRequests uint32 `mapstructure:"half_open_max_requests"`
}

// GetTimeout returns the timeout as time.Duration
func (c *CircuitBreakerConfig) GetTimeout() time.Duration {
	duration, err := time.ParseDuration(c.Timeout)
	if err != nil {
		return 10 * time.Second
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
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)

	// PostgreSQL defaults
	v.SetDefault("postgres.host", "localhost")
	v.SetDefault("postgres.port", 5432)
	v.SetDefault("postgres.user", "ratelimiter")
	v.SetDefault("postgres.password", "secret")
	v.SetDefault("postgres.database", "ratelimiter")
	v.SetDefault("postgres.max_connections", 20)

	// Redis defaults
	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.pool_size", 10)
	v.SetDefault("redis.timeout", "1s")

	// Cache defaults
	v.SetDefault("cache.tenant_ttl", "5m")
	v.SetDefault("cache.profile_ttl", "5m")

	// Quotas defaults
	v.SetDefault("quotas.default_max_profiles", 10)
	v.SetDefault("quotas.default_max_requests_per_day", 10000)

	// Circuit breaker defaults
	v.SetDefault("circuit_breaker.enabled", true)
	v.SetDefault("circuit_breaker.failure_threshold", 5)
	v.SetDefault("circuit_breaker.timeout", "10s")
	v.SetDefault("circuit_breaker.half_open_max_requests", 3)

	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
}

// Validate checks the configuration for errors
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.Postgres.Host == "" {
		return fmt.Errorf("postgres host is required")
	}

	if c.Redis.Addr == "" {
		return fmt.Errorf("redis address is required")
	}

	if c.Quotas.DefaultMaxProfiles <= 0 {
		return fmt.Errorf("default_max_profiles must be positive")
	}

	if c.Quotas.DefaultMaxRequestsPerDay <= 0 {
		return fmt.Errorf("default_max_requests_per_day must be positive")
	}

	return nil
}

// GetAddress returns the server address in host:port format
func (c *Config) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}
