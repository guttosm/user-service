package http

import (
	"github.com/gin-gonic/gin"
	_ "github.com/guttosm/user-service/docs"
	"github.com/guttosm/user-service/internal/middleware"
	"github.com/guttosm/user-service/internal/util/jwtutil"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// NewRouter sets up the HTTP routes for the user service.
//
// Parameters:
// - handler (*Handler): The HTTP handler containing the logic for authentication.
// - validator (jwtutil.TokenService): The service used to validate JWT tokens.
//
// Returns:
// - *gin.Engine: The configured Gin router with public and protected endpoints.
func NewRouter(handler *Handler, validator jwtutil.TokenService) *gin.Engine {
	router := gin.Default()

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	public := router.Group("/api")
	{
		public.POST("/register", handler.Register)
		public.POST("/login", handler.Login)
	}

	protected := router.Group("/api")
	protected.Use(middleware.AuthMiddleware(validator))

	return router
}
