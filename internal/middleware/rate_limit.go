package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimitMiddleware provides a simple in-memory rate limiter per IP.
func RateLimitMiddleware(limit int, window time.Duration) gin.HandlerFunc {
	type client struct {
		lastSeen time.Time
		count    int
	}
	var (
		clients = make(map[string]*client)
		mu      sync.Mutex
	)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()
		mu.Lock()
		cl, ok := clients[ip]
		if !ok || now.Sub(cl.lastSeen) > window {
			cl = &client{lastSeen: now, count: 1}
			clients[ip] = cl
		} else {
			cl.count++
			cl.lastSeen = now
		}
		count := cl.count
		mu.Unlock()
		if count > limit {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}
