package api

import (
	"github.com/gin-gonic/gin"
)

// SetupRouter configures and returns the Gin router
func SetupRouter(handler *Handler) *gin.Engine {
	router := gin.New()

	// Middleware
	router.Use(gin.Recovery())
	router.Use(loggingMiddleware())

	// Health check endpoint
	router.GET("/health", handler.HealthCheck)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Rate limit check endpoint (to be implemented)
		v1.POST("/rate-limit/check", handler.RateLimitCheck)
	}

	return router
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Request start time
		// TODO: Add proper request logging in future commit
		c.Next()
	}
}
