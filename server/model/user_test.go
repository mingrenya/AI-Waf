package model

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword_NotEmpty(t *testing.T) {
	u := &User{Username: "admin", Password: "secret123"}
	err := u.HashPassword()
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if u.Password == "secret123" {
		t.Error("password should be hashed, not stored as plaintext")
	}
	if len(u.Password) < 30 {
		t.Errorf("hashed password too short: %d chars", len(u.Password))
	}
}

func TestHashPassword_Verifiable(t *testing.T) {
	plain := "mypassword"
	u := &User{Password: plain}
	u.HashPassword()

	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(plain))
	if err != nil {
		t.Errorf("hash does not match original password: %v", err)
	}
}

func TestHashPassword_Uniqueness(t *testing.T) {
	u1 := &User{Password: "samepass"}
	u2 := &User{Password: "samepass"}
	u1.HashPassword()
	u2.HashPassword()
	if u1.Password == u2.Password {
		t.Error("same password should produce different hashes (salt)")
	}
}
