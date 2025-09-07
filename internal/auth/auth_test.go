package auth

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/guttosm/user-service/internal/domain/model"
	mlog "github.com/guttosm/user-service/internal/repository/mongo"
	driverMongo "go.mongodb.org/mongo-driver/mongo"
	driverOptions "go.mongodb.org/mongo-driver/mongo/options"
)

type memRepo struct {
	mu      sync.Mutex
	byID    map[string]*model.User
	byEmail map[string]*model.User
}

func newMemRepo() *memRepo {
	return &memRepo{byID: map[string]*model.User{}, byEmail: map[string]*model.User{}}
}

func (m *memRepo) Create(ctx context.Context, u *model.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.byEmail[u.Email]; ok {
		return errors.New("duplicate")
	}
	if u.ID == "" {
		u.ID = "id-1"
	}
	m.byEmail[u.Email] = &model.User{ID: u.ID, Email: u.Email, Password: u.Password, Role: u.Role}
	m.byID[u.ID] = m.byEmail[u.Email]
	return nil
}
func (m *memRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if u, ok := m.byEmail[email]; ok {
		cp := *u
		return &cp, nil
	}
	return nil, nil
}
func (m *memRepo) FindByID(ctx context.Context, id string) (*model.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if u, ok := m.byID[id]; ok {
		cp := *u
		return &cp, nil
	}
	return nil, nil
}

type tokenStub struct{}

func (tokenStub) Generate(userID, role string) (string, error) { return "tok", nil }
func (tokenStub) Validate(token string) (map[string]interface{}, error) {
	return map[string]interface{}{"ok": true}, nil
}

func TestAuthService_RegisterAndLogin(t *testing.T) {
	repo := newMemRepo()
	// Use test logger with a stub inserter to avoid real Mongo usage
	service := NewAuthService(repo, tokenStub{}, mlog.NewTestLogger(stubInserter{}))

	ctx := context.Background()
	u, err := service.Register(ctx, "a@b.com", "pw", "user")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if u.ID == "" {
		t.Fatalf("expected id set")
	}

	// duplicate
	if _, err := service.Register(ctx, "a@b.com", "pw2", "user"); err == nil {
		t.Fatalf("expected duplicate error")
	}

	tok, err := service.Login(ctx, "a@b.com", "pw")
	if err != nil || tok == "" {
		t.Fatalf("login: %v tok=%q", err, tok)
	}

	tok, err = service.Login(ctx, "a@b.com", "bad")
	if err == nil || tok != "" {
		t.Fatalf("expected invalid credentials")
	}
}

// stubInserter implements mongoInserter for tests
type stubInserter struct{}

func (stubInserter) InsertOne(ctx context.Context, document interface{}, _ ...*driverOptions.InsertOneOptions) (*driverMongo.InsertOneResult, error) {
	return &driverMongo.InsertOneResult{}, nil
}
