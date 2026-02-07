package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/souravkumar/distributed-rate-limiter/internal/common"
	"github.com/souravkumar/distributed-rate-limiter/internal/profile"
)

// Service handles rate limiting business logic
type Service struct {
	limiter     *TokenBucket
	profileRepo profile.Repository
	logger      zerolog.Logger
}

// NewService creates a new rate limit service
func NewService(limiter *TokenBucket, profileRepo profile.Repository, logger zerolog.Logger) *Service {
	return &Service{
		limiter:     limiter,
		profileRepo: profileRepo,
		logger:      logger,
	}
}

// Check performs a rate limit check for a given key and profile
func (s *Service) Check(ctx context.Context, tenantID string, req CheckRequest) (*CheckResponse, error) {
	// Look up the profile
	p, err := s.profileRepo.FindByName(ctx, tenantID, req.Profile)
	if err != nil {
		if err == common.ErrNotFound {
			return nil, fmt.Errorf("%w: profile '%s'", common.ErrNotFound, req.Profile)
		}
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID).
			Str("profile", req.Profile).
			Msg("failed to find profile")
		return nil, fmt.Errorf("failed to find profile: %w", err)
	}

	// Parse window duration (e.g. "1m", "30s", "1h")
	window, err := time.ParseDuration(p.Window)
	if err != nil {
		s.logger.Error().Err(err).
			Str("window", p.Window).
			Msg("invalid window duration in profile")
		return nil, fmt.Errorf("invalid window duration '%s': %w", p.Window, err)
	}

	// Run the token bucket algorithm
	resp, err := s.limiter.Allow(ctx, BucketConfig{
		TenantID: tenantID,
		Profile:  req.Profile,
		UserKey:  req.Key,
		Capacity: p.GetCapacity(),
		Limit:    p.Limit,
		Window:   window,
	})
	if err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID).
			Str("profile", req.Profile).
			Str("key", req.Key).
			Msg("rate limit check failed")
		return nil, fmt.Errorf("rate limit check failed: %w", err)
	}

	return resp, nil
}
