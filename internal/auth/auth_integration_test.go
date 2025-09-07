//go:build integration

package auth

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/guttosm/user-service/config"
	mlogger "github.com/guttosm/user-service/internal/repository/mongo"
	repo "github.com/guttosm/user-service/internal/repository/postgres"
	"github.com/guttosm/user-service/internal/util/jwtutil"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	driverMongo "go.mongodb.org/mongo-driver/mongo"
	driverOptions "go.mongodb.org/mongo-driver/mongo/options"
)

func startPostgresAuth(t *testing.T) (*sql.DB, func()) {
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
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{ContainerRequest: req, Started: true})
	require.NoError(t, err)
	host, err := c.Host(ctx)
	require.NoError(t, err)
	port, err := c.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)
	dsn := "postgres://app:pass@" + host + ":" + port.Port() + "/app?sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	require.Eventually(t, func() bool { return db.Ping() == nil }, 25*time.Second, 250*time.Millisecond)

	_, err = db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"; CREATE EXTENSION IF NOT EXISTS "pgcrypto";`)
	require.NoError(t, err)
	_, err = db.Exec(`CREATE OR REPLACE FUNCTION uuid_generate_v7() RETURNS UUID AS $$ DECLARE unix_ts_ms BIGINT; uuid_bytes BYTEA; BEGIN unix_ts_ms := EXTRACT(EPOCH FROM NOW()) * 1000; uuid_bytes := substring(int8send(unix_ts_ms), 3, 6) || substring(gen_random_bytes(2), 1, 2) || substring(gen_random_bytes(8), 1, 8); uuid_bytes := set_byte(uuid_bytes, 6, (get_byte(uuid_bytes, 6) & 15) | 112); uuid_bytes := set_byte(uuid_bytes, 8, (get_byte(uuid_bytes, 8) & 63) | 128); RETURN encode(uuid_bytes, 'hex')::UUID; END; $$ LANGUAGE plpgsql VOLATILE;`)
	require.NoError(t, err)
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (id UUID PRIMARY KEY DEFAULT uuid_generate_v7(), email VARCHAR(255) UNIQUE NOT NULL, password VARCHAR(255) NOT NULL, role VARCHAR(50) NOT NULL DEFAULT 'user');`)
	require.NoError(t, err)

	cleanup := func() { _ = db.Close(); _ = c.Terminate(ctx) }
	return db, cleanup
}

func startMongoAuth(t *testing.T) (*driverMongo.Client, string, func()) {
	t.Helper()
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "mongo:7",
		ExposedPorts: []string{"27017/tcp"},
		WaitingFor:   wait.ForListeningPort("27017/tcp").WithStartupTimeout(90 * time.Second),
	}
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{ContainerRequest: req, Started: true})
	require.NoError(t, err)
	host, err := c.Host(ctx)
	require.NoError(t, err)
	port, err := c.MappedPort(ctx, "27017/tcp")
	require.NoError(t, err)
	uri := "mongodb://" + host + ":" + port.Port()
	cli, err := driverMongo.Connect(ctx, driverOptions.Client().ApplyURI(uri))
	require.NoError(t, err)
	require.Eventually(t, func() bool { return cli.Ping(ctx, nil) == nil }, 20*time.Second, 200*time.Millisecond)
	cleanup := func() { _ = cli.Disconnect(context.Background()); _ = c.Terminate(ctx) }
	return cli, uri, cleanup
}

func TestAuth_Integration_RegisterLogin(t *testing.T) {
	db, pgCleanup := startPostgresAuth(t)
	defer pgCleanup()
	cli, _, mongoCleanup := startMongoAuth(t)
	defer mongoCleanup()

	repo := repo.NewUserRepository(db)
	tokens := jwtutil.NewJWTService(config.JWTConfig{Secret: "s", Issuer: "i", Audience: "a", ExpirationHour: 1})
	logger := mlogger.NewLogger(cli.Database("user_service_test"), "auth_logs")

	svc := NewAuthService(repo, tokens, logger)
	ctx := context.Background()

	u, err := svc.Register(ctx, "it@ex.com", "pw", "user")
	require.NoError(t, err)
	require.NotEmpty(t, u.ID)

	tok, err := svc.Login(ctx, "it@ex.com", "pw")
	require.NoError(t, err)
	require.NotEmpty(t, tok)
}
