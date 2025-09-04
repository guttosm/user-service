package jwtutil

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/guttosm/user-service/config"
)

var (
	ErrInvalidToken = errors.New("invalid or expired token")
)

// TokenService defines operations for generating and validating JWTs.
//
// Methods:
//   - Generate(userID, role): generates a signed JWT for a user.
//   - Validate(tokenString): validates a JWT and returns its claims.
type TokenService interface {
	Generate(userID, role string) (string, error)
	Validate(tokenString string) (map[string]interface{}, error)
}

// JWTService provides a concrete implementation of TokenService using HMAC signing.
type JWTService struct {
	cfg config.JWTConfig
}

// NewJWTService creates a new JWTService with the given configuration.
//
// Parameters:
//   - cfg: contains the JWT secret, expiration, and issuer.
//
// Returns:
//   - *JWTService: a ready-to-use service for generating and validating JWTs.
func NewJWTService(cfg config.JWTConfig) *JWTService {
	return &JWTService{cfg: cfg}
}

// Generate creates a signed JWT token with user claims.
//
// Parameters:
//   - userID: the unique identifier of the user.
//   - role: the access level or role of the user.
//
// Returns:
//   - string: a signed JWT token.
//   - error: if token creation fails.
func (j *JWTService) Generate(userID, role string) (string, error) {
	return generateToken(userID, role, j.cfg)
}

// Validate parses and validates the JWT token string.
//
// Parameters:
//   - tokenString: the JWT token to validate.
//
// Returns:
//   - map[string]interface{}: claims extracted from the token.
//   - error: if the token is invalid or expired.
func (j *JWTService) Validate(tokenString string) (map[string]interface{}, error) {
	return validateToken(tokenString, j.cfg)
}

// generateToken generates a signed JWT using provided configuration and user data.
//
// Parameters:
//   - userID: the unique identifier of the user.
//   - role: user's access role.
//   - cfg: the JWT configuration to use for signing.
//
// Returns:
//   - string: signed JWT token.
//   - error: any signing error.
func generateToken(userID, role string, cfg config.JWTConfig) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(time.Duration(cfg.ExpirationHour) * time.Hour).Unix(),
		"iss":     cfg.Issuer,
		"aud":     cfg.Audience,
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

// validateToken validates a signed JWT token using the secret.
//
// Parameters:
//   - tokenString: the token to be validated.
//   - secret: HMAC secret key used to sign the token.
//
// Returns:
//   - map[string]interface{}: claims from the token.
//   - error: if validation fails.
func validateToken(tokenString string, cfg config.JWTConfig) (map[string]interface{}, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(cfg.Secret), nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}
	// Issuer check
	if cfg.Issuer != "" && claims["iss"] != cfg.Issuer {
		return nil, ErrInvalidToken
	}
	// Audience check
	if cfg.Audience != "" {
		switch aud := claims["aud"].(type) {
		case string:
			if aud != cfg.Audience {
				return nil, ErrInvalidToken
			}
		case []interface{}:
			match := false
			for _, v := range aud {
				if s, ok := v.(string); ok && s == cfg.Audience {
					match = true
					break
				}
			}
			if !match {
				return nil, ErrInvalidToken
			}
		}
	}
	// Exp is enforced by Parse if using MapClaims.Valid(); we added custom check earlier.
	return claims, nil
}
