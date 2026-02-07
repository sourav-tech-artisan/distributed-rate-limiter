package profile

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/souravkumar/distributed-rate-limiter/internal/common"
)

const profileCacheTTL = 5 * time.Minute

// cachedRepository is a caching decorator around a Repository.
// It embeds the Repository interface so non-cached methods (Create, FindAllByTenant,
// CountByTenant) are automatically delegated to the inner repository.
// Update and Delete are overridden to add cache invalidation.
type cachedRepository struct {
	Repository // embedded interface — delegates all methods by default
	redis      *redis.Client
	logger     zerolog.Logger
}

// NewCachedRepository wraps an existing Repository with Redis caching.
// Caches FindByName results and invalidates on Update/Delete.
func NewCachedRepository(inner Repository, rdb *redis.Client, logger zerolog.Logger) Repository {
	return &cachedRepository{
		Repository: inner,
		redis:      rdb,
		logger:     logger,
	}
}

// FindByName checks the Redis cache first, falls back to the inner
// repository on miss, and populates the cache for subsequent calls.
func (cr *cachedRepository) FindByName(ctx context.Context, tenantID, name string) (*Profile, error) {
	cacheKey := profileCacheKey(tenantID, name)

	// Try cache first
	data, err := cr.redis.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var p Profile
		if err := json.Unmarshal(data, &p); err == nil {
			cr.logger.Debug().Str("cache_key", cacheKey).Msg("profile cache hit")
			return &p, nil
		}
		cr.logger.Warn().Err(err).Msg("failed to unmarshal cached profile")
	}

	// Cache miss — query database via inner repository
	p, err := cr.Repository.FindByName(ctx, tenantID, name)
	if err != nil {
		return nil, err
	}

	// Populate cache (best-effort, don't fail the request)
	if encoded, err := json.Marshal(p); err == nil {
		if err := cr.redis.Set(ctx, cacheKey, encoded, profileCacheTTL).Err(); err != nil {
			cr.logger.Warn().Err(err).Str("cache_key", cacheKey).Msg("failed to cache profile")
		}
	}

	return p, nil
}

// Update modifies a profile in the database and invalidates its cache entry.
func (cr *cachedRepository) Update(ctx context.Context, p *Profile) error {
	if err := cr.Repository.Update(ctx, p); err != nil {
		return err
	}
	cr.invalidate(ctx, p.TenantID, p.Name)
	return nil
}

// Delete removes a profile from the database and invalidates its cache entry.
func (cr *cachedRepository) Delete(ctx context.Context, tenantID, name string) error {
	if err := cr.Repository.Delete(ctx, tenantID, name); err != nil {
		return err
	}
	cr.invalidate(ctx, tenantID, name)
	return nil
}

// invalidate removes a profile from the Redis cache.
func (cr *cachedRepository) invalidate(ctx context.Context, tenantID, name string) {
	cacheKey := profileCacheKey(tenantID, name)
	if err := cr.redis.Del(ctx, cacheKey).Err(); err != nil {
		cr.logger.Warn().Err(err).Str("cache_key", cacheKey).Msg("failed to invalidate profile cache")
	}
}

// profileCacheKey builds the Redis key for a cached profile.
func profileCacheKey(tenantID, name string) string {
	return fmt.Sprintf("%s%s:%s", common.CacheProfileKeyPrefix, tenantID, name)
}
