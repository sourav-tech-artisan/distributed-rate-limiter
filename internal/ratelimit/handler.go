package ratelimit

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/souravkumar/distributed-rate-limiter/internal/auth"
	"github.com/souravkumar/distributed-rate-limiter/internal/common"
)

// Handler handles HTTP requests for rate limiting
type Handler struct {
	service *Service
	logger  zerolog.Logger
}

// NewHandler creates a new rate limit handler
func NewHandler(service *Service, logger zerolog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// Check handles POST /api/v1/rate-limit/check
func (h *Handler) Check(c *gin.Context) {
	tenant := getTenant(c)
	if tenant == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req CheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	resp, err := h.service.Check(c.Request.Context(), tenant.ID, tenant.MaxRequestsPerDay, req)
	if err != nil {
		h.logger.Error().Err(err).Msg("rate limit check failed")

		if errors.Is(err, common.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		if errors.Is(err, common.ErrQuotaExceeded) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Daily request quota exceeded",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Rate limit check failed"})
		return
	}

	// Return 429 if rate limited, 200 if allowed
	if !resp.Allowed {
		c.JSON(http.StatusTooManyRequests, resp)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// getTenant retrieves the authenticated tenant from Gin context
func getTenant(c *gin.Context) *auth.Tenant {
	if t, exists := c.Get("tenant"); exists {
		if tenant, ok := t.(*auth.Tenant); ok {
			return tenant
		}
	}
	return nil
}
