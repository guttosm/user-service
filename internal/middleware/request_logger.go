package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger logs method, path, status, and latency for each request.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		method := c.Request.Method
		path := c.Request.URL.Path
		c.Next()
		latency := time.Since(start)
		status := c.Writer.Status()
		rid, _ := c.Get(RequestIDKey)
		log.Printf("request_id=%v method=%s path=%s status=%d latency_ms=%d", rid, method, path, status, latency.Milliseconds())
	}
}

type client struct {
	lastSeen time.Time
	count    int
}

var (
	clients = make(map[string]*client)
	window  = time.Minute
	limit   = 60
)

// RateLimiter limits the number of requests a client can make.
func RateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()
		cl, ok := clients[ip]
		if !ok || now.Sub(cl.lastSeen) > window {
			cl = &client{lastSeen: now, count: 1}
			clients[ip] = cl
		} else {
			cl.count++
			cl.lastSeen = now
		}
		if cl.count > limit {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}
