// Code generated manually to mock auth.Service.
// Pattern similar to other mocks. Adjust if mockery tooling is adopted later.
package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/guttosm/user-service/internal/domain/model"
)

type MockAuthService struct{ mock.Mock }

func NewMockAuthService(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockAuthService {
	m := &MockAuthService{}
	m.Test(t)
	// ensure expectations asserted
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (m *MockAuthService) Register(ctx context.Context, email, password, role string) (*model.User, error) {
	args := m.Called(ctx, email, password, role)
	var user *model.User
	if u := args.Get(0); u != nil {
		user = u.(*model.User)
	}
	return user, args.Error(1)
}

func (m *MockAuthService) Login(ctx context.Context, email, password string) (string, error) {
	args := m.Called(ctx, email, password)
	return args.String(0), args.Error(1)
}
