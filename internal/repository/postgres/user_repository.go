package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/guttosm/user-service/internal/domain/model"
)

// UserRepository defines the contract for user-related database operations.
//
// Implementations of this interface are responsible for creating,
// retrieving, and managing user records in the persistence layer.
type UserRepository interface {

	// Create inserts a new user into the database.
	//
	// Parameters:
	//   - user (*model.User): The user object to be created.
	//                         Its ID should be set after successful insertion.
	//
	// Returns:
	//   - error: Any error encountered during the insert operation.
	Create(ctx context.Context, user *model.User) error

	// FindByEmail retrieves a user based on their email address.
	//
	// Parameters:
	//   - email (string): The email address of the user to retrieve.
	//
	// Returns:
	//   - *model.User: The found user, or nil if not found.
	//   - error: Any error encountered during the query.
	FindByEmail(ctx context.Context, email string) (*model.User, error)

	// FindByID retrieves a user based on their unique identifier.
	//
	// Parameters:
	//   - id (string): The UUID of the user to retrieve.
	//
	// Returns:
	//   - *model.User: The found user, or nil if not found.
	//   - error: Any error encountered during the query.
	FindByID(ctx context.Context, id string) (*model.User, error)
}

// PostgresUserRepository provides PostgreSQL-based implementation for user persistence.
type PostgresUserRepository struct {
	db *sql.DB
}

// NewUserRepository initializes a new instance of PostgresUserRepository.
//
// Parameters:
//   - db: An open PostgreSQL connection.
//
// Returns:
//   - *PostgresUserRepository: A struct that implements user-related persistence operations.
func NewUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

// Create inserts a new user into the "users" table with auto-generated UUID v7.
//
// Parameters:
//   - user (*model.User): The user object to be stored. ID is populated after insertion.
//
// Returns:
//   - error: Any error that occurred during the insert operation.
func (r *PostgresUserRepository) Create(ctx context.Context, user *model.User) error {
	normalized := strings.ToLower(user.Email)
	query := `INSERT INTO users (email, password, role) VALUES ($1, $2, $3) RETURNING id;`
	if err := r.db.QueryRowContext(ctx, query, normalized, user.Password, user.Role).Scan(&user.ID); err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	user.Email = normalized
	return nil
}

// FindByEmail returns a user by their email address.
//
// Parameters:
//   - email (string): The email to search for.
//
// Returns:
//   - *model.User: The user found, or nil if not found.
//   - error: Any error encountered during the query.
func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `SELECT id, email, password, role FROM users WHERE email = $1`
	row := r.db.QueryRowContext(ctx, query, strings.ToLower(email))
	var user model.User
	if err := row.Scan(&user.ID, &user.Email, &user.Password, &user.Role); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find by email: %w", err)
	}
	return &user, nil
}

// FindByID retrieves a user from the database using their UUID.
//
// Parameters:
//   - id (string): The user ID to search for.
//
// Returns:
//   - *model.User: The user found, or nil if not found.
//   - error: Any error encountered during the query.
func (r *PostgresUserRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	query := `SELECT id, email, password, role FROM users WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)
	var user model.User
	if err := row.Scan(&user.ID, &user.Email, &user.Password, &user.Role); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find by id: %w", err)
	}
	return &user, nil
}
