package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/guttosm/user-service/internal/auth"
	"github.com/guttosm/user-service/internal/dto"
	"github.com/guttosm/user-service/internal/middleware"
)

// Handler provides HTTP handlers for user authentication routes.
type Handler struct {
	authService auth.Service
}

// NewHandler returns a new Handler instance.
//
// Parameters:
// - authService: business logic for user registration and login.
func NewHandler(authService auth.Service) *Handler {
	return &Handler{
		authService: authService,
	}
}

// Register handles user registration requests.
//
// @Summary Register new user
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "User info"
// @Success 201 {object} dto.RegisterResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	user, err := h.authService.Register(req.Email, req.Password, req.Role)
	if err != nil {
		middleware.AbortWithError(c, http.StatusBadRequest, "Registration failed", err)
		return
	}

	resp := dto.RegisterResponse{
		ID:    user.ID,
		Email: user.Email,
		Role:  user.Role,
	}
	c.JSON(http.StatusCreated, resp)
}

// Login handles user login requests.
//
// @Summary Login user
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "User credentials"
// @Success 200 {object} dto.LoginResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	token, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		middleware.AbortWithError(c, http.StatusUnauthorized, "Invalid credentials", err)
		return
	}

	resp := dto.LoginResponse{
		Token: token,
	}
	c.JSON(http.StatusOK, resp)
}
