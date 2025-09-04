package http

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/guttosm/user-service/internal/domain/model"
	amocks "github.com/guttosm/user-service/internal/mocks/auth"
	"github.com/guttosm/user-service/internal/util/jwtutil"
)

// stubTokenService implements jwtutil.TokenService minimally for router construction.
type stubTokenService struct{}

func (s stubTokenService) Generate(userID, role string) (string, error) { return "stub", nil }
func (s stubTokenService) Validate(tok string) (map[string]interface{}, error) {
	return map[string]interface{}{"user_id": "stub-user", "role": "user"}, nil
}

func TestNewRouter_CoreRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockAuth := amocks.NewMockAuthService(t)

	h := NewHandler(mockAuth)
	r := NewRouter(h, stubTokenService{})

	// healthz
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "ok")

	// register
	mockAuth.On("Register", mock.Anything, "r@test.io", "strongpw", "user").Return(&model.User{ID: "id-1", Email: "r@test.io", Role: "user"}, nil)
	regBody := bytes.NewBufferString(`{"email":"r@test.io","password":"strongpw","role":"user"}`)
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/api/register", regBody)
	req2.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w2, req2)
	require.Equal(t, http.StatusCreated, w2.Code)
	assert.Contains(t, w2.Body.String(), "\"id\":\"id-1\"")

	// login
	mockAuth.On("Login", mock.Anything, "r@test.io", "strongpw").Return("token-xyz", nil)
	loginBody := bytes.NewBufferString(`{"email":"r@test.io","password":"strongpw"}`)
	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodPost, "/api/login", loginBody)
	req3.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w3, req3)
	require.Equal(t, http.StatusOK, w3.Code)
	assert.Contains(t, w3.Body.String(), "token-xyz")
}

// compile-time assertion that stub implements interface
var _ jwtutil.TokenService = (*stubTokenService)(nil)
