package app

import (
	"testing"

	"github.com/guttosm/user-service/config"
)

// Ensure InitPostgres returns an error when connection parameters are invalid
func TestInitPostgres_FailsOnInvalidConn(t *testing.T) {
	cfg := &config.Config{Postgres: config.PostgresConfig{
		Host: "127.0.0.1", Port: "0", User: "u", Password: "p", DBName: "db", SSLMode: "disable",
	}}
	if _, err := InitPostgres(cfg); err == nil {
		t.Fatalf("expected error on invalid connection params")
	}
}

// Verify InitMongo returns an error with invalid URI
func TestInitMongo_ErrorOnInvalidURI(t *testing.T) {
	cfg := &config.Config{Mongo: config.MongoConfig{URI: "mongodb://127.0.0.1:0", Name: "test"}}
	_, _, cleanup, err := InitMongo(cfg)
	if cleanup != nil {
		cleanup()
	}
	if err == nil {
		t.Fatalf("expected error for invalid mongo uri")
	}
}

// Ensure InitializeApp fails when config is missing or invalid connections; without real Postgres/Mongo it should error.
func TestInitializeApp_FailsWithoutDeps(t *testing.T) {
	// Set minimal config expected by InitializeApp with invalid DBs
	config.AppConfig = &config.Config{
		Server:   config.ServerConfig{Port: "8080"},
		Postgres: config.PostgresConfig{Host: "127.0.0.1", Port: "0", User: "u", Password: "p", DBName: "db", SSLMode: "disable"},
		JWT:      config.JWTConfig{Secret: "s", Issuer: "i", Audience: "a", ExpirationHour: 1},
		Mongo:    config.MongoConfig{URI: "mongodb://127.0.0.1:0", Name: "x"},
	}
	if _, _, err := InitializeApp(); err == nil {
		t.Fatalf("expected initialization to fail without databases")
	}
}
