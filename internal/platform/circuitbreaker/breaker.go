package circuitbreaker

import (
	"github.com/rs/zerolog"
	"github.com/sony/gobreaker"
	"github.com/souravkumar/distributed-rate-limiter/internal/platform/config"
)

// New creates a circuit breaker with the given name and configuration.
// The breaker opens after consecutive failures exceed the threshold,
// and transitions to half-open after the configured timeout.
func New(name string, cfg *config.CircuitBreakerConfig, logger zerolog.Logger) *gobreaker.CircuitBreaker {
	settings := gobreaker.Settings{
		Name:        name,
		MaxRequests: cfg.HalfOpenMaxRequests,
		Timeout:     cfg.GetTimeout(),
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= uint32(cfg.FailureThreshold)
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			logger.Warn().
				Str("breaker", name).
				Str("from", from.String()).
				Str("to", to.String()).
				Msg("circuit breaker state changed")
		},
	}

	return gobreaker.NewCircuitBreaker(settings)
}
