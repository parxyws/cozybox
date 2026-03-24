package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	xRequestIDKey = "X-Request-ID"
)

func (m *ManagerMiddleware) RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.Request.Header.Get(xRequestIDKey)
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set(xRequestIDKey, requestID)
		//c.Header(xRequestIDKey, requestID)
		c.Next()
	}
}
