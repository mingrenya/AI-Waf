package service

import (
	"context"

	"github.com/mingrenya/AI-Waf/server/model"
	"github.com/mingrenya/AI-Waf/server/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// --- mock repos ---

type mockUserRepo struct {
	users map[string]*model.User // username -> user
}

func (m *mockUserRepo) FindByID(ctx context.Context, id bson.ObjectID) (*model.User, error) {
	return nil, nil
}
func (m *mockUserRepo) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	if u, ok := m.users[username]; ok {
		return u, nil
	}
	return nil, nil
}
func (m *mockUserRepo) Create(ctx context.Context, user *model.User) error {
	m.users[user.Username] = user
	return nil
}
func (m *mockUserRepo) Update(ctx context.Context, user *model.User) error { return nil }
func (m *mockUserRepo) Delete(ctx context.Context, id bson.ObjectID) error  { return nil }
func (m *mockUserRepo) UpdateLastLogin(ctx context.Context, id bson.ObjectID) error { return nil }
func (m *mockUserRepo) FindAll(ctx context.Context) ([]*model.User, error) {
	var users []*model.User
	for _, u := range m.users {
		users = append(users, u)
	}
	return users, nil
}
func (m *mockUserRepo) InitAdminUser() error { return nil }

type mockRoleRepo struct{}

func (m *mockRoleRepo) FindByID(ctx context.Context, id string) (*model.Role, error) {
	return nil, nil
}
func (m *mockRoleRepo) FindByName(ctx context.Context, name string) (*model.Role, error) {
	return nil, nil
}
func (m *mockRoleRepo) Create(ctx context.Context, role *model.Role) error { return nil }
func (m *mockRoleRepo) Update(ctx context.Context, role *model.Role) error { return nil }
func (m *mockRoleRepo) Delete(ctx context.Context, id string) error        { return nil }
func (m *mockRoleRepo) FindAll(ctx context.Context) ([]*model.Role, error) {
	return nil, nil
}
func (m *mockRoleRepo) InitDefaultRoles() error { return nil }

// Ensure mocks implement interfaces
var _ repository.UserRepository = (*mockUserRepo)(nil)
var _ repository.RoleRepository = (*mockRoleRepo)(nil)
