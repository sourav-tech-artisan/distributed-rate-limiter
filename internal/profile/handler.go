package profile

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/souravkumar/distributed-rate-limiter/internal/auth"
	"github.com/souravkumar/distributed-rate-limiter/internal/common"
)

// Handler handles HTTP requests for profile management
type Handler struct {
	service *Service
	logger  zerolog.Logger
}

// NewHandler creates a new profile handler
func NewHandler(service *Service, logger zerolog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// Create handles POST /api/v1/profiles
func (h *Handler) Create(c *gin.Context) {
	tenant := getTenant(c)
	if tenant == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	resp, err := h.service.Create(c.Request.Context(), tenant.ID, tenant.MaxProfiles, req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// List handles GET /api/v1/profiles
func (h *Handler) List(c *gin.Context) {
	tenant := getTenant(c)
	if tenant == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	resp, err := h.service.List(c.Request.Context(), tenant.ID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Get handles GET /api/v1/profiles/:name
func (h *Handler) Get(c *gin.Context) {
	tenant := getTenant(c)
	if tenant == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	name := c.Param("name")
	resp, err := h.service.Get(c.Request.Context(), tenant.ID, name)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Update handles PUT /api/v1/profiles/:name
func (h *Handler) Update(c *gin.Context) {
	tenant := getTenant(c)
	if tenant == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	name := c.Param("name")
	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	resp, err := h.service.Update(c.Request.Context(), tenant.ID, name, req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Delete handles DELETE /api/v1/profiles/:name
func (h *Handler) Delete(c *gin.Context) {
	tenant := getTenant(c)
	if tenant == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	name := c.Param("name")
	if err := h.service.Delete(c.Request.Context(), tenant.ID, name); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile deleted"})
}

// handleError maps domain errors to HTTP responses
func (h *Handler) handleError(c *gin.Context, err error) {
	h.logger.Error().Err(err).Msg("profile operation failed")

	if errors.Is(err, common.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if errors.Is(err, common.ErrDuplicateEntry) {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	if errors.Is(err, common.ErrInvalidInput) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if errors.Is(err, common.ErrQuotaExceeded) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
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
