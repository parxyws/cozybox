package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/parxyws/cozybox/internal/tools/contextkey"
)

// TenantScopeMiddleware verifies that the request carries a valid tenant context
// (set by AuthMiddleware) and blocks requests without one.
// This is the critical security boundary — it ensures every downstream query
// operates within a tenant scope.
func (m *ManagerMiddleware) TenantScopeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.Request.Context().Value(contextkey.TenantID)
		if tenantID == nil || tenantID.(string) == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "tenant context required"})
			return
		}

		c.Next()
	}
}
