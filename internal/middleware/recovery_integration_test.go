//go:build integration

package middleware

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecoveryMiddleware_Integration_TableDriven(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RecoveryMiddleware())

	router.GET("/panic", func(c *gin.Context) { panic("kaboom") })
	router.GET("/ok", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"data": "alive"}) })

	srv := httptest.NewServer(router)
	defer srv.Close()

	tests := []struct {
		name          string
		path          string
		wantStatus    int
		wantBodySub   string
		wantJSONField string
		wantJSONValue any
	}{
		{
			name:          "panic → 500 JSON envelope",
			path:          "/panic",
			wantStatus:    http.StatusInternalServerError,
			wantBodySub:   "Internal server error",
			wantJSONField: "message",
			wantJSONValue: "Internal server error",
		},
		{
			name:          "ok → 200 JSON",
			path:          "/ok",
			wantStatus:    http.StatusOK,
			wantJSONField: "data",
			wantJSONValue: "alive",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, srv.URL+tc.path, nil)
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {

				}
			}(resp.Body)

			assert.Equal(t, tc.wantStatus, resp.StatusCode)

			var asJSON map[string]any
			_ = json.NewDecoder(resp.Body).Decode(&asJSON)

			if len(asJSON) == 0 && tc.wantBodySub != "" {
			}

			if tc.wantJSONField != "" {
				val, ok := asJSON[tc.wantJSONField]
				require.True(t, ok, "json should contain key %q; got: %#v", tc.wantJSONField, asJSON)
				if tc.wantJSONValue != nil {
					assert.Equal(t, tc.wantJSONValue, val)
				}
			}
		})
	}
}
