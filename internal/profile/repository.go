package profile

import (
	"context"
	"errors"
	"strings"

	"github.com/souravkumar/distributed-rate-limiter/internal/common"
	"gorm.io/gorm"
)

// Repository handles database operations for profiles
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new profile Repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create inserts a new profile into the database
func (r *Repository) Create(ctx context.Context, profile *Profile) error {
	result := r.db.WithContext(ctx).Create(profile)
	if result.Error != nil {
		if isDuplicateKeyError(result.Error) {
			return common.ErrDuplicateEntry
		}
		return result.Error
	}
	return nil
}

// FindByName retrieves a profile by tenant ID and name
func (r *Repository) FindByName(ctx context.Context, tenantID, name string) (*Profile, error) {
	var profile Profile
	result := r.db.WithContext(ctx).
		Where("tenant_id = ? AND name = ?", tenantID, name).
		First(&profile)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, result.Error
	}
	return &profile, nil
}

// FindAllByTenant retrieves all profiles for a tenant
func (r *Repository) FindAllByTenant(ctx context.Context, tenantID string) ([]Profile, error) {
	var profiles []Profile
	result := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Find(&profiles)
	if result.Error != nil {
		return nil, result.Error
	}
	return profiles, nil
}

// Update modifies an existing profile
func (r *Repository) Update(ctx context.Context, profile *Profile) error {
	result := r.db.WithContext(ctx).Save(profile)
	return result.Error
}

// Delete removes a profile by tenant ID and name
func (r *Repository) Delete(ctx context.Context, tenantID, name string) error {
	result := r.db.WithContext(ctx).
		Where("tenant_id = ? AND name = ?", tenantID, name).
		Delete(&Profile{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

// CountByTenant returns the number of profiles for a tenant
func (r *Repository) CountByTenant(ctx context.Context, tenantID string) (int64, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&Profile{}).
		Where("tenant_id = ?", tenantID).
		Count(&count)
	return count, result.Error
}

// isDuplicateKeyError checks if the error is a unique constraint violation
func isDuplicateKeyError(err error) bool {
	return strings.Contains(err.Error(), "duplicate key")
}
