package jwtutil

import (
	"testing"

	"github.com/golang-jwt/jwt/v4"
	"github.com/guttosm/user-service/config"
)

// Ensure we reject tokens not signed with HMAC (e.g., 'none' algo)
func TestJWT_RejectsNonHMACAlgorithms(t *testing.T) {
	cfg := config.JWTConfig{Secret: "s", ExpirationHour: 1, Issuer: "iss", Audience: "aud"}
	svc := NewJWTService(cfg)

	claims := jwt.MapClaims{"user_id": "u", "role": "r", "iss": cfg.Issuer, "aud": cfg.Audience}
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	// note: requires explicit opt for unsafe signing
	tokStr, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("sign none: %v", err)
	}

	if _, err := svc.Validate(tokStr); err == nil {
		t.Fatalf("expected validation failure for non-HMAC algorithm")
	}
}
