package model

// User represents a system user entity, typically persisted in the database.
//
// Fields:
//   - ID: Unique identifier for the user (usually UUID).
//   - Email: User's email address, used as login credential.
//   - Password: Hashed password (never store plain-text passwords).
//   - Role: Role or access level assigned to the user (e.g., "admin", "user").
type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}
