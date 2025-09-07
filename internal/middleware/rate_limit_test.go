package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRateLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RateLimitMiddleware(2, time.Minute)) // allow 2 requests
	r.GET("/ping", func(c *gin.Context) { c.String(200, "pong") })

	mkReq := func() *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		req.RemoteAddr = "10.1.2.3:12345" // stable IP
		r.ServeHTTP(w, req)
		return w
	}

	w1 := mkReq()
	assert.Equal(t, 200, w1.Code)
	w2 := mkReq()
	assert.Equal(t, 200, w2.Code)
	w3 := mkReq()
	assert.Equal(t, http.StatusTooManyRequests, w3.Code)
	assert.Contains(t, w3.Body.String(), "rate limit exceeded")
}

func TestRateLimit_WindowReset(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RateLimitMiddleware(1, 200*time.Millisecond)) // 1 per window
	r.GET("/ping", func(c *gin.Context) { c.String(200, "pong") })

	mk := func() *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		req.RemoteAddr = "10.9.9.9:1234"
		r.ServeHTTP(w, req)
		return w
	}

	w1 := mk()
	assert.Equal(t, 200, w1.Code)
	w2 := mk()
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
	time.Sleep(250 * time.Millisecond) // beyond window
	w3 := mk()
	assert.Equal(t, 200, w3.Code)
}
