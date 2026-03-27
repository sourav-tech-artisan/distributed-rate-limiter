package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/souravkumar/distributed-rate-limiter/internal/common"
	"github.com/souravkumar/distributed-rate-limiter/internal/profile"
)

// Service handles rate limiting business logic
type Service struct {
	registry    *Registry
	quota       *QuotaTracker
	profileRepo profile.Repository
	logger      zerolog.Logger
}

// NewService creates a new rate limit service
func NewService(registry *Registry, quota *QuotaTracker, profileRepo profile.Repository, logger zerolog.Logger) *Service {
	return &Service{
		registry:    registry,
		quota:       quota,
		profileRepo: profileRepo,
		logger:      logger,
	}
}

// Check performs a rate limit check for a given key and profile.
// It first checks the tenant's daily quota, then dispatches to the appropriate
// algorithm based on the profile configuration.
// On Redis/circuit breaker failures, it fails open (allows the request) to avoid
// blocking all traffic when the infrastructure is degraded.
func (s *Service) Check(ctx context.Context, tenantID string, maxRequestsPerDay int, req CheckRequest) (*CheckResponse, error) {
	// Check daily quota first
	if err := s.quota.Check(ctx, tenantID, maxRequestsPerDay); err != nil {
		if errors.Is(err, common.ErrQuotaExceeded) {
			return nil, err
		}
		s.logger.Warn().Err(err).
			Str("tenant_id", tenantID).
			Msg("quota check failed, proceeding with rate limit check")
	}

	// Look up the profile
	p, err := s.profileRepo.FindByName(ctx, tenantID, req.Profile)
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
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

	// Resolve the limiter for this profile's algorithm
	limiter, err := s.registry.Get(common.Algorithm(p.Algorithm))
	if err != nil {
		return nil, fmt.Errorf("%w: %s", common.ErrInvalidInput, err)
	}

	// Run the rate limit algorithm
	resp, err := limiter.Allow(ctx, LimitConfig{
		TenantID: tenantID,
		Profile:  req.Profile,
		UserKey:  req.Key,
		Capacity: p.GetCapacity(),
		Limit:    p.Limit,
		Window:   window,
	})
	if err != nil {
		// Rate limit check failed (Redis/circuit breaker issue) — fail open
		s.logger.Warn().Err(err).
			Str("tenant_id", tenantID).
			Str("profile", req.Profile).
			Str("key", req.Key).
			Msg("rate limit check failed, allowing request (fail-open)")
		return &CheckResponse{
			Allowed:   true,
			Limit:     p.Limit,
			Remaining: p.Limit,
			Reset:     time.Now().Add(window).Unix(),
		}, nil
	}

	return resp, nil
}
