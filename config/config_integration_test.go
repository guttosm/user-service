//go:build integration

package config

import (
	"os"
	"testing"
)

func TestLoadConfig_FromDotEnv(t *testing.T) {
	content := "SERVER_PORT=7070\nPOSTGRES_HOST=envhost\nJWT_SECRET=envsecret\nJWT_ISSUER=iss\nJWT_AUDIENCE=aud\n"
	if err := os.WriteFile(".env", []byte(content), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}
	defer os.Remove(".env")

	// Clear env to ensure .env is used
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("POSTGRES_HOST")
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("JWT_ISSUER")
	os.Unsetenv("JWT_AUDIENCE")

	LoadConfig()
	if AppConfig.Server.Port != "7070" || AppConfig.Postgres.Host != "envhost" || AppConfig.JWT.Secret != "envsecret" {
		t.Fatalf("expected values from .env, got: %+v", AppConfig)
	}
}
