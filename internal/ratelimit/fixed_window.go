package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sony/gobreaker"
	"github.com/souravkumar/distributed-rate-limiter/internal/common"
)

// fixedWindowScript is a Lua script for atomic fixed window rate limiting.
//
// The window is defined by flooring the current time to the nearest window boundary.
// A single counter tracks requests within that window and resets naturally via TTL.
//
// KEYS[1] = rate limit key (includes window start for natural expiry)
// ARGV[1] = limit (max requests per window)
// ARGV[2] = window size in seconds (used as TTL)
// ARGV[3] = current timestamp in unix seconds
//
// Returns: {allowed (0/1), remaining, reset_unix_seconds}
var fixedWindowScript = redis.NewScript(`
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

local count = redis.call('INCR', key)

-- Set TTL on first request in this window
if count == 1 then
    redis.call('EXPIRE', key, window)
end

local allowed = 0
local remaining = limit - count
if remaining >= 0 then
    allowed = 1
else
    remaining = 0
end

-- Reset = start of current window + window size
local window_start = math.floor(now / window) * window
local reset = window_start + window

return {allowed, remaining, reset}
`)

// FixedWindow implements the fixed window rate limiting algorithm using Redis.
// It divides time into fixed intervals and counts requests per interval.
// Simple and memory-efficient, but allows up to 2x burst at window boundaries.
type FixedWindow struct {
	client *redis.Client
	cb     *gobreaker.CircuitBreaker
}

// NewFixedWindow creates a new FixedWindow limiter
func NewFixedWindow(client *redis.Client, cb *gobreaker.CircuitBreaker) *FixedWindow {
	return &FixedWindow{client: client, cb: cb}
}

// Allow checks if a request is allowed under the fixed window rate limit
func (fw *FixedWindow) Allow(ctx context.Context, cfg LimitConfig) (*CheckResponse, error) {
	windowSeconds := int(cfg.Window.Seconds())
	if windowSeconds <= 0 {
		return nil, fmt.Errorf("invalid window duration")
	}

	now := time.Now().Unix()

	// Include window start in key so each window gets its own counter
	windowStart := (now / int64(windowSeconds)) * int64(windowSeconds)
	key := fmt.Sprintf("%s%s:%s:%s:%d", common.RateLimitKeyPrefix, cfg.TenantID, cfg.Profile, cfg.UserKey, windowStart)

	raw, err := fw.cb.Execute(func() (interface{}, error) {
		return fixedWindowScript.Run(ctx, fw.client, []string{key},
			cfg.Limit,
			windowSeconds,
			now,
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
