package common

// Authentication constants
const (
	// BcryptCost is the cost factor for bcrypt password hashing
	BcryptCost = 12

	// APIKeyPrefix is prepended to all API keys for easy identification
	APIKeyPrefix = "rl_"

	// APIKeyBytes is the number of random bytes for API key generation
	// 32 bytes = 64 hex characters
	APIKeyBytes = 32
)

// Redis key prefixes
const (
	// CacheTenantKeyPrefix is the Redis key prefix for tenant cache
	CacheTenantKeyPrefix = "cache:tenant:"

	// CacheProfileKeyPrefix is the Redis key prefix for profile cache
	CacheProfileKeyPrefix = "cache:profile:"

	// RateLimitKeyPrefix is the Redis key prefix for rate limit state
	RateLimitKeyPrefix = "rl:"

	// UsageKeyPrefix is the Redis key prefix for daily usage counters
	UsageKeyPrefix = "usage:"
)

// Algorithm represents a rate limiting algorithm type
type Algorithm string

const (
	AlgorithmTokenBucket   Algorithm = "token_bucket"
	AlgorithmSlidingWindow Algorithm = "sliding_window" // future
	AlgorithmFixedWindow   Algorithm = "fixed_window"   // future
	AlgorithmLeakyBucket   Algorithm = "leaky_bucket"   // future
)

// IsValid checks if the algorithm is a known/supported type
func (a Algorithm) IsValid() bool {
	switch a {
	case AlgorithmTokenBucket:
		return true
	default:
		return false
	}
}

// String returns the string representation of the algorithm
func (a Algorithm) String() string {
	return string(a)
}
