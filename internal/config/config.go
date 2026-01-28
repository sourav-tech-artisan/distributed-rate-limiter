package config

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Profile naming rules
const (
	MaxProfileNameLength = 64
	MaxKeyLength         = 256
)

var profileNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

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

// RateLimiterConfig holds rate limiter configuration with profiles
type RateLimiterConfig struct {
	DefaultProfile string             `mapstructure:"default_profile"`
	Profiles       map[string]Profile `mapstructure:"profiles"`
}

// Profile represents a rate limiting configuration for a specific use case
type Profile struct {
	Algorithm   string                   `mapstructure:"algorithm"`
	Limit       int                      `mapstructure:"limit"`
	Window      string                   `mapstructure:"window"`
	TokenBucket *TokenBucketProfileConfig `mapstructure:"token_bucket"`
}

// GetWindow returns the window as time.Duration
func (p *Profile) GetWindow() time.Duration {
	duration, err := time.ParseDuration(p.Window)
	if err != nil {
		return time.Minute
	}
	return duration
}

// GetCapacity returns the token bucket capacity, defaulting to limit if not set
func (p *Profile) GetCapacity() int {
	if p.TokenBucket != nil && p.TokenBucket.Capacity > 0 {
		return p.TokenBucket.Capacity
	}
	return p.Limit
}

// GetRefillRate returns tokens per second based on limit and window
func (p *Profile) GetRefillRate() float64 {
	window := p.GetWindow()
	if window == 0 {
		return 0
	}
	return float64(p.Limit) / window.Seconds()
}

// GetKeyTTL returns the TTL for Redis keys (3x window duration)
func (p *Profile) GetKeyTTL() time.Duration {
	return p.GetWindow() * 3
}

// TokenBucketProfileConfig holds token bucket specific configuration within a profile
type TokenBucketProfileConfig struct {
	Capacity int `mapstructure:"capacity"`
}

// CircuitBreakerConfig holds circuit breaker configuration
type CircuitBreakerConfig struct {
	Enabled             bool   `mapstructure:"enabled"`
	FailureThreshold    uint32 `mapstructure:"failure_threshold"`
	Timeout             string `mapstructure:"timeout"`
	HalfOpenMaxRequests uint32 `mapstructure:"half_open_max_requests"`
	FailMode            string `mapstructure:"fail_mode"`
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

	// Redis defaults
	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.pool_size", 10)
	v.SetDefault("redis.timeout", "1s")

	// Rate limiter defaults
	v.SetDefault("rate_limiter.default_profile", "default")

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

// Validate checks the configuration for errors
func (c *Config) Validate() error {
	// Check that at least one profile exists
	if len(c.RateLimiter.Profiles) == 0 {
		return fmt.Errorf("at least one rate limiter profile must be defined")
	}

	// Validate default profile exists
	if _, exists := c.RateLimiter.Profiles[c.RateLimiter.DefaultProfile]; !exists {
		return fmt.Errorf("default profile '%s' not found in profiles", c.RateLimiter.DefaultProfile)
	}

	// Validate each profile
	for name, profile := range c.RateLimiter.Profiles {
		if err := validateProfileName(name); err != nil {
			return fmt.Errorf("invalid profile name '%s': %w", name, err)
		}
		if err := validateProfile(name, &profile); err != nil {
			return err
		}
	}

	return nil
}

func validateProfileName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("profile name cannot be empty")
	}
	if len(name) > MaxProfileNameLength {
		return fmt.Errorf("profile name exceeds %d characters", MaxProfileNameLength)
	}
	if !profileNameRegex.MatchString(name) {
		return fmt.Errorf("profile name must contain only alphanumeric characters and underscores")
	}
	return nil
}

func validateProfile(name string, p *Profile) error {
	if p.Algorithm == "" {
		return fmt.Errorf("profile '%s': algorithm is required", name)
	}
	if p.Algorithm != "token_bucket" {
		return fmt.Errorf("profile '%s': unsupported algorithm '%s' (only 'token_bucket' is supported)", name, p.Algorithm)
	}
	if p.Limit <= 0 {
		return fmt.Errorf("profile '%s': limit must be positive", name)
	}
	if p.Window == "" {
		return fmt.Errorf("profile '%s': window is required", name)
	}
	if _, err := time.ParseDuration(p.Window); err != nil {
		return fmt.Errorf("profile '%s': invalid window duration '%s'", name, p.Window)
	}
	return nil
}

// GetAddress returns the server address in host:port format
func (c *Config) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// GetProfile returns a profile by name, or nil if not found
func (c *Config) GetProfile(name string) *Profile {
	if profile, exists := c.RateLimiter.Profiles[name]; exists {
		return &profile
	}
	return nil
}

// GetDefaultProfile returns the default profile
func (c *Config) GetDefaultProfile() *Profile {
	return c.GetProfile(c.RateLimiter.DefaultProfile)
}

// ProfileExists checks if a profile exists
func (c *Config) ProfileExists(name string) bool {
	_, exists := c.RateLimiter.Profiles[name]
	return exists
}

// GetProfileNames returns all profile names
func (c *Config) GetProfileNames() []string {
	names := make([]string, 0, len(c.RateLimiter.Profiles))
	for name := range c.RateLimiter.Profiles {
		names = append(names, name)
	}
	return names
}
