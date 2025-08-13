package app

import (
	"github.com/gin-gonic/gin"
	"github.com/guttosm/user-service/config"
	"github.com/guttosm/user-service/internal/auth"
	"github.com/guttosm/user-service/internal/http"
	"github.com/guttosm/user-service/internal/repository/mongo"
	repository "github.com/guttosm/user-service/internal/repository/postgres"
	"github.com/guttosm/user-service/internal/util/jwtutil"
)

// InitializeApp sets up all dependencies of the application and returns
// the Gin router, a cleanup function for graceful shutdown, and any error encountered.
//
// Returns:
//   - *gin.Engine: the HTTP router
//   - func(): cleanup function to release DB connections
//   - error: any initialization error
func InitializeApp() (*gin.Engine, func(), error) {
	cfg := config.AppConfig

	pgDB, err := InitPostgres(cfg)
	if err != nil {
		return nil, nil, err
	}

	_, mongoDB, mongoCleanup, err := InitMongo(cfg)
	if err != nil {
		_ = pgDB.Close()
		return nil, nil, err
	}

	eventLogger := mongo.NewLogger(mongoDB, "auth_logs")

	userRepo := repository.NewUserRepository(pgDB)
	tokenService := jwtutil.NewJWTService(cfg.JWT)
	authService := auth.NewAuthService(userRepo, tokenService, eventLogger)

	handler := http.NewHandler(authService)
	router := http.NewRouter(handler, tokenService)

	// Cleanup resources on shutdown
	cleanup := func() {
		_ = pgDB.Close()
		mongoCleanup()
	}

	return router, cleanup, nil
}
