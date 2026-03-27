package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sony/gobreaker"
	"github.com/souravkumar/distributed-rate-limiter/internal/common"
)

// leakyBucketScript is a Lua script for atomic leaky bucket rate limiting.
//
// The leaky bucket works as a queue: requests fill the bucket, and it drains
// at a constant rate (limit / window). If the bucket is full, the request is rejected.
// This produces the smoothest output rate with no bursting.
//
// KEYS[1] = rate limit key
// ARGV[1] = capacity (max queue size)
// ARGV[2] = drain rate (requests per second)
// ARGV[3] = current timestamp in unix milliseconds
// ARGV[4] = TTL in seconds for the key
//
// Returns: {allowed (0/1), remaining_capacity, reset_unix_seconds}
var leakyBucketScript = redis.NewScript(`
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local drain_rate = tonumber(ARGV[2])
local now_ms = tonumber(ARGV[3])
local ttl = tonumber(ARGV[4])

-- Get current bucket state
local bucket = redis.call('HMGET', key, 'level', 'last_drain_ms')
local level = tonumber(bucket[1])
local last_drain_ms = tonumber(bucket[2])

-- Initialize bucket if it doesn't exist
if level == nil then
    level = 0
    last_drain_ms = now_ms
end

-- Drain the bucket based on elapsed time
local elapsed_seconds = (now_ms - last_drain_ms) / 1000
local drained = elapsed_seconds * drain_rate
level = math.max(0, level - drained)
last_drain_ms = now_ms

-- Try to add request to the bucket
local allowed = 0
if level < capacity then
    level = level + 1
    allowed = 1
end

-- remaining = how many more requests can fit
local remaining = math.max(0, math.floor(capacity - level))

-- reset = when the bucket will be fully drained
local now_seconds = math.floor(now_ms / 1000)
local reset = now_seconds
if level > 0 and drain_rate > 0 then
    reset = now_seconds + math.ceil(level / drain_rate)
end

-- Save state
redis.call('HMSET', key, 'level', level, 'last_drain_ms', last_drain_ms)
redis.call('EXPIRE', key, ttl)

return {allowed, remaining, reset}
`)

// LeakyBucket implements the leaky bucket rate limiting algorithm using Redis.
// It models a queue that drains at a constant rate. Unlike token bucket which
// allows bursts up to capacity, leaky bucket enforces a smooth, constant output rate.
type LeakyBucket struct {
	client *redis.Client
	cb     *gobreaker.CircuitBreaker
}

// NewLeakyBucket creates a new LeakyBucket limiter
func NewLeakyBucket(client *redis.Client, cb *gobreaker.CircuitBreaker) *LeakyBucket {
	return &LeakyBucket{client: client, cb: cb}
}

// Allow checks if a request is allowed under the leaky bucket rate limit
func (lb *LeakyBucket) Allow(ctx context.Context, cfg LimitConfig) (*CheckResponse, error) {
	windowSeconds := cfg.Window.Seconds()
	if windowSeconds <= 0 {
		return nil, fmt.Errorf("invalid window duration")
	}

	drainRate := float64(cfg.Limit) / windowSeconds

	key := fmt.Sprintf("%s%s:%s:%s", common.RateLimitKeyPrefix, cfg.TenantID, cfg.Profile, cfg.UserKey)

	ttlSeconds := int(windowSeconds * 3)
	if ttlSeconds < 1 {
		ttlSeconds = 1
	}

	nowMs := time.Now().UnixMilli()

	raw, err := lb.cb.Execute(func() (interface{}, error) {
		return leakyBucketScript.Run(ctx, lb.client, []string{key},
			cfg.Capacity,
			drainRate,
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

	if !allowed {
		retryAfter := resetUnix - time.Now().Unix()
		if retryAfter < 1 {
			retryAfter = 1
		}
		resp.RetryAfter = retryAfter
	}

	return resp, nil
}
