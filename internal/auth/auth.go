package auth

import (
	"context"
	"errors"
	"os"
	"strconv"

	"github.com/guttosm/user-service/internal/domain/model"
	"github.com/guttosm/user-service/internal/repository/mongo"
	repository "github.com/guttosm/user-service/internal/repository/postgres"
	jwt "github.com/guttosm/user-service/internal/util/jwtutil"
	"golang.org/x/crypto/bcrypt"
)

// Service defines the interface for authentication-related business operations.
//
// Methods:
//   - Register: creates a new user with a hashed password.
//   - Login: verifies credentials and returns a signed JWT token.
type Service interface {
	Register(ctx context.Context, email, password, role string) (*model.User, error)
	Login(ctx context.Context, email, password string) (string, error)
}

// TokenValidator defines the interface for validating JWT tokens.
//
// Methods:
//   - ValidateToken: parses and validates the given token string.
type TokenValidator interface {
	ValidateToken(token string) (map[string]interface{}, error)
}

// AuthService implements the Service interface and handles user authentication logic.
// It coordinates user repository access, password hashing, token generation,
// and audit logging of authentication events.
type AuthService struct {
	repo   repository.UserRepository
	logger *mongo.Logger
	tokens jwt.TokenService
}

// NewAuthService returns a new instance of AuthService.
//
// Parameters:
//   - repo: abstraction for user data storage.
//   - tokenService: implementation of TokenService for generating JWTs.
//   - logger: handles event logging to MongoDB.
//
// Returns:
//   - *AuthService: a configured instance of the authentication service.
func NewAuthService(repo repository.UserRepository, token jwt.TokenService, logger *mongo.Logger) *AuthService {
	return &AuthService{
		repo:   repo,
		logger: logger,
		tokens: token,
	}
}

// Register creates a new user, hashing the password before storage, and logs the registration event.
//
// Parameters:
//   - email: unique identifier for the user.
//   - password: plain-text password to be hashed.
//   - role: the access level or role assigned to the user.
//
// Returns:
//   - *model.User: the persisted user with ID.
//   - error: if creation fails or email already exists.
func (s *AuthService) Register(ctx context.Context, email, password, role string) (*model.User, error) {
	existing, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("user already exists")
	}

	cost := bcrypt.DefaultCost
	if v := os.Getenv("BCRYPT_COST"); v != "" {
		if parsed, perr := strconv.Atoi(v); perr == nil && parsed >= bcrypt.MinCost && parsed <= bcrypt.MaxCost {
			cost = parsed
		}
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Email:    email,
		Password: string(hashedPassword),
		Role:     role,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	_ = s.logger.Log(ctx, mongo.AuthLog{
		EventType: "register",
		UserID:    user.ID,
	})

	return user, nil
}

// Login validates a user's credentials and returns a signed JWT token on success.
//
// Parameters:
//   - email: the user email.
//   - password: plain-text password for authentication.
//
// Returns:
//   - string: signed JWT token if successful.
//   - error: if authentication fails or token generation fails.
func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	token, err := s.tokens.Generate(user.ID, user.Role)
	if err != nil {
		return "", err
	}

	_ = s.logger.Log(ctx, mongo.AuthLog{
		EventType: "login",
		UserID:    user.ID,
	})

	return token, nil
}

// ValidateToken parses and validates the JWT using the configured TokenService.
//
// Parameters:
//   - token: the raw JWT string to be validated.
//
// Returns:
//   - map[string]interface{}: token claims if valid.
//   - error: validation error or parsing failure.
func (s *AuthService) ValidateToken(token string) (map[string]interface{}, error) {
	return s.tokens.Validate(token)
}
