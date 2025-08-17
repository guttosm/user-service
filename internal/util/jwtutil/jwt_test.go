package jwtutil

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/guttosm/user-service/config"
)

func tdCfg(expHours int) config.JWTConfig {
	return config.JWTConfig{
		Secret:         "test-secret-123",
		ExpirationHour: expHours,
		Issuer:         "user-service-test",
	}
}

func signRS256(claims jwt.MapClaims) (string, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", err
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return tok.SignedString(priv)
}

func TestJWTService(t *testing.T) {
	now := time.Now().UTC()
	cfg := tdCfg(1)
	validToken, err := generateToken("u-123", "admin", cfg)
	require.NoError(t, err)

	expiredClaims := jwt.MapClaims{
		"user_id": "u-exp",
		"role":    "user",
		"exp":     now.Add(-2 * time.Minute).Unix(),
		"iss":     cfg.Issuer,
	}
	expiredToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims).SignedString([]byte(cfg.Secret))
	require.NoError(t, err)

	rsToken, err := signRS256(jwt.MapClaims{
		"user_id": "u-rs",
		"role":    "user",
		"exp":     now.Add(30 * time.Minute).Unix(),
		"iss":     cfg.Issuer,
	})
	require.NoError(t, err)

	type want struct {
		err       bool
		userID    string
		role      string
		issuer    string
		expDeltaS int
	}

	tests := []struct {
		name  string
		setup func() TokenService
		token func() string
		want  want
	}{
		{
			name:  "valid token → claims ok",
			setup: func() TokenService { return NewJWTService(cfg) },
			token: func() string { return validToken },
			want:  want{err: false, userID: "u-123", role: "admin", issuer: cfg.Issuer, expDeltaS: 6},
		},
		{
			name: "wrong secret → invalid",
			setup: func() TokenService {
				w := cfg
				w.Secret = "another-secret"
				return NewJWTService(w)
			},
			token: func() string { return validToken },
			want:  want{err: true},
		},
		{
			name:  "expired token → invalid",
			setup: func() TokenService { return NewJWTService(cfg) },
			token: func() string { return expiredToken },
			want:  want{err: true},
		},
		{
			name:  "wrong algorithm (RS256) → invalid",
			setup: func() TokenService { return NewJWTService(cfg) },
			token: func() string { return rsToken },
			want:  want{err: true},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc := tc.setup()
			claims, err := svc.Validate(tc.token())

			if tc.want.err {
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidToken)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want.userID, claims["user_id"])
			assert.Equal(t, tc.want.role, claims["role"])
			assert.Equal(t, tc.want.issuer, claims["iss"])

			expFloat, ok := claims["exp"].(float64)
			require.True(t, ok, "exp should be numeric")
			exp := time.Unix(int64(expFloat), 0)
			assert.WithinDuration(t, now.Add(time.Hour), exp, time.Duration(tc.want.expDeltaS)*time.Second)
		})
	}
}
