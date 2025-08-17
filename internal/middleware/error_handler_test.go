package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(ErrorHandler)

	r.GET("/ok", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	r.GET("/err", func(c *gin.Context) {
		_ = c.Error(errors.New("boom"))
	})

	r.GET("/written_with_error", func(c *gin.Context) {
		_ = c.Error(errors.New("ignored after write"))
		c.JSON(http.StatusTeapot, gin.H{"status": "i am a teapot"}) // writer is already written
	})

	r.GET("/abort_helper", func(c *gin.Context) {
		AbortWithError(c, http.StatusBadRequest, "bad input", errors.New("oops"))
		// Ensure any next handler would NOT run
	})

	r.GET("/abort_chain", func(c *gin.Context) {
		AbortWithError(c, http.StatusUnauthorized, "unauthorized", nil)
	}, func(c *gin.Context) {
		// this SHOULD NOT run if abort works
		c.Header("X-Should-Not-Run", "true")
	})

	r.GET("/multi", func(c *gin.Context) {
		_ = c.Error(errors.New("first"))
		_ = c.Error(errors.New("second"))
	})

	type want struct {
		status             int
		bodySubstr         string
		expectJSONContent  bool
		headerMustNotExist string
	}

	tests := []struct {
		name string
		path string
		want want
	}{
		{
			name: "no error → pass through",
			path: "/ok",
			want: want{
				status:            http.StatusOK,
				bodySubstr:        "ok",
				expectJSONContent: false,
			},
		},
		{
			name: "error collected → 500 JSON",
			path: "/err",
			want: want{
				status:            http.StatusInternalServerError,
				bodySubstr:        "An unexpected error occurred",
				expectJSONContent: true,
			},
		},
		{
			name: "already written → do not override",
			path: "/written_with_error",
			want: want{
				status:            http.StatusTeapot,
				bodySubstr:        "teapot",
				expectJSONContent: true,
			},
		},
		{
			name: "AbortWithError → 400 JSON and abort",
			path: "/abort_helper",
			want: want{
				status:            http.StatusBadRequest,
				bodySubstr:        "bad input",
				expectJSONContent: true,
			},
		},
		{
			name: "AbortWithError aborts the chain (next handler not executed)",
			path: "/abort_chain",
			want: want{
				status:             http.StatusUnauthorized,
				bodySubstr:         "unauthorized",
				expectJSONContent:  true,
				headerMustNotExist: "X-Should-Not-Run",
			},
		},
		{
			name: "multiple errors → still 500 JSON standard message",
			path: "/multi",
			want: want{
				status:            http.StatusInternalServerError,
				bodySubstr:        "An unexpected error occurred",
				expectJSONContent: true,
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			r.ServeHTTP(w, req)

			require.Equal(t, tc.want.status, w.Code)

			body := w.Body.String()
			if tc.want.bodySubstr != "" {
				assert.True(t, strings.Contains(body, tc.want.bodySubstr), "body=%s", body)
			}

			if tc.want.expectJSONContent {
				assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
			} else {
				assert.NotContains(t, w.Header().Get("Content-Type"), "application/json")
			}

			if tc.want.headerMustNotExist != "" {
				_, present := w.Result().Header[tc.want.headerMustNotExist]
				assert.False(t, present, "header %q must not be set", tc.want.headerMustNotExist)
			}
		})
	}
}
