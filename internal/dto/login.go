package dto

// LoginRequest represents the payload to authenticate a user.
//
// Fields:
//   - Email: the registered user email.
//   - Password: the plain-text password.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"strongpassword"`
}

// LoginResponse represents the response containing a JWT token after successful login.
//
// Fields:
//   - Token: the signed JWT token.
type LoginResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}
