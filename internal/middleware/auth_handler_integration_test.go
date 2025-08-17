//go:build integration

package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/guttosm/user-service/config"
	"github.com/guttosm/user-service/internal/util/jwtutil"
)

func TestAuthMiddleware_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.JWTConfig{
		Secret:         "int-secret",
		ExpirationHour: 1,
		Issuer:         "user-service-int",
	}
	svc := jwtutil.NewJWTService(cfg)

	newRouter := func(svc jwtutil.TokenService) *gin.Engine {
		r := gin.New()
		r.Use(AuthMiddleware(svc))
		r.GET("/me", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"user_id": c.GetString("user_id"),
			})
		})
		return r
	}

	validTok, err := svc.Generate("u-42", "member")
	require.NoError(t, err)

	wrongCfg := cfg
	wrongCfg.Secret = "other"
	wrongSvc := jwtutil.NewJWTService(wrongCfg)
	wrongTok, err := wrongSvc.Generate("u-99", "member")
	require.NoError(t, err)

	r := newRouter(svc)

	type want struct {
		status int
		substr string
	}

	tests := []struct {
		name       string
		authHeader string
		want       want
	}{
		{
			name:       "missing header → 401",
			authHeader: "",
			want:       want{status: http.StatusUnauthorized, substr: "Missing or malformed Authorization header"},
		},
		{
			name:       "valid token → 200 and user_id available",
			authHeader: "Bearer " + validTok,
			want:       want{status: http.StatusOK, substr: "u-42"},
		},
		{
			name:       "wrong secret → 401",
			authHeader: "Bearer " + wrongTok,
			want:       want{status: http.StatusUnauthorized, substr: "Invalid or expired token"},
		},
	}

	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, ts.URL+"/me", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tc.want.status, resp.StatusCode)
			
			buf := new(strings.Builder)
			_, _ = io.Copy(buf, resp.Body)
			assert.Contains(t, buf.String(), tc.want.substr)
		})
	}
}
