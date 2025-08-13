package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	jwt "github.com/guttosm/user-service/internal/util/jwtutil"
)

// AuthMiddleware returns a Gin middleware that validates JWT Bearer tokens,
// extracts the "user_id" claim, and injects it into the request context.
//
// Expected header:
//
//	Authorization: Bearer <token>
//
// Behavior:
//   - If the Authorization header is missing or malformed, it aborts with 401.
//   - If the token is invalid or expired, it aborts with 401.
//   - If the "user_id" claim is missing or invalid, it aborts with 401.
//   - On success, it sets "user_id" in the context for downstream handlers.
//
// Parameters:
//   - svc: the auth.Service implementation used to validate the token.
func AuthMiddleware(svc jwt.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		const bearerPrefix = "Bearer "
		authHeader := c.GetHeader("Authorization")

		if !strings.HasPrefix(authHeader, bearerPrefix) {
			AbortWithError(c, http.StatusUnauthorized, "Missing or malformed Authorization header", nil)
			return
		}

		token := strings.TrimPrefix(authHeader, bearerPrefix)

		claims, err := svc.Validate(token)
		if err != nil {
			AbortWithError(c, http.StatusUnauthorized, "Invalid or expired token", err)
			return
		}

		rawUserID, ok := claims["user_id"]
		if !ok {
			AbortWithError(c, http.StatusUnauthorized, "Missing 'user_id' claim in token", nil)
			return
		}

		userID, ok := rawUserID.(string)
		if !ok || strings.TrimSpace(userID) == "" {
			AbortWithError(c, http.StatusUnauthorized, "'user_id' claim must be a non-empty string", nil)
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}
