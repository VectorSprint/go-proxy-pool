package decodo_test

import (
	"testing"

	"github.com/VectorSprint/go-proxy-pool/pkg/decodo"
)

func TestNewAuthNormalizesCredentials(t *testing.T) {
	auth, err := decodo.NewAuth("  username  ", "  password  ")
	if err != nil {
		t.Fatalf("NewAuth() error = %v", err)
	}

	if auth.Username != "username" {
		t.Fatalf("username = %q, want %q", auth.Username, "username")
	}

	if auth.Password != "password" {
		t.Fatalf("password = %q, want %q", auth.Password, "password")
	}
}

func TestNewAuthRejectsPrefixedUsername(t *testing.T) {
	_, err := decodo.NewAuth("user-username", "password")
	if err == nil {
		t.Fatal("NewAuth() error = nil, want error")
	}
}

func TestAuthValidateRejectsMissingPassword(t *testing.T) {
	auth := decodo.Auth{Username: "username"}
	if err := auth.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want error")
	}
}
