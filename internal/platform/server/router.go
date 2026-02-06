package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/souravkumar/distributed-rate-limiter/internal/auth"
	"github.com/souravkumar/distributed-rate-limiter/internal/profile"
	"github.com/souravkumar/distributed-rate-limiter/internal/ratelimit"
)

// RouterConfig holds all the handlers and middleware needed for routing
type RouterConfig struct {
	AuthHandler      *auth.Handler
	AuthMiddleware   gin.HandlerFunc
	ProfileHandler   *profile.Handler
	RateLimitHandler *ratelimit.Handler
}

// NewRouter creates and configures the Gin router
func NewRouter(cfg *RouterConfig) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Global middleware
	router.Use(gin.Recovery())
	router.Use(requestLogger())

	// Health check (no auth required)
	router.GET("/health", healthCheck)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Auth routes (no API key required)
		authRoutes := v1.Group("/auth")
		{
			authRoutes.POST("/register", cfg.AuthHandler.Register)
			authRoutes.POST("/login", cfg.AuthHandler.Login)
		}

		// Protected routes (API key required)
		protected := v1.Group("")
		protected.Use(cfg.AuthMiddleware)
		{
			// Profile routes
			profiles := protected.Group("/profiles")
			{
				profiles.POST("", cfg.ProfileHandler.Create)
				profiles.GET("", cfg.ProfileHandler.List)
				profiles.GET("/:name", cfg.ProfileHandler.Get)
				profiles.PUT("/:name", cfg.ProfileHandler.Update)
				profiles.DELETE("/:name", cfg.ProfileHandler.Delete)
			}

			// Rate limit routes
			rl := protected.Group("/rate-limit")
			{
				rl.POST("/check", cfg.RateLimitHandler.Check)
			}
		}
	}

	return router
}

// healthCheck handles the health check endpoint
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}

// requestLogger is a middleware that logs all requests
func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
