package repository

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/guttosm/user-service/internal/domain/model"
)

func TestPostgresUserRepository(t *testing.T) {

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewUserRepository(db)

	insertQ := regexp.QuoteMeta(`
		INSERT INTO users (email, password, role)
		VALUES ($1, $2, $3)
		RETURNING id;
	`)
	findByEmailQ := regexp.QuoteMeta(`SELECT id, email, password, role FROM users WHERE email = $1`)
	findByIDQ := regexp.QuoteMeta(`SELECT id, email, password, role FROM users WHERE id = $1`)

	type op string
	const (
		opCreate      op = "Create"
		opFindByEmail op = "FindByEmail"
		opFindByID    op = "FindByID"
	)

	type args struct {
		user  *model.User
		email string
		id    string
	}
	type expect struct {
		setup func()
		err   bool
		nil   bool // for getters
	}

	tests := []struct {
		name string
		op   op
		arg  args
		exp  expect
	}{

		{
			name: "Create/success",
			op:   opCreate,
			arg:  args{user: &model.User{Email: "a@b.com", Password: "hash", Role: "member"}},
			exp: expect{
				setup: func() {
					mock.ExpectQuery(insertQ).
						WithArgs("a@b.com", "hash", "member").
						WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("11111111-1111-1111-1111-111111111111"))
				},
				err: false,
			},
		},
		{
			name: "Create/db_error",
			op:   opCreate,
			arg:  args{user: &model.User{Email: "x@x.com", Password: "pw", Role: "admin"}},
			exp: expect{
				setup: func() {
					mock.ExpectQuery(insertQ).
						WithArgs("x@x.com", "pw", "admin").
						WillReturnError(errors.New("insert failed"))
				},
				err: true,
			},
		},

		{
			name: "FindByEmail/found",
			op:   opFindByEmail,
			arg:  args{email: "hit@x.com"},
			exp: expect{
				setup: func() {
					rows := sqlmock.NewRows([]string{"id", "email", "password", "role"}).
						AddRow("22222222-2222-2222-2222-222222222222", "hit@x.com", "hash", "member")
					mock.ExpectQuery(findByEmailQ).WithArgs("hit@x.com").WillReturnRows(rows)
				},
				err: false, nil: false,
			},
		},
		{
			name: "FindByEmail/not_found",
			op:   opFindByEmail,
			arg:  args{email: "missing@x.com"},
			exp: expect{
				setup: func() {
					mock.ExpectQuery(findByEmailQ).WithArgs("missing@x.com").WillReturnError(sql.ErrNoRows)
				},
				err: false, nil: true,
			},
		},
		{
			name: "FindByEmail/db_error",
			op:   opFindByEmail,
			arg:  args{email: "err@x.com"},
			exp: expect{
				setup: func() {
					mock.ExpectQuery(findByEmailQ).WithArgs("err@x.com").WillReturnError(errors.New("boom"))
				},
				err: true, nil: true,
			},
		},

		{
			name: "FindByID/found",
			op:   opFindByID,
			arg:  args{id: "33333333-3333-3333-3333-333333333333"},
			exp: expect{
				setup: func() {
					rows := sqlmock.NewRows([]string{"id", "email", "password", "role"}).
						AddRow("33333333-3333-3333-3333-333333333333", "z@y.com", "hash2", "admin")
					mock.ExpectQuery(findByIDQ).WithArgs("33333333-3333-3333-3333-333333333333").WillReturnRows(rows)
				},
				err: false, nil: false,
			},
		},
		{
			name: "FindByID/not_found",
			op:   opFindByID,
			arg:  args{id: "00000000-0000-0000-0000-000000000000"},
			exp: expect{
				setup: func() {
					mock.ExpectQuery(findByIDQ).WithArgs("00000000-0000-0000-0000-000000000000").WillReturnError(sql.ErrNoRows)
				},
				err: false, nil: true,
			},
		},
		{
			name: "FindByID/db_error",
			op:   opFindByID,
			arg:  args{id: "bad"},
			exp: expect{
				setup: func() {
					mock.ExpectQuery(findByIDQ).WithArgs("bad").WillReturnError(errors.New("db down"))
				},
				err: true, nil: true,
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tc.exp.setup()

			switch tc.op {
			case opCreate:
				err := repo.Create(context.Background(), tc.arg.user)
				if tc.exp.err {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
					assert.NotEmpty(t, tc.arg.user.ID)
				}
			case opFindByEmail:
				got, err := repo.FindByEmail(context.Background(), tc.arg.email)
				if tc.exp.err {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
					if tc.exp.nil {
						assert.Nil(t, got)
					} else {
						require.NotNil(t, got)
						assert.Equal(t, tc.arg.email, got.Email)
					}
				}
			case opFindByID:
				got, err := repo.FindByID(context.Background(), tc.arg.id)
				if tc.exp.err {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
					if tc.exp.nil {
						assert.Nil(t, got)
					} else {
						require.NotNil(t, got)
						assert.Equal(t, tc.arg.id, got.ID)
					}
				}
			default:
				t.Fatalf("unknown op: %s", tc.op)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
