package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const RequestIDKey = "request_id"

// RequestID injects a unique request id into context and response header.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := uuid.NewString()
		c.Set(RequestIDKey, id)
		c.Writer.Header().Set("X-Request-ID", id)
		c.Next()
	}
}
