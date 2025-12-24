package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
	// Rate limiter will be added in future commits
}

// NewHandler creates a new handler instance
func NewHandler() *Handler {
	return &Handler{}
}

// HealthCheck handles health check requests
func (h *Handler) HealthCheck(c *gin.Context) {
	log.Debug().Msg("health check requested")

	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "rate-limiter",
	})
}

// RateLimitCheck will handle rate limit check requests (to be implemented)
func (h *Handler) RateLimitCheck(c *gin.Context) {
	// TODO: Implement in future commit
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "not implemented yet",
	})
}
