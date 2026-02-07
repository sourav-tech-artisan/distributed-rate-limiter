package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/souravkumar/distributed-rate-limiter/internal/common"
	"gorm.io/gorm"
)

// Repository defines data access operations for tenants
type Repository interface {
	Create(ctx context.Context, tenant *Tenant) error
	GetByID(ctx context.Context, id string) (*Tenant, error)
	GetByEmail(ctx context.Context, email string) (*Tenant, error)
	GetByAPIKey(ctx context.Context, apiKey string) (*Tenant, error)
	GetActiveByAPIKey(ctx context.Context, apiKey string) (*Tenant, error)
}

type repository struct {
	db *gorm.DB
}

// NewRepository creates a new PostgreSQL-backed tenant repository
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// Create inserts a new tenant into the database
func (r *repository) Create(ctx context.Context, tenant *Tenant) error {
	result := r.db.WithContext(ctx).Create(tenant)
	if result.Error != nil {
		// Check for unique constraint violations
		if isDuplicateKeyError(result.Error, "email") {
			return common.ErrDuplicateEntry
		}
		if isDuplicateKeyError(result.Error, "api_key") {
			return common.ErrDuplicateEntry
		}
		return result.Error
	}
	return nil
}

// GetByID retrieves a tenant by their ID
func (r *repository) GetByID(ctx context.Context, id string) (*Tenant, error) {
	var tenant Tenant
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&tenant)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, result.Error
	}
	return &tenant, nil
}

// GetByEmail retrieves a tenant by their email address
func (r *repository) GetByEmail(ctx context.Context, email string) (*Tenant, error) {
	var tenant Tenant
	result := r.db.WithContext(ctx).Where("email = ?", email).First(&tenant)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, result.Error
	}
	return &tenant, nil
}

// GetByAPIKey retrieves a tenant by their API key
func (r *repository) GetByAPIKey(ctx context.Context, apiKey string) (*Tenant, error) {
	var tenant Tenant
	result := r.db.WithContext(ctx).Where("api_key = ?", apiKey).First(&tenant)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, result.Error
	}
	return &tenant, nil
}

// GetActiveByAPIKey retrieves an active tenant by their API key
// Returns ErrInactive if the tenant exists but is not active
func (r *repository) GetActiveByAPIKey(ctx context.Context, apiKey string) (*Tenant, error) {
	tenant, err := r.GetByAPIKey(ctx, apiKey)
	if err != nil {
		return nil, err
	}
	if !tenant.IsActive {
		return nil, common.ErrInactive
	}
	return tenant, nil
}

// isDuplicateKeyError checks if the error is a duplicate key violation
func isDuplicateKeyError(err error, field string) bool {
	// PostgreSQL unique violation error code is 23505
	errStr := err.Error()
	return strings.Contains(errStr, "duplicate key") && strings.Contains(errStr, field)
}
