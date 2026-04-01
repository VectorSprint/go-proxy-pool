package nethttp_test

import (
	"net/http"
	"testing"

	"github.com/VectorSprint/go-proxy-pool/pkg/decodo"
	"github.com/VectorSprint/go-proxy-pool/pkg/decodo/adapter/nethttp"
)

func TestProxyURLFromConfig(t *testing.T) {
	cfg := decodo.Config{
		Auth: decodo.Auth{
			Username: "username",
			Password: "password",
		},
	}

	proxyURL, err := nethttp.ProxyURL(cfg)
	if err != nil {
		t.Fatalf("ProxyURL() error = %v", err)
	}

	if got := proxyURL.String(); got != "http://user-username:password@gate.decodo.com:7000" {
		t.Fatalf("proxy url = %q", got)
	}
}

func TestProxyFuncFromConfig(t *testing.T) {
	cfg := decodo.Config{
		Auth: decodo.Auth{
			Username: "username",
			Password: "password",
		},
	}

	proxyFunc, err := nethttp.ProxyFunc(cfg)
	if err != nil {
		t.Fatalf("ProxyFunc() error = %v", err)
	}

	req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}

	proxyURL, err := proxyFunc(req)
	if err != nil {
		t.Fatalf("proxy func error = %v", err)
	}

	if got := proxyURL.String(); got != "http://user-username:password@gate.decodo.com:7000" {
		t.Fatalf("proxy url = %q", got)
	}
}

func TestProxyFuncReturnsErrorForInvalidConfig(t *testing.T) {
	cfg := decodo.Config{
		Auth: decodo.Auth{
			Username: "", // invalid: empty username
			Password: "password",
		},
	}

	_, err := nethttp.ProxyFunc(cfg)
	if err == nil {
		t.Fatal("ProxyFunc() error = nil, want error for invalid config")
	}
}

func TestProxyURLFromLease(t *testing.T) {
	lease := decodo.Lease{
		ProxyURL: "http://user-username-session-session-1-sessionduration-30:password@gate.decodo.com:7000",
	}

	proxyURL, err := nethttp.ProxyURLFromLease(lease)
	if err != nil {
		t.Fatalf("ProxyURLFromLease() error = %v", err)
	}

	if got := proxyURL.String(); got != lease.ProxyURL {
		t.Fatalf("proxy url = %q, want %q", got, lease.ProxyURL)
	}
}

func TestProxyURLSOCKS5(t *testing.T) {
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

	proxyURL, err := nethttp.ProxyURLSOCKS5(cfg)
	if err != nil {
		t.Fatalf("ProxyURLSOCKS5() error = %v", err)
	}

	if proxyURL.Scheme != "socks5h" {
		t.Fatalf("scheme = %q, want %q", proxyURL.Scheme, "socks5h")
	}

	if proxyURL.Host != "gate.decodo.com:7000" {
		t.Fatalf("host = %q, want %q", proxyURL.Host, "gate.decodo.com:7000")
	}
}

func TestProxyURLSOCKS5WithDedicatedEndpoint(t *testing.T) {
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

	proxyURL, err := nethttp.ProxyURLSOCKS5(cfg)
	if err != nil {
		t.Fatalf("ProxyURLSOCKS5() error = %v", err)
	}

	// SOCKS5 ignores dedicated endpoint, always uses gate.decodo.com:7000
	if proxyURL.Host != "gate.decodo.com:7000" {
		t.Fatalf("host = %q, want %q", proxyURL.Host, "gate.decodo.com:7000")
	}
}

func TestProxyURLSOCKS5WithAppliedPreset(t *testing.T) {
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
			ID:              "session-1",
			DurationMinutes: 30,
		},
	}

	cfg.ApplyPreset()

	proxyURL, err := nethttp.ProxyURLSOCKS5(cfg)
	if err != nil {
		t.Fatalf("ProxyURLSOCKS5() error = %v", err)
	}

	if proxyURL.Host != "gate.decodo.com:7000" {
		t.Fatalf("host = %q, want %q", proxyURL.Host, "gate.decodo.com:7000")
	}

	if proxyURL.User.Username() != "user-username-country-us-city-new_york-session-session-1-sessionduration-30" {
		t.Fatalf("username = %q", proxyURL.User.Username())
	}
}

func TestProxyURLSOCKS5FromLease(t *testing.T) {
	lease := decodo.Lease{
		ProxyURL: "http://user-username-country-us-session-session-1-sessionduration-30:password@us.decodo.com:10001",
	}

	proxyURL, err := nethttp.ProxyURLSOCKS5FromLease(lease)
	if err != nil {
		t.Fatalf("ProxyURLSOCKS5FromLease() error = %v", err)
	}

	if proxyURL.Scheme != "socks5h" {
		t.Fatalf("scheme = %q, want %q", proxyURL.Scheme, "socks5h")
	}

	if proxyURL.Host != "gate.decodo.com:7000" {
		t.Fatalf("host = %q, want %q", proxyURL.Host, "gate.decodo.com:7000")
	}

	if proxyURL.User.Username() != "user-username-country-us-session-session-1-sessionduration-30" {
		t.Fatalf("username = %q", proxyURL.User.Username())
	}
}

func TestProxyFuncSOCKS5(t *testing.T) {
	cfg := decodo.Config{
		Auth: decodo.Auth{
			Username: "username",
			Password: "password",
		},
	}

	proxyURL, err := nethttp.ProxyFuncSOCKS5(cfg)
	if err != nil {
		t.Fatalf("ProxyFuncSOCKS5() error = %v", err)
	}

	if proxyURL.Scheme != "socks5h" {
		t.Fatalf("scheme = %q, want %q", proxyURL.Scheme, "socks5h")
	}
}
