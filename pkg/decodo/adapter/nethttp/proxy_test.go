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
