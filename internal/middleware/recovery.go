package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/guttosm/user-service/internal/domain/dto"
)

// RecoveryMiddleware returns a Gin middleware that gracefully recovers from any panics,
// logs the stack trace for debugging, and returns a standardized JSON error response.
//
// Behavior:
//   - Uses defer to catch any panic that occurs during request handling.
//   - Prints the recovered panic value and stack trace to stdout (can be adapted to structured logging).
//   - Returns a 500 Internal Server Error response using dto.NewErrorResponse.
//
// Returns:
//   - gin.HandlerFunc: A middleware function for use in Gin router.
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// Log the panic and stack trace (you could replace this with a logger if needed)
				fmt.Printf("[PANIC RECOVERED] %v\n%s\n", r, debug.Stack())

				// Respond with standardized error structure
				errResponse := dto.NewErrorResponse("Internal server error", fmt.Errorf("%v", r))
				c.AbortWithStatusJSON(http.StatusInternalServerError, errResponse)
			}
		}()

		c.Next()
	}
}
