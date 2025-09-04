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
