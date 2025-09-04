package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID())
	r.GET("/rid", func(c *gin.Context) {
		id := c.GetString(RequestIDKey)
		if id == "" {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.String(http.StatusOK, id)
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/rid", nil))
	assert.Equal(t, http.StatusOK, w.Code)
	head := w.Header().Get("X-Request-ID")
	assert.NotEmpty(t, head)
	assert.Contains(t, w.Body.String(), head)
}
