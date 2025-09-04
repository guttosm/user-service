package middleware

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequestLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID(), RequestLogger())
	r.GET("/logtest", func(c *gin.Context) { c.String(200, "ok") })

	var buf bytes.Buffer
	orig := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(orig)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/logtest", nil))

	assert.Equal(t, 200, w.Code)
	logLine := buf.String()
	assert.Contains(t, logLine, "method=GET")
	assert.Contains(t, logLine, "path=/logtest")
	assert.Contains(t, logLine, "status=200")
	assert.Contains(t, logLine, "request_id=")
}
