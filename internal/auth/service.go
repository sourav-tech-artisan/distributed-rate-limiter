package auth

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/souravkumar/distributed-rate-limiter/internal/common"
	"github.com/souravkumar/distributed-rate-limiter/internal/platform/config"
)

// Service handles authentication business logic
type Service struct {
	repo   *Repository
	config *config.Config
	logger zerolog.Logger
}

// NewService creates a new auth service
func NewService(repo *Repository, cfg *config.Config, logger zerolog.Logger) *Service {
	return &Service{
		repo:   repo,
		config: cfg,
		logger: logger,
	}
}

// Register creates a new tenant account
func (s *Service) Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error) {
	// Hash password
	passwordHash, err := HashPassword(req.Password)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to hash password")
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Generate API key
	apiKey, err := GenerateAPIKey()
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to generate api key")
		return nil, fmt.Errorf("failed to generate api key: %w", err)
	}

	// Create tenant
	tenant := &Tenant{
		Email:             req.Email,
		PasswordHash:      passwordHash,
		APIKey:            apiKey,
		MaxProfiles:       s.config.Quotas.DefaultMaxProfiles,
		MaxRequestsPerDay: s.config.Quotas.DefaultMaxRequestsPerDay,
		IsActive:          true,
	}

	if err := s.repo.Create(ctx, tenant); err != nil {
		if err == common.ErrDuplicateEntry {
			return nil, fmt.Errorf("email already registered")
		}
		s.logger.Error().Err(err).Str("email", req.Email).Msg("failed to create tenant")
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	s.logger.Info().Str("email", req.Email).Str("tenant_id", tenant.ID).Msg("tenant registered")

	return &RegisterResponse{
		APIKey:  apiKey,
		Message: "Account created successfully",
	}, nil
}

// Login authenticates a tenant and returns their API key
func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	// Get tenant by email
	tenant, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		if err == common.ErrNotFound {
			return nil, common.ErrUnauthorized
		}
		s.logger.Error().Err(err).Str("email", req.Email).Msg("failed to get tenant")
		return nil, fmt.Errorf("login failed: %w", err)
	}

	// Check if tenant is active
	if !tenant.IsActive {
		s.logger.Warn().Str("email", req.Email).Msg("inactive tenant attempted login")
		return nil, common.ErrForbidden
	}

	// Verify password
	if !CheckPassword(req.Password, tenant.PasswordHash) {
		s.logger.Warn().Str("email", req.Email).Msg("invalid password attempt")
		return nil, common.ErrUnauthorized
	}

	s.logger.Info().Str("email", req.Email).Str("tenant_id", tenant.ID).Msg("tenant logged in")

	return &LoginResponse{
		APIKey: tenant.APIKey,
	}, nil
}
