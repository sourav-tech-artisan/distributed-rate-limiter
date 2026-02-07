package server

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/souravkumar/distributed-rate-limiter/internal/auth"
	"github.com/souravkumar/distributed-rate-limiter/internal/common"
)

// Context keys for storing tenant information
type contextKey string

const (
	// TenantContextKey is the key for storing tenant in context
	TenantContextKey contextKey = "tenant"

	// APIKeyHeader is the header name for API key authentication
	APIKeyHeader = "X-API-Key"
)

// AuthMiddleware creates a middleware that validates API keys
func AuthMiddleware(authRepo auth.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader(APIKeyHeader)
		if apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "API key required",
			})
			return
		}

		// Get tenant by API key
		tenant, err := authRepo.GetActiveByAPIKey(c.Request.Context(), apiKey)
		if err != nil {
			if err == common.ErrNotFound {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "Invalid API key",
				})
				return
			}

			if err == common.ErrInactive {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error": "Account is inactive",
				})
				return
			}

			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Authentication failed",
			})
			return
		}

		// Store tenant in context for downstream handlers
		ctx := context.WithValue(c.Request.Context(), TenantContextKey, tenant)
		c.Request = c.Request.WithContext(ctx)

		// Also store in Gin context for easy access
		c.Set("tenant", tenant)
		c.Set("tenant_id", tenant.ID)

		c.Next()
	}
}

// GetTenantFromContext retrieves the tenant from the request context
func GetTenantFromContext(ctx context.Context) *auth.Tenant {
	if tenant, ok := ctx.Value(TenantContextKey).(*auth.Tenant); ok {
		return tenant
	}
	return nil
}

// GetTenantFromGin retrieves the tenant from Gin context
func GetTenantFromGin(c *gin.Context) *auth.Tenant {
	if tenant, exists := c.Get("tenant"); exists {
		if t, ok := tenant.(*auth.Tenant); ok {
			return t
		}
	}
	return nil
}

// GetTenantIDFromGin retrieves the tenant ID from Gin context
func GetTenantIDFromGin(c *gin.Context) string {
	if tenantID, exists := c.Get("tenant_id"); exists {
		if id, ok := tenantID.(string); ok {
			return id
		}
	}
	return ""
}
