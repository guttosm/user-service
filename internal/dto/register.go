package dto

// RegisterRequest represents the payload to create a new user account.
//
// Fields:
//   - Email: the email address of the new user (must be valid format).
//   - Password: plain-text password to be hashed (minimum 6 characters).
//   - Role: the role or permission level (e.g., "user", "admin").
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"strongpassword"`
	Role     string `json:"role" binding:"required" example:"user"`
}

// RegisterResponse represents the response after successful user registration.
//
// Fields:
//   - ID: the unique identifier of the new user.
//   - Email: the registered email address.
//   - Role: the assigned role of the user.
type RegisterResponse struct {
	ID    string `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Email string `json:"email" example:"user@example.com"`
	Role  string `json:"role" example:"user"`
}
