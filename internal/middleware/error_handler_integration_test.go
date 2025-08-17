//go:build integration

package middleware

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorHandler_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(ErrorHandler)

	router.GET("/panicfree_ok", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	router.GET("/collect_error", func(c *gin.Context) {
		_ = c.Error(assert.AnError) // any error; middleware formats the standardized body
	})

	router.GET("/prewritten", func(c *gin.Context) {
		_ = c.Error(assert.AnError)
		c.JSON(http.StatusTeapot, gin.H{"msg": "i am a teapot"})
	})

	router.GET("/abort", func(c *gin.Context) {
		AbortWithError(c, http.StatusForbidden, "nope", nil)
	})

	srv := httptest.NewServer(router)
	defer srv.Close()

	type check func(t *testing.T, resp *http.Response)

	tests := []struct {
		name       string
		urlPath    string
		wantStatus int
		verify     check
	}{
		{
			name:       "OK path returns 200 JSON",
			urlPath:    "/panicfree_ok",
			wantStatus: http.StatusOK,
			verify: func(t *testing.T, resp *http.Response) {
				defer resp.Body.Close()
				var m map[string]any
				require.NoError(t, json.NewDecoder(resp.Body).Decode(&m))
				assert.Equal(t, true, m["ok"])
			},
		},
		{
			name:       "Collected error → 500 JSON with standard message",
			urlPath:    "/collect_error",
			wantStatus: http.StatusInternalServerError,
			verify: func(t *testing.T, resp *http.Response) {
				defer resp.Body.Close()
				raw, _ := io.ReadAll(resp.Body)
				assert.True(t, strings.Contains(string(raw), "An unexpected error occurred"), "body=%s", string(raw))
				assert.Contains(t, resp.Header.Get("Content-Type"), "application/json")
			},
		},
		{
			name:       "Prewritten response not overridden",
			urlPath:    "/prewritten",
			wantStatus: http.StatusTeapot,
			verify: func(t *testing.T, resp *http.Response) {
				defer resp.Body.Close()
				var m map[string]any
				require.NoError(t, json.NewDecoder(resp.Body).Decode(&m))
				assert.Equal(t, "i am a teapot", m["msg"])
			},
		},
		{
			name:       "AbortWithError returns provided status and JSON",
			urlPath:    "/abort",
			wantStatus: http.StatusForbidden,
			verify: func(t *testing.T, resp *http.Response) {
				defer resp.Body.Close()
				raw, _ := io.ReadAll(resp.Body)
				assert.True(t, strings.Contains(string(raw), "nope"), "body=%s", string(raw))
				assert.Contains(t, resp.Header.Get("Content-Type"), "application/json")
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, srv.URL+tc.urlPath, nil)
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			require.Equal(t, tc.wantStatus, resp.StatusCode)
			tc.verify(t, resp)
		})
	}
}
