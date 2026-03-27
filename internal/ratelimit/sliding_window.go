package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sony/gobreaker"
	"github.com/souravkumar/distributed-rate-limiter/internal/common"
)

// slidingWindowScript is a Lua script for atomic sliding window rate limiting
// using a Redis sorted set.
//
// Each request is stored as a member with its timestamp as the score.
// Old entries outside the window are trimmed, and the remaining count is checked.
//
// KEYS[1] = rate limit key
// ARGV[1] = limit (max requests per window)
// ARGV[2] = window size in milliseconds
// ARGV[3] = current timestamp in unix milliseconds
// ARGV[4] = TTL in seconds for the key
//
// Returns: {allowed (0/1), remaining, reset_unix_seconds}
var slidingWindowScript = redis.NewScript(`
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window_ms = tonumber(ARGV[2])
local now_ms = tonumber(ARGV[3])
local ttl = tonumber(ARGV[4])

-- Remove entries outside the current window
local window_start = now_ms - window_ms
redis.call('ZREMRANGEBYSCORE', key, '-inf', window_start)

-- Count current requests in the window
local count = redis.call('ZCARD', key)

local allowed = 0
local remaining = limit - count
if remaining > 0 then
    -- Add this request with timestamp as score and a unique member
    redis.call('ZADD', key, now_ms, now_ms .. ':' .. math.random(1000000))
    allowed = 1
    remaining = remaining - 1
end

-- Set TTL so the key expires after the window passes
redis.call('EXPIRE', key, ttl)

-- Reset = when the oldest entry in the current window expires
local oldest = redis.call('ZRANGE', key, 0, 0, 'WITHSCORES')
local reset = math.floor(now_ms / 1000) + math.ceil(window_ms / 1000)
if #oldest >= 2 then
    reset = math.floor((tonumber(oldest[2]) + window_ms) / 1000)
end

return {allowed, remaining, reset}
`)

// SlidingWindow implements the sliding window log rate limiting algorithm using Redis.
// It uses a sorted set to track exact request timestamps within the window.
// Most accurate algorithm -- no boundary burst problem -- but uses more memory
// (one sorted set entry per request within the window).
type SlidingWindow struct {
	client *redis.Client
	cb     *gobreaker.CircuitBreaker
}

// NewSlidingWindow creates a new SlidingWindow limiter
func NewSlidingWindow(client *redis.Client, cb *gobreaker.CircuitBreaker) *SlidingWindow {
	return &SlidingWindow{client: client, cb: cb}
}

// Allow checks if a request is allowed under the sliding window rate limit
func (sw *SlidingWindow) Allow(ctx context.Context, cfg LimitConfig) (*CheckResponse, error) {
	windowMs := cfg.Window.Milliseconds()
	if windowMs <= 0 {
		return nil, fmt.Errorf("invalid window duration")
	}

	key := fmt.Sprintf("%s%s:%s:%s", common.RateLimitKeyPrefix, cfg.TenantID, cfg.Profile, cfg.UserKey)

	ttlSeconds := int(cfg.Window.Seconds() * 2)
	if ttlSeconds < 1 {
		ttlSeconds = 1
	}

	nowMs := time.Now().UnixMilli()

	raw, err := sw.cb.Execute(func() (interface{}, error) {
		return slidingWindowScript.Run(ctx, sw.client, []string{key},
			cfg.Limit,
			windowMs,
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
