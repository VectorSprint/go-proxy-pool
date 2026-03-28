package nethttp_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/VectorSprint/go-proxy-pool/pkg/decodo"
	"github.com/VectorSprint/go-proxy-pool/pkg/decodo/adapter/nethttp"
)

type liveProxyResponse struct {
	Proxy struct {
		IP string `json:"ip"`
	} `json:"proxy"`
	Country struct {
		Code string `json:"code"`
	} `json:"country"`
}

func TestProxyURLSOCKS5LiveConnectivity(t *testing.T) {
	auth := liveTestAuth(t)

	cfg := decodo.Config{
		Auth: auth,
		Targeting: decodo.Targeting{
			Country: "us",
		},
		Session: decodo.Session{
			Type:            decodo.SessionTypeSticky,
			ID:              fmt.Sprintf("live-socks5-%d", time.Now().UnixNano()),
			DurationMinutes: 10,
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

	first := curlLiveProxy(t, proxyURL.String(), 40*time.Second)
	if first.Country.Code != "US" {
		t.Fatalf("country code = %q, want %q", first.Country.Code, "US")
	}
	if first.Proxy.IP == "" {
		t.Fatal("proxy ip is empty")
	}

	second := curlLiveProxy(t, proxyURL.String(), 40*time.Second)
	if second.Proxy.IP != first.Proxy.IP {
		t.Fatalf("sticky socks5 ip mismatch: first=%q second=%q", first.Proxy.IP, second.Proxy.IP)
	}
}

func TestCountryEndpointDoesNotWorkForSOCKS5Live(t *testing.T) {
	auth := liveTestAuth(t)

	cfg := decodo.Config{
		Auth: auth,
		Targeting: decodo.Targeting{
			Country: "us",
		},
		Session: decodo.Session{
			Type:            decodo.SessionTypeSticky,
			ID:              fmt.Sprintf("live-socks5-country-endpoint-%d", time.Now().UnixNano()),
			DurationMinutes: 10,
		},
	}

	proxyURL, err := nethttp.ProxyURLSOCKS5(cfg)
	if err != nil {
		t.Fatalf("ProxyURLSOCKS5() error = %v", err)
	}

	proxyURL.Host = "us.decodo.com:7000"

	output, err := curlLiveProxyOutput(proxyURL.String(), 15*time.Second)
	if err == nil {
		t.Fatalf("expected socks5 country endpoint to fail, got success: %s", output)
	}

	lowerOutput := strings.ToLower(output)
	if !strings.Contains(lowerOutput, "timed out") && !strings.Contains(lowerOutput, "timeout") {
		t.Fatalf("expected timeout-like failure for socks5 country endpoint, got: %s", output)
	}
}

func liveTestAuth(t *testing.T) decodo.Auth {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping live socks5 test in short mode")
	}

	username := strings.TrimSpace(os.Getenv("DECODO_SOCKS5_TEST_USERNAME"))
	password := strings.TrimSpace(os.Getenv("DECODO_SOCKS5_TEST_PASSWORD"))
	if username == "" || password == "" {
		t.Skip("set DECODO_SOCKS5_TEST_USERNAME and DECODO_SOCKS5_TEST_PASSWORD to run live socks5 tests")
	}

	auth, err := decodo.NewAuth(username, password)
	if err != nil {
		t.Fatalf("NewAuth() error = %v", err)
	}

	if _, err := exec.LookPath("curl"); err != nil {
		t.Skip("curl is required to run live socks5 tests")
	}

	return auth
}

func curlLiveProxy(t *testing.T, proxyURL string, timeout time.Duration) liveProxyResponse {
	t.Helper()

	output, err := curlLiveProxyOutput(proxyURL, timeout)
	if err != nil {
		t.Fatalf("curl through proxy failed: %v\n%s", err, output)
	}

	var response liveProxyResponse
	if err := json.Unmarshal([]byte(output), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v\n%s", err, output)
	}

	return response
}

func curlLiveProxyOutput(proxyURL string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout+5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(
		ctx,
		"curl",
		"-sS",
		"--max-time", fmt.Sprintf("%.0f", timeout.Seconds()),
		"-x", proxyURL,
		"https://ip.decodo.com/json",
	)
	output, err := cmd.CombinedOutput()
	if ctx.Err() != nil && errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return string(output), ctx.Err()
	}

	return string(output), err
}
