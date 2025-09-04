package http

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/guttosm/user-service/internal/domain/model"
	amocks "github.com/guttosm/user-service/internal/mocks/auth"
)

func TestHandler_Register(t *testing.T) {
	gin.SetMode(gin.TestMode)

	makeRouter := func(svc *amocks.MockAuthService) *gin.Engine {
		h := NewHandler(svc)
		r := gin.New()
		r.POST("/api/register", h.Register)
		return r
	}

	tests := []struct {
		name       string
		body       string
		configure  func(m *amocks.MockAuthService)
		wantStatus int
		wantSubstr string
	}{
		{
			name:       "invalid json → 400",
			body:       `{"email":1}`,
			wantStatus: http.StatusBadRequest,
			wantSubstr: "Invalid request",
		},
		{
			name: "service error → 400",
			body: `{"email":"x@y.com","password":"secret12","role":"user"}`,
			configure: func(m *amocks.MockAuthService) {
				m.On("Register", mock.Anything, "x@y.com", "secret12", "user").Return(nil, errors.New("boom"))
			},
			wantStatus: http.StatusBadRequest,
			wantSubstr: "Registration failed",
		},
		{
			name: "success → 201",
			body: `{"email":"a@b.com","password":"strongpw","role":"admin"}`,
			configure: func(m *amocks.MockAuthService) {
				m.On("Register", mock.Anything, "a@b.com", "strongpw", "admin").Return(&model.User{ID: "u-1", Email: "a@b.com", Role: "admin"}, nil)
			},
			wantStatus: http.StatusCreated,
			wantSubstr: "\"id\":\"u-1\"",
		},
	}

	for _, tc := range tests {
		c := tc
		t.Run(c.name, func(t *testing.T) {
			mockSvc := amocks.NewMockAuthService(t)
			if c.configure != nil {
				c.configure(mockSvc)
			}
			r := makeRouter(mockSvc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBufferString(c.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
			require.Equal(t, c.wantStatus, w.Code)
			assert.Contains(t, w.Body.String(), c.wantSubstr)
		})
	}
}

func TestHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)
	makeRouter := func(svc *amocks.MockAuthService) *gin.Engine {
		h := NewHandler(svc)
		r := gin.New()
		r.POST("/api/login", h.Login)
		return r
	}

	tests := []struct {
		name       string
		body       string
		configure  func(m *amocks.MockAuthService)
		wantStatus int
		wantSubstr string
	}{
		{
			name:       "invalid json → 400",
			body:       `{"email":5}`,
			wantStatus: http.StatusBadRequest,
			wantSubstr: "Invalid request",
		},
		{
			name: "invalid creds → 401",
			body: `{"email":"x@y.com","password":"pw"}`,
			configure: func(m *amocks.MockAuthService) {
				m.On("Login", mock.Anything, "x@y.com", "pw").Return("", errors.New("bad"))
			},
			wantStatus: http.StatusUnauthorized,
			wantSubstr: "Invalid credentials",
		},
		{
			name: "success → 200",
			body: `{"email":"user@ex.com","password":"pw"}`,
			configure: func(m *amocks.MockAuthService) {
				m.On("Login", mock.Anything, "user@ex.com", "pw").Return("tok123", nil)
			},
			wantStatus: http.StatusOK,
			wantSubstr: "tok123",
		},
	}

	for _, tc := range tests {
		c := tc
		t.Run(c.name, func(t *testing.T) {
			mockSvc := amocks.NewMockAuthService(t)
			if c.configure != nil {
				c.configure(mockSvc)
			}
			r := makeRouter(mockSvc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(c.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
			require.Equal(t, c.wantStatus, w.Code)
			assert.Contains(t, w.Body.String(), c.wantSubstr)
		})
	}
}
