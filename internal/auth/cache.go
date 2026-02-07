package auth

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/souravkumar/distributed-rate-limiter/internal/common"
)

const tenantCacheTTL = 5 * time.Minute

// cachedRepository is a caching decorator around a Repository.
// It embeds the Repository interface so non-cached methods (Create, GetByEmail, etc.)
// are automatically delegated to the inner repository.
type cachedRepository struct {
	Repository // embedded interface — delegates all methods by default
	redis      *redis.Client
	logger     zerolog.Logger
}

// NewCachedRepository wraps an existing Repository with Redis caching.
// The returned Repository transparently caches GetActiveByAPIKey results.
func NewCachedRepository(inner Repository, rdb *redis.Client, logger zerolog.Logger) Repository {
	return &cachedRepository{
		Repository: inner,
		redis:      rdb,
		logger:     logger,
	}
}

// GetActiveByAPIKey checks the Redis cache first, falls back to the inner
// repository on miss, and populates the cache for subsequent calls.
func (cr *cachedRepository) GetActiveByAPIKey(ctx context.Context, apiKey string) (*Tenant, error) {
	cacheKey := common.CacheTenantKeyPrefix + apiKey

	// Try cache first
	data, err := cr.redis.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var tenant Tenant
		if err := json.Unmarshal(data, &tenant); err == nil {
			cr.logger.Debug().Str("cache_key", cacheKey).Msg("tenant cache hit")
			return &tenant, nil
		}
		cr.logger.Warn().Err(err).Msg("failed to unmarshal cached tenant")
	}

	// Cache miss — query database via inner repository
	tenant, err := cr.Repository.GetActiveByAPIKey(ctx, apiKey)
	if err != nil {
		return nil, err
	}

	// Populate cache (best-effort, don't fail the request)
	if encoded, err := json.Marshal(tenant); err == nil {
		if err := cr.redis.Set(ctx, cacheKey, encoded, tenantCacheTTL).Err(); err != nil {
			cr.logger.Warn().Err(err).Str("cache_key", cacheKey).Msg("failed to cache tenant")
		}
	}

	return tenant, nil
}

// InvalidateByAPIKey removes a tenant from the cache.
func (cr *cachedRepository) InvalidateByAPIKey(ctx context.Context, apiKey string) {
	cacheKey := common.CacheTenantKeyPrefix + apiKey
	if err := cr.redis.Del(ctx, cacheKey).Err(); err != nil {
		cr.logger.Warn().Err(err).Str("cache_key", cacheKey).Msg("failed to invalidate tenant cache")
	}
}
