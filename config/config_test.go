package config

import (
	"os"
	"testing"
)

func TestLoadConfig_FromEnv(t *testing.T) {
	t.Setenv("SERVER_PORT", "9090")
	t.Setenv("POSTGRES_HOST", "pg-host")
	t.Setenv("POSTGRES_PORT", "5432")
	t.Setenv("POSTGRES_USER", "user")
	t.Setenv("POSTGRES_PASSWORD", "pass")
	t.Setenv("POSTGRES_DB", "db")
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("JWT_ISSUER", "issuer")
	t.Setenv("JWT_AUDIENCE", "aud")

	// Ensure no .env interferes
	_ = os.Remove(".env")

	LoadConfig()

	if AppConfig.Server.Port != "9090" {
		t.Fatalf("expected port 9090, got %s", AppConfig.Server.Port)
	}
	if AppConfig.Postgres.Host != "pg-host" || AppConfig.Postgres.DBName != "db" {
		t.Fatalf("postgres config not loaded from env: %+v", AppConfig.Postgres)
	}
	if AppConfig.JWT.Secret != "secret" || AppConfig.JWT.Issuer != "issuer" || AppConfig.JWT.Audience != "aud" {
		t.Fatalf("jwt config not loaded correctly: %+v", AppConfig.JWT)
	}
}
