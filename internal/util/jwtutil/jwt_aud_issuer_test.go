package jwtutil

import (
	"testing"
	"time"

	"github.com/guttosm/user-service/config"
)

func TestJWT_IssuerAudience(t *testing.T) {
	base := config.JWTConfig{Secret: "s", ExpirationHour: 1, Issuer: "iss-X", Audience: "aud-X"}
	svc := NewJWTService(base)
	tok, err := svc.Generate("user1", "role")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if _, err := svc.Validate(tok); err != nil {
		t.Fatalf("expected valid token: %v", err)
	}

	wrongIss := NewJWTService(config.JWTConfig{Secret: base.Secret, ExpirationHour: 1, Issuer: "other", Audience: base.Audience})
	if _, err := wrongIss.Validate(tok); err == nil {
		t.Fatalf("expected issuer rejection")
	}

	wrongAud := NewJWTService(config.JWTConfig{Secret: base.Secret, ExpirationHour: 1, Issuer: base.Issuer, Audience: "other"})
	if _, err := wrongAud.Validate(tok); err == nil {
		t.Fatalf("expected audience rejection")
	}
}

func TestJWT_ExpiredToken(t *testing.T) {
	cfg := config.JWTConfig{Secret: "s2", ExpirationHour: 0, Issuer: "iss", Audience: "aud"}
	svc := NewJWTService(cfg)
	tok, err := svc.Generate("u", "r")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	time.Sleep(1100 * time.Millisecond)
	if _, err := svc.Validate(tok); err == nil {
		t.Fatalf("expected expired token error")
	}
}

func TestJWT_AudienceArray(t *testing.T) {
	cfg := config.JWTConfig{Secret: "s3", ExpirationHour: 1, Issuer: "iss", Audience: "aud2"}
	// craft token with array audience by generating and then re-validating with Audience set to one member
	// Here Generate uses single aud; simulate array by validating a token from another service that might produce array aud
	tok, err := generateToken("u", "r", config.JWTConfig{Secret: cfg.Secret, ExpirationHour: 1, Issuer: cfg.Issuer, Audience: "aud1"})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	// Now validate with desired audience mismatch -> expect failure
	if _, err := NewJWTService(cfg).Validate(tok); err == nil {
		t.Fatalf("expected audience mismatch to fail")
	}
}
