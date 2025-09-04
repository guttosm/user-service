//go:build integration

package jwtutil

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/guttosm/user-service/config"
)

func TestGinAuthMiddleware_WithJWTService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.JWTConfig{Secret: "gin-int-secret", ExpirationHour: 1, Issuer: "user-service-int", Audience: "aud-int"}
	svc := NewJWTService(cfg)

	authMW := func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		claims, err := svc.Validate(tokenStr)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Set("claims", claims)
		c.Next()
	}

	r := gin.New()
	r.GET("/me", authMW, func(c *gin.Context) {
		claims := c.MustGet("claims").(map[string]any)
		c.JSON(http.StatusOK, claims)
	})
	ts := httptest.NewServer(r)
	defer ts.Close()

	validToken, err := svc.Generate("u-42", "member")
	require.NoError(t, err)

	wrongCfg := cfg
	wrongCfg.Secret = "oops"
	wrongToken, err := generateToken("u-1", "x", wrongCfg)
	require.NoError(t, err)

	expiredCfg := cfg
	expiredCfg.ExpirationHour = 0
	wrongAudienceSvc := NewJWTService(config.JWTConfig{Secret: cfg.Secret, ExpirationHour: 1, Issuer: cfg.Issuer, Audience: "other-aud"})
	expiredToken, err := generateToken("u-99", "ghost", expiredCfg)
	require.NoError(t, err)
	time.Sleep(1 * time.Second)

	tests := []struct {
		name       string
		token      string
		wantStatus int
		wantUserID string
	}{
		{
			name:       "valid token",
			token:      validToken,
			wantStatus: http.StatusOK,
			wantUserID: "u-42",
		},
		{
			name:       "wrong audience",
			token:      func() string { tk, _ := wrongAudienceSvc.Generate("u-77", "user"); return tk }(),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid secret",
			token:      wrongToken,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "expired token",
			token:      expiredToken,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "missing token",
			token:      "",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, ts.URL+"/me", nil)
			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					t.Errorf("Failed to close response body: %v", err)
				}
			}(resp.Body)

			assert.Equal(t, tc.wantStatus, resp.StatusCode)

			if resp.StatusCode == http.StatusOK {
				var body map[string]any
				require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
				assert.Equal(t, tc.wantUserID, body["user_id"])
			}
		})
	}
}
