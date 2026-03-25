package decodo_test

import (
	"testing"
	"time"

	"github.com/VectorSprint/go-proxy-pool/pkg/decodo"
)

func TestBuildRotatingProxyURL(t *testing.T) {
	cfg := decodo.Config{
		Auth: decodo.Auth{
			Username: "username",
			Password: "password",
		},
	}

	proxyURL, err := cfg.ProxyURL()
	if err != nil {
		t.Fatalf("ProxyURL() error = %v", err)
	}

	if got := proxyURL.Scheme; got != "http" {
		t.Fatalf("scheme = %q, want %q", got, "http")
	}

	if got := proxyURL.Host; got != "gate.decodo.com:7000" {
		t.Fatalf("host = %q, want %q", got, "gate.decodo.com:7000")
	}

	if got := proxyURL.User.Username(); got != "user-username" {
		t.Fatalf("username = %q, want %q", got, "user-username")
	}
}

func TestBuildStickyProxyURLWithTargeting(t *testing.T) {
	cfg := decodo.Config{
		Auth: decodo.Auth{
			Username: "username",
			Password: "password",
		},
		Targeting: decodo.Targeting{
			Country: "us",
			City:    "new_york",
		},
		Session: decodo.Session{
			Type:            decodo.SessionTypeSticky,
			ID:              "task-1",
			DurationMinutes: 30,
		},
	}

	proxyURL, err := cfg.ProxyURL()
	if err != nil {
		t.Fatalf("ProxyURL() error = %v", err)
	}

	if got := proxyURL.User.Username(); got != "user-username-country-us-city-new_york-session-task-1-sessionduration-30" {
		t.Fatalf("username = %q", got)
	}
}

func TestConfigValidationRejectsInvalidTargetingCombination(t *testing.T) {
	cfg := decodo.Config{
		Auth: decodo.Auth{
			Username: "username",
			Password: "password",
		},
		Targeting: decodo.Targeting{
			Country: "us",
			ASN:     20057,
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() error = nil, want error")
	}
}

func TestConfigValidationRejectsStickyWithoutSessionID(t *testing.T) {
	cfg := decodo.Config{
		Auth: decodo.Auth{
			Username: "username",
			Password: "password",
		},
		Session: decodo.Session{
			Type:            decodo.SessionTypeSticky,
			DurationMinutes: 10,
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() error = nil, want error")
	}
}

func TestConfigValidationRejectsStickyDurationOutOfRange(t *testing.T) {
	cfg := decodo.Config{
		Auth: decodo.Auth{
			Username: "username",
			Password: "password",
		},
		Session: decodo.Session{
			Type:            decodo.SessionTypeSticky,
			ID:              "task-1",
			DurationMinutes: 1441,
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() error = nil, want error")
	}
}

func TestConfigDefaults(t *testing.T) {
	cfg := decodo.Config{
		Auth: decodo.Auth{
			Username: "username",
			Password: "password",
		},
		Session: decodo.Session{
			Type: decodo.SessionTypeSticky,
			ID:   "task-1",
		},
	}

	normalized, err := cfg.Normalized()
	if err != nil {
		t.Fatalf("Normalized() error = %v", err)
	}

	if normalized.Endpoint != "gate.decodo.com" {
		t.Fatalf("endpoint = %q, want %q", normalized.Endpoint, "gate.decodo.com")
	}

	if normalized.Port != 7000 {
		t.Fatalf("port = %d, want %d", normalized.Port, 7000)
	}

	if normalized.Session.DurationMinutes != 10 {
		t.Fatalf("duration = %d, want %d", normalized.Session.DurationMinutes, 10)
	}
}

func TestSessionTTL(t *testing.T) {
	cfg := decodo.Config{
		Auth: decodo.Auth{
			Username: "username",
			Password: "password",
		},
		Session: decodo.Session{
			Type:            decodo.SessionTypeSticky,
			ID:              "task-1",
			DurationMinutes: 30,
		},
	}

	normalized, err := cfg.Normalized()
	if err != nil {
		t.Fatalf("Normalized() error = %v", err)
	}

	if normalized.Session.TTL() != 30*time.Minute {
		t.Fatalf("TTL() = %s, want %s", normalized.Session.TTL(), 30*time.Minute)
	}
}
