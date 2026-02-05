package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/souravkumar/distributed-rate-limiter/internal/common"
)

// Handler handles HTTP requests for authentication
type Handler struct {
	service *Service
	logger  zerolog.Logger
}

// NewHandler creates a new auth handler
func NewHandler(service *Service, logger zerolog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// Register handles POST /api/v1/auth/register
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	resp, err := h.service.Register(c.Request.Context(), req)
	if err != nil {
		h.logger.Error().Err(err).Msg("registration failed")

		// Check for duplicate email
		if err.Error() == "email already registered" {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Email already registered",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Registration failed",
		})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// Login handles POST /api/v1/auth/login
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	resp, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		h.logger.Error().Err(err).Msg("login failed")

		if err == common.ErrUnauthorized {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid email or password",
			})
			return
		}

		if err == common.ErrForbidden {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Account is inactive",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Login failed",
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}
