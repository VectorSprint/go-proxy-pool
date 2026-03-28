package httpcloak_test

import (
	"testing"

	"github.com/VectorSprint/go-proxy-pool/pkg/decodo"
	"github.com/VectorSprint/go-proxy-pool/pkg/decodo/adapter/httpcloak"
)

func TestProxyStringFromConfig(t *testing.T) {
	cfg := decodo.Config{
		Auth: decodo.Auth{
			Username: "username",
			Password: "password",
		},
		Targeting: decodo.Targeting{
			Country: "us",
		},
	}

	proxy, err := httpcloak.ProxyString(cfg)
	if err != nil {
		t.Fatalf("ProxyString() error = %v", err)
	}

	want := "http://user-username-country-us:password@gate.decodo.com:7000"
	if proxy != want {
		t.Fatalf("proxy = %q, want %q", proxy, want)
	}
}

func TestProxyStringFromLease(t *testing.T) {
	lease := decodo.Lease{
		ProxyURL: "http://user-username-country-us-session-session-1-sessionduration-30:password@gate.decodo.com:7000",
	}

	proxy := httpcloak.ProxyStringFromLease(lease)
	if proxy != lease.ProxyURL {
		t.Fatalf("proxy = %q, want %q", proxy, lease.ProxyURL)
	}
}

func TestProxyStringSOCKS5(t *testing.T) {
	cfg := decodo.Config{
		Auth: decodo.Auth{
			Username: "username",
			Password: "password",
		},
		Targeting: decodo.Targeting{
			Country: "us",
		},
		Session: decodo.Session{
			Type:            decodo.SessionTypeSticky,
			ID:              "session-1",
			DurationMinutes: 30,
		},
	}

	proxy, err := httpcloak.ProxyStringSOCKS5(cfg)
	if err != nil {
		t.Fatalf("ProxyStringSOCKS5() error = %v", err)
	}

	// SOCKS5 uses gate.decodo.com:7000 with socks5h scheme
	want := "socks5h://user-username-country-us-session-session-1-sessionduration-30:password@gate.decodo.com:7000"
	if proxy != want {
		t.Fatalf("proxy = %q, want %q", proxy, want)
	}
}

func TestProxyStringSOCKS5WithDedicatedEndpoint(t *testing.T) {
	spec, _ := decodo.NewEndpointSpec("us.decodo.com", 10000, decodo.PortRange{
		Start: 10001,
		End:   29999,
	})

	cfg := decodo.Config{
		Auth: decodo.Auth{
			Username: "username",
			Password: "password",
		},
		EndpointSpec: spec,
		Session: decodo.Session{
			Type:            decodo.SessionTypeSticky,
			ID:              "session-1",
			DurationMinutes: 30,
		},
	}

	proxy, err := httpcloak.ProxyStringSOCKS5(cfg)
	if err != nil {
		t.Fatalf("ProxyStringSOCKS5() error = %v", err)
	}

	// SOCKS5 ignores dedicated endpoint, always uses gate.decodo.com:7000
	want := "socks5h://user-username-session-session-1-sessionduration-30:password@gate.decodo.com:7000"
	if proxy != want {
		t.Fatalf("proxy = %q, want %q", proxy, want)
	}
}

func TestProxyStringSOCKS5FromLease(t *testing.T) {
	lease := decodo.Lease{
		ProxyURL: "http://user-username-country-us-session-session-1-sessionduration-30:password@us.decodo.com:10001",
	}

	proxy, err := httpcloak.ProxyStringSOCKS5FromLease(lease)
	if err != nil {
		t.Fatalf("ProxyStringSOCKS5FromLease() error = %v", err)
	}

	// Should convert HTTP proxy URL to SOCKS5 with gate.decodo.com:7000
	want := "socks5h://user-username-country-us-session-session-1-sessionduration-30:password@gate.decodo.com:7000"
	if proxy != want {
		t.Fatalf("proxy = %q, want %q", proxy, want)
	}
}
