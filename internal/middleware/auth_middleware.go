package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/parxyws/cozybox/internal/tools/contextkey"
	"github.com/parxyws/cozybox/internal/tools/util"
)

// AuthMiddleware validates the JWT access token from the Authorization header
// and injects user_id, tenant_id, and session_id into the request context.
func (m *ManagerMiddleware) AuthMiddleware(jwt util.TokenMaker) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			return
		}

		claims, err := jwt.VerifyAccessToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		// Inject claims into Gin context (for handlers)
		c.Set(string(contextkey.UserID), claims.UserID)
		c.Set(string(contextkey.TenantID), claims.TenantID)
		c.Set("session_id", claims.SessionID)

		// Inject into Go context (for service/repo layer)
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, contextkey.UserID, claims.UserID)
		ctx = context.WithValue(ctx, contextkey.TenantID, claims.TenantID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
