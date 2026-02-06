package profile

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/souravkumar/distributed-rate-limiter/internal/common"
)

// Service handles profile business logic
type Service struct {
	repo   *Repository
	logger zerolog.Logger
}

// NewService creates a new profile service
func NewService(repo *Repository, logger zerolog.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// Create creates a new rate limiting profile for a tenant
func (s *Service) Create(ctx context.Context, tenantID string, maxProfiles int, req CreateRequest) (*Response, error) {
	// Validate algorithm
	algo := common.Algorithm(req.Algorithm)
	if !algo.IsValid() {
		return nil, fmt.Errorf("%w: unsupported algorithm '%s'", common.ErrInvalidInput, req.Algorithm)
	}

	// Check profile quota
	count, err := s.repo.CountByTenant(ctx, tenantID)
	if err != nil {
		s.logger.Error().Err(err).Str("tenant_id", tenantID).Msg("failed to count profiles")
		return nil, fmt.Errorf("failed to check quota: %w", err)
	}

	if int(count) >= maxProfiles {
		return nil, fmt.Errorf("%w: maximum %d profiles allowed", common.ErrQuotaExceeded, maxProfiles)
	}

	// Create profile
	profile := &Profile{
		TenantID:  tenantID,
		Name:      req.Name,
		Algorithm: req.Algorithm,
		Limit:     req.Limit,
		Window:    req.Window,
		Capacity:  req.Capacity,
	}

	if err := s.repo.Create(ctx, profile); err != nil {
		if err == common.ErrDuplicateEntry {
			return nil, fmt.Errorf("%w: profile '%s' already exists", common.ErrDuplicateEntry, req.Name)
		}
		s.logger.Error().Err(err).Str("tenant_id", tenantID).Str("name", req.Name).Msg("failed to create profile")
		return nil, fmt.Errorf("failed to create profile: %w", err)
	}

	s.logger.Info().Str("tenant_id", tenantID).Str("name", req.Name).Msg("profile created")

	resp := ToResponse(profile)
	return &resp, nil
}

// List returns all profiles for a tenant
func (s *Service) List(ctx context.Context, tenantID string) (*ListResponse, error) {
	profiles, err := s.repo.FindAllByTenant(ctx, tenantID)
	if err != nil {
		s.logger.Error().Err(err).Str("tenant_id", tenantID).Msg("failed to list profiles")
		return nil, fmt.Errorf("failed to list profiles: %w", err)
	}

	resp := ToListResponse(profiles)
	return &resp, nil
}

// Get returns a single profile by name
func (s *Service) Get(ctx context.Context, tenantID, name string) (*Response, error) {
	profile, err := s.repo.FindByName(ctx, tenantID, name)
	if err != nil {
		if err == common.ErrNotFound {
			return nil, fmt.Errorf("%w: profile '%s'", common.ErrNotFound, name)
		}
		s.logger.Error().Err(err).Str("tenant_id", tenantID).Str("name", name).Msg("failed to get profile")
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	resp := ToResponse(profile)
	return &resp, nil
}

// Update modifies an existing profile
func (s *Service) Update(ctx context.Context, tenantID, name string, req UpdateRequest) (*Response, error) {
	// Find existing profile
	profile, err := s.repo.FindByName(ctx, tenantID, name)
	if err != nil {
		if err == common.ErrNotFound {
			return nil, fmt.Errorf("%w: profile '%s'", common.ErrNotFound, name)
		}
		return nil, fmt.Errorf("failed to find profile: %w", err)
	}

	// Apply updates
	if req.Limit != nil {
		profile.Limit = *req.Limit
	}
	if req.Window != nil {
		profile.Window = *req.Window
	}
	if req.Capacity != nil {
		profile.Capacity = req.Capacity
	}

	if err := s.repo.Update(ctx, profile); err != nil {
		s.logger.Error().Err(err).Str("tenant_id", tenantID).Str("name", name).Msg("failed to update profile")
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	s.logger.Info().Str("tenant_id", tenantID).Str("name", name).Msg("profile updated")

	resp := ToResponse(profile)
	return &resp, nil
}

// Delete removes a profile
func (s *Service) Delete(ctx context.Context, tenantID, name string) error {
	if err := s.repo.Delete(ctx, tenantID, name); err != nil {
		if err == common.ErrNotFound {
			return fmt.Errorf("%w: profile '%s'", common.ErrNotFound, name)
		}
		s.logger.Error().Err(err).Str("tenant_id", tenantID).Str("name", name).Msg("failed to delete profile")
		return fmt.Errorf("failed to delete profile: %w", err)
	}

	s.logger.Info().Str("tenant_id", tenantID).Str("name", name).Msg("profile deleted")
	return nil
}
