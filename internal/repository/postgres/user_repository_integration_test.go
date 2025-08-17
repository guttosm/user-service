//go:build integration

package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/guttosm/user-service/internal/domain/model"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func startPostgresWithMigrations(t *testing.T) (*sql.DB, func()) {
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
	pgC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, err := pgC.Host(ctx)
	require.NoError(t, err)
	port, err := pgC.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)

	dsn := "postgres://app:pass@" + host + ":" + port.Port() + "/app?sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)

	require.Eventually(t, func() bool { return db.Ping() == nil }, 25*time.Second, 250*time.Millisecond)

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
`)
	require.NoError(t, err)

	_, err = db.Exec(`
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
`)
	require.NoError(t, err)

	cleanup := func() {
		_ = db.Close()
		_ = pgC.Terminate(ctx)
	}
	return db, cleanup
}

func truncateUsers(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec(`TRUNCATE TABLE users RESTART IDENTITY;`)
	require.NoError(t, err)
}

func TestPostgresUserRepository_Integration(t *testing.T) {
	db, cleanup := startPostgresWithMigrations(t)
	defer cleanup()

	repo := NewUserRepository(db)

	type op string
	const (
		opCreate    op = "Create"
		opFindEmail op = "FindByEmail"
		opFindID    op = "FindByID"
	)

	type args struct {
		u     *model.User
		email string
		id    string
		seed  *model.User
	}
	type exp struct {
		err   bool
		nil   bool
		check func(t *testing.T, got *model.User, in args)
	}

	mk := func(email, pw, role string) *model.User {
		return &model.User{Email: email, Password: pw, Role: role}
	}

	tests := []struct {
		name string
		op   op
		in   args
		exp  exp
	}{
		{
			name: "Create/success",
			op:   opCreate,
			in:   args{u: mk("ok@case.com", "pwhash", "member")},
			exp: exp{
				err: false,
				check: func(t *testing.T, _ *model.User, in args) {
					got, err := repo.FindByEmail(in.u.Email)
					require.NoError(t, err)
					require.NotNil(t, got)
					assert.Equal(t, in.u.Email, got.Email)
					assert.Equal(t, in.u.Role, got.Role)
					assert.NotEmpty(t, got.ID) // should be UUID v7 from DB default
				},
			},
		},
		{
			name: "Create/duplicate_email",
			op:   opCreate,
			in:   args{u: mk("dup@x.com", "pw", "member"), seed: mk("dup@x.com", "pw", "member")},
			exp:  exp{err: true},
		},
		{
			name: "FindByEmail/found",
			op:   opFindEmail,
			in:   args{email: "hit@x.com", seed: mk("hit@x.com", "pw", "member")},
			exp: exp{err: false, nil: false, check: func(t *testing.T, got *model.User, in args) {
				assert.Equal(t, in.email, got.Email)
				assert.NotEmpty(t, got.ID)
			}},
		},
		{
			name: "FindByEmail/not_found",
			op:   opFindEmail,
			in:   args{email: "missing@x.com", seed: mk("other@x.com", "pw", "member")},
			exp:  exp{err: false, nil: true},
		},
		{
			name: "FindByID/found",
			op:   opFindID,
			in:   args{seed: mk("found@x.com", "pw", "admin")},
			exp: exp{err: false, nil: false, check: func(t *testing.T, got *model.User, _ args) {
				assert.Equal(t, "found@x.com", got.Email)
				assert.NotEmpty(t, got.ID)
			}},
		},
		{
			name: "FindByID/not_found",
			op:   opFindID,
			in:   args{id: "00000000-0000-0000-0000-000000000000", seed: mk("seed@x.com", "pw", "member")},
			exp:  exp{err: false, nil: true},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			truncateUsers(t, db)

			if tc.in.seed != nil {
				require.NoError(t, repo.Create(tc.in.seed))
				if tc.op == opFindID && tc.in.id == "" {
					tc.in.id = tc.in.seed.ID
				}
			}

			switch tc.op {
			case opCreate:
				err := repo.Create(tc.in.u)
				if tc.exp.err {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
					assert.NotEmpty(t, tc.in.u.ID)
				}
				if tc.exp.check != nil {
					tc.exp.check(t, nil, tc.in)
				}

			case opFindEmail:
				got, err := repo.FindByEmail(tc.in.email)
				require.NoError(t, err)
				if tc.exp.nil {
					assert.Nil(t, got)
				} else {
					require.NotNil(t, got)
					if tc.exp.check != nil {
						tc.exp.check(t, got, tc.in)
					}
				}

			case opFindID:
				got, err := repo.FindByID(tc.in.id)
				require.NoError(t, err)
				if tc.exp.nil {
					assert.Nil(t, got)
				} else {
					require.NotNil(t, got)
					if tc.exp.check != nil {
						tc.exp.check(t, got, tc.in)
					}
				}
			}
		})
	}
}
