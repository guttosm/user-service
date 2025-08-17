package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRecoveryMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(RecoveryMiddleware())

	r.GET("/panic", func(c *gin.Context) { panic("boom") })
	r.GET("/ok", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	r.GET("/badreq", func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "bad"})
	})

	type out struct {
		status                int
		bodyContains          string
		expectJSONContentType bool
		checkLogContains      []string
	}

	tests := []struct {
		name          string
		path          string
		captureStdout bool
		out           out
	}{
		{
			name:          "panic → 500 JSON and logs",
			path:          "/panic",
			captureStdout: true,
			out: out{
				status:                http.StatusInternalServerError,
				bodyContains:          "Internal server error",
				expectJSONContentType: true,
				checkLogContains:      []string{"[PANIC RECOVERED]", "boom"},
			},
		},
		{
			name: "no panic → next handler runs",
			path: "/ok",
			out: out{
				status:       http.StatusOK,
				bodyContains: "ok",
			},
		},
		{
			name: "aborted 400 (no panic) passes through",
			path: "/badreq",
			out: out{
				status:                http.StatusBadRequest,
				bodyContains:          "bad",
				expectJSONContentType: true,
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var restore func()
			var buf bytes.Buffer
			if tc.captureStdout {
				old := os.Stdout
				rpr, wp, _ := os.Pipe()
				os.Stdout = wp
				restore = func() {
					_ = wp.Close()
					os.Stdout = old

					_, _ = io.Copy(&buf, rpr)
					_ = rpr.Close()
				}
			}

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			r.ServeHTTP(w, req)

			if restore != nil {
				restore()
			}

			assert.Equal(t, tc.out.status, w.Code)

			body := w.Body.String()
			if tc.out.bodyContains != "" {
				assert.Contains(t, body, tc.out.bodyContains)
			}

			if tc.out.expectJSONContentType {
				ct := w.Header().Get("Content-Type")
				assert.Contains(t, ct, "application/json")
			}

			for _, sub := range tc.out.checkLogContains {
				assert.True(t, strings.Contains(buf.String(), sub), "log should contain %q; got: %s", sub, buf.String())
			}
		})
	}
}
