package middleware

import (
	"net/http"
	"net/http/httptest"
	_ "strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mjwt "github.com/guttosm/user-service/internal/mocks/jwtutil"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type exp struct {
		expectValidateToken string
		returnClaims        map[string]any
		returnErr           error
	}
	type want struct {
		status       int
		bodyContains string
		expectUserID string
	}

	tests := []struct {
		name       string
		authHeader string
		exp        exp
		want       want
	}{
		{
			name:       "missing header → 401",
			authHeader: "",
			want:       want{status: http.StatusUnauthorized, bodyContains: "Missing or malformed Authorization header"},
		},
		{
			name:       "malformed header (no token) → 401",
			authHeader: "Bearer",
			want:       want{status: http.StatusUnauthorized, bodyContains: "Missing or malformed Authorization header"},
		},
		{
			name:       "invalid token → 401",
			authHeader: "Bearer badtok",
			exp:        exp{expectValidateToken: "badtok", returnErr: assert.AnError},
			want:       want{status: http.StatusUnauthorized, bodyContains: "Invalid or expired token"},
		},
		{
			name:       "missing user_id claim → 401",
			authHeader: "Bearer good",
			exp:        exp{expectValidateToken: "good", returnClaims: map[string]any{}},
			want:       want{status: http.StatusUnauthorized, bodyContains: "Missing 'user_id' claim"},
		},
		{
			name:       "empty user_id → 401",
			authHeader: "Bearer ok",
			exp:        exp{expectValidateToken: "ok", returnClaims: map[string]any{"user_id": ""}},
			want:       want{status: http.StatusUnauthorized, bodyContains: "non-empty string"},
		},
		{
			name:       "non-string user_id → 401",
			authHeader: "Bearer x",
			exp:        exp{expectValidateToken: "x", returnClaims: map[string]any{"user_id": 123}},
			want:       want{status: http.StatusUnauthorized, bodyContains: "non-empty string"},
		},
		{
			name:       "happy path → 200",
			authHeader: "Bearer tok",
			exp:        exp{expectValidateToken: "tok", returnClaims: map[string]any{"user_id": "u-1"}},
			want:       want{status: http.StatusOK, expectUserID: "u-1"},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			mock := mjwt.NewMockTokenService(t)
			if tc.exp.expectValidateToken != "" {
				mock.EXPECT().
					Validate(tc.exp.expectValidateToken).
					Return(tc.exp.returnClaims, tc.exp.returnErr)
			}

			r := gin.New()
			r.Use(AuthMiddleware(mock))
			r.GET("/me", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"user_id": c.GetString("user_id")})
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/me", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			r.ServeHTTP(w, req)

			require.Equal(t, tc.want.status, w.Code)
			body := w.Body.String()
			if tc.want.bodyContains != "" {
				assert.Contains(t, body, tc.want.bodyContains)
			}
			if tc.want.expectUserID != "" {
				assert.Contains(t, body, tc.want.expectUserID)
			}
		})
	}
}
