package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/souravkumar/distributed-rate-limiter/internal/common"
)

// Limiter defines the interface that all rate limiting algorithms must implement.
type Limiter interface {
	Allow(ctx context.Context, cfg LimitConfig) (*CheckResponse, error)
}

// LimitConfig holds the parameters for a rate limit check, shared across all algorithms.
type LimitConfig struct {
	TenantID string
	Profile  string
	UserKey  string
	Capacity int
	Limit    int
	Window   time.Duration
}

// Registry maps algorithm names to their Limiter implementations.
type Registry struct {
	limiters map[common.Algorithm]Limiter
}

// NewRegistry creates a Registry and registers the provided algorithm-limiter pairs.
func NewRegistry(entries map[common.Algorithm]Limiter) *Registry {
	return &Registry{limiters: entries}
}

// Get returns the Limiter for the given algorithm, or an error if unsupported.
func (r *Registry) Get(algo common.Algorithm) (Limiter, error) {
	l, ok := r.limiters[algo]
	if !ok {
		return nil, fmt.Errorf("unsupported algorithm: %s", algo)
	}
	return l, nil
}
