package decodo_test

import (
	"testing"

	"github.com/VectorSprint/go-proxy-pool/pkg/decodo"
)

func TestNewEndpointSpec(t *testing.T) {
	spec, err := decodo.NewEndpointSpec("ca.decodo.com", 20000, decodo.PortRange{
		Start: 20001,
		End:   29999,
	})
	if err != nil {
		t.Fatalf("NewEndpointSpec() error = %v", err)
	}

	if spec.Host != "ca.decodo.com" {
		t.Fatalf("host = %q, want %q", spec.Host, "ca.decodo.com")
	}

	if spec.RotatingPort != 20000 {
		t.Fatalf("rotating port = %d, want %d", spec.RotatingPort, 20000)
	}
}

func TestConfigUsesDedicatedRotatingEndpointPort(t *testing.T) {
	spec, err := decodo.NewEndpointSpec("ca.decodo.com", 20000, decodo.PortRange{
		Start: 20001,
		End:   29999,
	})
	if err != nil {
		t.Fatalf("NewEndpointSpec() error = %v", err)
	}

	cfg := decodo.Config{
		Auth: decodo.Auth{
			Username: "username",
			Password: "password",
		},
		EndpointSpec: spec,
		Targeting: decodo.Targeting{
			Country: "ca",
		},
	}

	proxyURL, err := cfg.ProxyURL()
	if err != nil {
		t.Fatalf("ProxyURL() error = %v", err)
	}

	if got := proxyURL.Host; got != "ca.decodo.com:20000" {
		t.Fatalf("host = %q, want %q", got, "ca.decodo.com:20000")
	}
}

func TestConfigUsesStickyPortFromDedicatedRange(t *testing.T) {
	spec, err := decodo.NewEndpointSpec("ca.decodo.com", 20000, decodo.PortRange{
		Start: 20001,
		End:   29999,
	})
	if err != nil {
		t.Fatalf("NewEndpointSpec() error = %v", err)
	}

	cfg := decodo.Config{
		Auth: decodo.Auth{
			Username: "username",
			Password: "password",
		},
		EndpointSpec: spec,
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

	if got := proxyURL.Host; got != "ca.decodo.com:20001" {
		t.Fatalf("host = %q, want %q", got, "ca.decodo.com:20001")
	}
}

func TestConfigValidationRejectsStickyPortOutsideRange(t *testing.T) {
	spec, err := decodo.NewEndpointSpec("ca.decodo.com", 20000, decodo.PortRange{
		Start: 20001,
		End:   29999,
	})
	if err != nil {
		t.Fatalf("NewEndpointSpec() error = %v", err)
	}

	cfg := decodo.Config{
		Auth: decodo.Auth{
			Username: "username",
			Password: "password",
		},
		EndpointSpec: spec,
		Port:         20000,
		Session: decodo.Session{
			Type:            decodo.SessionTypeSticky,
			ID:              "task-1",
			DurationMinutes: 30,
		},
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want error")
	}
}
