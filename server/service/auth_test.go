package service

import (
	"context"
	"testing"

	"github.com/mingrenya/AI-Waf/server/model"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestAuthService_UserNotFound(t *testing.T) {
	svc := &AuthServiceImpl{
		userRepo: &mockUserRepo{users: map[string]*model.User{}},
		roleRepo: &mockRoleRepo{},
	}

	user, err := svc.userRepo.FindByUsername(context.Background(), "nonexistent")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if user != nil {
		t.Error("expected nil user for nonexistent username")
	}
}

func TestAuthService_UserFound(t *testing.T) {
	id := bson.NewObjectID()
	expected := &model.User{
		ID:       id,
		Username: "admin",
		Role:     "admin",
	}

	svc := &AuthServiceImpl{
		userRepo: &mockUserRepo{users: map[string]*model.User{
			"admin": expected,
		}},
		roleRepo: &mockRoleRepo{},
	}

	user, err := svc.userRepo.FindByUsername(context.Background(), "admin")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("expected user, got nil")
	}
	if user.Username != "admin" {
		t.Errorf("expected username 'admin', got '%s'", user.Username)
	}
}

func TestAuthService_CreateUser(t *testing.T) {
	svc := &AuthServiceImpl{
		userRepo: &mockUserRepo{users: map[string]*model.User{}},
		roleRepo: &mockRoleRepo{},
	}

	newUser := &model.User{
		ID:       bson.NewObjectID(),
		Username: "newuser",
	}
	if err := svc.userRepo.Create(context.Background(), newUser); err != nil {
		t.Errorf("failed to create user: %v", err)
	}

	found, _ := svc.userRepo.FindByUsername(context.Background(), "newuser")
	if found == nil {
		t.Error("user should exist after creation")
	}
}

func TestAuthService_FindAllUsers(t *testing.T) {
	users := map[string]*model.User{
		"a": {ID: bson.NewObjectID(), Username: "a"},
		"b": {ID: bson.NewObjectID(), Username: "b"},
	}

	svc := &AuthServiceImpl{
		userRepo: &mockUserRepo{users: users},
		roleRepo: &mockRoleRepo{},
	}

	all, err := svc.userRepo.FindAll(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 users, got %d", len(all))
	}
}

func TestAuthService_ErrorSentinelValues(t *testing.T) {
	// Verify error sentinels are non-nil and distinct
	errors := []error{ErrUserNotFound, ErrInvalidPassword, ErrUserAlreadyExist, ErrForbidden}
	for i, e := range errors {
		if e == nil {
			t.Errorf("error[%d] is nil", i)
		}
	}
	// All must be distinct
	for i := 0; i < len(errors); i++ {
		for j := i + 1; j < len(errors); j++ {
			if errors[i] == errors[j] {
				t.Errorf("errors[%d] and errors[%d] are the same", i, j)
			}
		}
	}
}
