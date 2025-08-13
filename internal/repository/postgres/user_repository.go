package repository

import (
	"database/sql"
	"errors"

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
	Create(user *model.User) error

	// FindByEmail retrieves a user based on their email address.
	//
	// Parameters:
	//   - email (string): The email address of the user to retrieve.
	//
	// Returns:
	//   - *model.User: The found user, or nil if not found.
	//   - error: Any error encountered during the query.
	FindByEmail(email string) (*model.User, error)

	// FindByID retrieves a user based on their unique identifier.
	//
	// Parameters:
	//   - id (string): The UUID of the user to retrieve.
	//
	// Returns:
	//   - *model.User: The found user, or nil if not found.
	//   - error: Any error encountered during the query.
	FindByID(id string) (*model.User, error)
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
func (r *PostgresUserRepository) Create(user *model.User) error {
	query := `
		INSERT INTO users (email, password, role)
		VALUES ($1, $2, $3)
		RETURNING id;
	`
	return r.db.QueryRow(query, user.Email, user.Password, user.Role).Scan(&user.ID)
}

// FindByEmail returns a user by their email address.
//
// Parameters:
//   - email (string): The email to search for.
//
// Returns:
//   - *model.User: The user found, or nil if not found.
//   - error: Any error encountered during the query.
func (r *PostgresUserRepository) FindByEmail(email string) (*model.User, error) {
	query := `SELECT id, email, password, role FROM users WHERE email = $1`

	row := r.db.QueryRow(query, email)

	var user model.User
	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.Role)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
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
func (r *PostgresUserRepository) FindByID(id string) (*model.User, error) {
	query := `SELECT id, email, password, role FROM users WHERE id = $1`

	row := r.db.QueryRow(query, id)

	var user model.User
	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.Role)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}
