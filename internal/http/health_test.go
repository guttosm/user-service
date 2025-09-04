package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	okHandler := NewHealthHandler(func() error { return nil })
	errHandler := NewHealthHandler(func() error { return errors.New("db down") })

	t.Run("/healthz always ok", func(t *testing.T) {
		r := gin.New()
		okHandler.Register(r)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/healthz", nil))
		require.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "ok")
	})

	t.Run("/readyz ready", func(t *testing.T) {
		r := gin.New()
		okHandler.Register(r)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/readyz", nil))
		require.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "ready")
	})

	t.Run("/readyz degraded on error", func(t *testing.T) {
		r := gin.New()
		errHandler.Register(r)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/readyz", nil))
		require.Equal(t, 503, w.Code)
		assert.Contains(t, w.Body.String(), "degraded")
	})
}
