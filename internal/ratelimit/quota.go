package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/sony/gobreaker"
	"github.com/souravkumar/distributed-rate-limiter/internal/common"
)

const quotaTTL = 48 * time.Hour

// QuotaTracker tracks daily API usage per tenant using Redis counters.
// Each tenant has a daily counter key: usage:{tenant_id}:{YYYY-MM-DD}
// with a 48-hour TTL to allow cross-midnight queries.
type QuotaTracker struct {
	client *redis.Client
	cb     *gobreaker.CircuitBreaker
	logger zerolog.Logger
}

// NewQuotaTracker creates a new quota tracker
func NewQuotaTracker(client *redis.Client, cb *gobreaker.CircuitBreaker, logger zerolog.Logger) *QuotaTracker {
	return &QuotaTracker{
		client: client,
		cb:     cb,
		logger: logger,
	}
}

// Check atomically increments the daily usage counter and returns an error
// if the tenant has exceeded their MaxRequestsPerDay quota.
// Returns common.ErrQuotaExceeded if the quota is exceeded.
func (qt *QuotaTracker) Check(ctx context.Context, tenantID string, maxRequests int) error {
	today := time.Now().UTC().Format("2006-01-02")
	key := fmt.Sprintf("%s%s:%s", common.UsageKeyPrefix, tenantID, today)

	result, err := qt.cb.Execute(func() (interface{}, error) {
		count, err := qt.client.Incr(ctx, key).Result()
		if err != nil {
			return nil, err
		}

		// Set TTL on first increment (when key is newly created)
		if count == 1 {
			qt.client.Expire(ctx, key, quotaTTL)
		}

		return count, nil
	})
	if err != nil {
		return fmt.Errorf("failed to check usage quota: %w", err)
	}

	count := result.(int64)
	if int(count) > maxRequests {
		qt.logger.Warn().
			Str("tenant_id", tenantID).
			Int64("usage", count).
			Int("limit", maxRequests).
			Msg("daily quota exceeded")
		return common.ErrQuotaExceeded
	}

	return nil
}
