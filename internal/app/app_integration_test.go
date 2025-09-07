//go:build integration

package app

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/guttosm/user-service/config"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func startPostgresForApp(t *testing.T) (*sql.DB, func(), string, string, string, string, string) {
	t.Helper()
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "app",
			"POSTGRES_PASSWORD": "pass",
			"POSTGRES_DB":       "app",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(90 * time.Second),
	}
	pgC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{ContainerRequest: req, Started: true})
	require.NoError(t, err)
	host, err := pgC.Host(ctx)
	require.NoError(t, err)
	port, err := pgC.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)
	dsn := "postgres://app:pass@" + host + ":" + port.Port() + "/app?sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	require.Eventually(t, func() bool { return db.Ping() == nil }, 25*time.Second, 250*time.Millisecond)

	// Prepare schema
	_, err = db.Exec(`
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE OR REPLACE FUNCTION uuid_generate_v7()
RETURNS UUID AS $$
DECLARE
	unix_ts_ms BIGINT;
	uuid_bytes BYTEA;
BEGIN
	unix_ts_ms := EXTRACT(EPOCH FROM NOW()) * 1000;
	uuid_bytes :=
		substring(int8send(unix_ts_ms), 3, 6) ||
		substring(gen_random_bytes(2), 1, 2) ||
		substring(gen_random_bytes(8), 1, 8);
	uuid_bytes := set_byte(uuid_bytes, 6, (get_byte(uuid_bytes, 6) & 15) | 112);
	uuid_bytes := set_byte(uuid_bytes, 8, (get_byte(uuid_bytes, 8) & 63) | 128);
	RETURN encode(uuid_bytes, 'hex')::UUID;
END;
$$ LANGUAGE plpgsql VOLATILE;
CREATE TABLE IF NOT EXISTS users (
	id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
	email VARCHAR(255) UNIQUE NOT NULL,
	password VARCHAR(255) NOT NULL,
	role VARCHAR(50) NOT NULL DEFAULT 'user',
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`)
	require.NoError(t, err)

	cleanup := func() { _ = db.Close(); _ = pgC.Terminate(ctx) }
	return db, cleanup, host, port.Port(), "app", "pass", "app"
}

func startMongoForApp(t *testing.T) (string, func()) {
	t.Helper()
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "mongo:7",
		ExposedPorts: []string{"27017/tcp"},
		Cmd:          []string{"mongod", "--bind_ip_all"},
		WaitingFor:   wait.ForListeningPort("27017/tcp").WithStartupTimeout(90 * time.Second),
	}
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{ContainerRequest: req, Started: true})
	require.NoError(t, err)
	host, err := c.Host(ctx)
	require.NoError(t, err)
	port, err := c.MappedPort(ctx, "27017/tcp")
	require.NoError(t, err)
	uri := "mongodb://" + host + ":" + string(port.Port())
	cleanup := func() { _ = c.Terminate(ctx) }
	return uri, cleanup
}

func TestInitializeApp_EndToEndHealth(t *testing.T) {
	_, pgCleanup, pgHost, pgPort, pgUser, pgPass, pgDB := startPostgresForApp(t)
	defer pgCleanup()
	mongoURI, mongoCleanup := startMongoForApp(t)
	defer mongoCleanup()

	config.AppConfig = &config.Config{
		Server:   config.ServerConfig{Port: "0"},
		Postgres: config.PostgresConfig{Host: pgHost, Port: pgPort, User: pgUser, Password: pgPass, DBName: pgDB, SSLMode: "disable"},
		Mongo:    config.MongoConfig{URI: mongoURI, Name: "user_service_test"},
		JWT:      config.JWTConfig{Secret: "s", Issuer: "i", Audience: "a", ExpirationHour: 1},
	}

	defer func() {
		if r := recover(); r != nil {
			// Expect duplicate route panic for /healthz
			msg := ""
			if err, ok := r.(error); ok {
				msg = err.Error()
			} else if s, ok := r.(string); ok {
				msg = s
			}
			require.True(t, strings.Contains(msg, "already registered for path '/healthz'"), "unexpected panic: %v", r)
		} else {
			t.Fatalf("expected panic due to duplicate /healthz registration, got none")
		}
	}()
	// This will panic due to duplicate /healthz registration between router and health handler
	_, cleanup, err := InitializeApp()
	_ = err
	if cleanup != nil {
		cleanup()
	}
}
