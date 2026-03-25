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
