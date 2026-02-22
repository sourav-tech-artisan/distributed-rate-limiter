package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sony/gobreaker"
	"github.com/souravkumar/distributed-rate-limiter/internal/common"
)

// tokenBucketScript is a Lua script for atomic token bucket operations.
// It handles refilling tokens based on elapsed time and consuming one token.
// This runs atomically in Redis, making it safe for distributed use.
//
// KEYS[1] = rate limit key
// ARGV[1] = capacity (max tokens)
// ARGV[2] = refill rate (tokens per second)
// ARGV[3] = current timestamp (unix milliseconds)
// ARGV[4] = TTL in seconds for the key
//
// Returns: {allowed (0/1), remaining_tokens, reset_unix_seconds}
var tokenBucketScript = redis.NewScript(`
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local refill_rate = tonumber(ARGV[2])
local now_ms = tonumber(ARGV[3])
local ttl = tonumber(ARGV[4])

-- Get current bucket state
local bucket = redis.call('HMGET', key, 'tokens', 'last_refill_ms')
local tokens = tonumber(bucket[1])
local last_refill_ms = tonumber(bucket[2])

-- Initialize bucket if it doesn't exist
if tokens == nil then
    tokens = capacity
    last_refill_ms = now_ms
end

-- Calculate token refill: convert elapsed ms to seconds for rate calc
local elapsed_seconds = (now_ms - last_refill_ms) / 1000
local refill = elapsed_seconds * refill_rate
tokens = math.min(capacity, tokens + refill)
last_refill_ms = now_ms

-- Try to consume one token
local allowed = 0
if tokens >= 1 then
    tokens = tokens - 1
    allowed = 1
end

-- Calculate reset time (when bucket will be full again) in unix seconds
local tokens_needed = capacity - tokens
local now_seconds = math.floor(now_ms / 1000)
local reset = now_seconds
if tokens_needed > 0 and refill_rate > 0 then
    reset = now_seconds + math.ceil(tokens_needed / refill_rate)
end

-- Save updated state (timestamps in ms for precision)
redis.call('HMSET', key, 'tokens', tokens, 'last_refill_ms', last_refill_ms)
redis.call('EXPIRE', key, ttl)

return {allowed, math.floor(tokens), reset}
`)

// TokenBucket implements the token bucket rate limiting algorithm using Redis
type TokenBucket struct {
	client *redis.Client
	cb     *gobreaker.CircuitBreaker
}

// NewTokenBucket creates a new TokenBucket limiter
func NewTokenBucket(client *redis.Client, cb *gobreaker.CircuitBreaker) *TokenBucket {
	return &TokenBucket{client: client, cb: cb}
}

// BucketConfig holds the configuration for a rate limit check
type BucketConfig struct {
	TenantID string
	Profile  string
	UserKey  string
	Capacity int
	Limit    int
	Window   time.Duration
}

// Allow checks if a request is allowed under the rate limit
func (tb *TokenBucket) Allow(ctx context.Context, cfg BucketConfig) (*CheckResponse, error) {
	// Build Redis key: rl:{tenant_id}:{profile}:{user_key}
	key := fmt.Sprintf("%s%s:%s:%s", common.RateLimitKeyPrefix, cfg.TenantID, cfg.Profile, cfg.UserKey)

	// Calculate refill rate (tokens per second)
	windowSeconds := cfg.Window.Seconds()
	if windowSeconds <= 0 {
		return nil, fmt.Errorf("invalid window duration")
	}
	refillRate := float64(cfg.Limit) / windowSeconds

	// TTL = 3x window to keep state around for a while
	ttlSeconds := int(windowSeconds * 3)
	if ttlSeconds < 1 {
		ttlSeconds = 1
	}

	nowMs := time.Now().UnixMilli()

	// Execute Lua script atomically, protected by circuit breaker
	raw, err := tb.cb.Execute(func() (interface{}, error) {
		return tokenBucketScript.Run(ctx, tb.client, []string{key},
			cfg.Capacity,
			refillRate,
			nowMs,
			ttlSeconds,
		).Int64Slice()
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute rate limit check: %w", err)
	}

	result := raw.([]int64)

	allowed := result[0] == 1
	remaining := int(result[1])
	resetUnix := result[2]

	resp := &CheckResponse{
		Allowed:   allowed,
		Limit:     cfg.Limit,
		Remaining: remaining,
		Reset:     resetUnix,
	}

	// Add retry_after in seconds if rate limited
	if !allowed {
		retryAfter := resetUnix - time.Now().Unix()
		if retryAfter < 1 {
			retryAfter = 1
		}
		resp.RetryAfter = retryAfter
	}

	return resp, nil
}
