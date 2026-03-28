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

func TestConfigPresetReturnsUSEndpoint(t *testing.T) {
	cfg := decodo.Config{
		Auth: decodo.Auth{
			Username: "username",
			Password: "password",
		},
		Targeting: decodo.Targeting{
			Country: "us",
		},
	}

	preset, ok := cfg.Preset()
	if !ok {
		t.Fatal("Preset() returned false, want true for US country")
	}

	if preset.Host != "us.decodo.com" {
		t.Fatalf("preset.Host = %q, want %q", preset.Host, "us.decodo.com")
	}

	if preset.RotatingPort != 10000 {
		t.Fatalf("preset.RotatingPort = %d, want %d", preset.RotatingPort, 10000)
	}

	if preset.StickyPortRange.Start != 10001 || preset.StickyPortRange.End != 29999 {
		t.Fatalf("sticky port range = %v, want {Start: 10001, End: 29999}", preset.StickyPortRange)
	}
}

func TestConfigPresetReturnsCityEndpoint(t *testing.T) {
	cfg := decodo.Config{
		Auth: decodo.Auth{
			Username: "username",
			Password: "password",
		},
		Targeting: decodo.Targeting{
			Country: "us",
			City:    "new_york",
		},
	}

	preset, ok := cfg.Preset()
	if !ok {
		t.Fatal("Preset() returned false, want true for new_york city")
	}

	if preset.Host != "city.decodo.com" {
		t.Fatalf("preset.Host = %q, want %q", preset.Host, "city.decodo.com")
	}

	if preset.RotatingPort != 21000 {
		t.Fatalf("preset.RotatingPort = %d, want %d", preset.RotatingPort, 21000)
	}

	if preset.StickyPortRange.Start != 21001 || preset.StickyPortRange.End != 21049 {
		t.Fatalf("sticky port range = %v, want {Start: 21001, End: 21049}", preset.StickyPortRange)
	}
}

func TestConfigPresetReturnsFalseWhenEndpointSpecSet(t *testing.T) {
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
			Country: "us",
		},
	}

	_, ok := cfg.Preset()
	if ok {
		t.Fatal("Preset() returned true, want false when EndpointSpec is set")
	}
}

func TestConfigPresetReturnsFalseForUnknownCountry(t *testing.T) {
	cfg := decodo.Config{
		Auth: decodo.Auth{
			Username: "username",
			Password: "password",
		},
		Targeting: decodo.Targeting{
			Country: "xx",
		},
	}

	_, ok := cfg.Preset()
	if ok {
		t.Fatal("Preset() returned true, want false for unknown country")
	}
}

func TestApplyPresetUpdatesEndpointSpec(t *testing.T) {
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

	if !cfg.EndpointSpec.IsZero() {
		t.Fatal("expected EndpointSpec to be zero before ApplyPreset")
	}

	cfg.ApplyPreset()

	if cfg.EndpointSpec.IsZero() {
		t.Fatal("expected EndpointSpec to be non-zero after ApplyPreset")
	}

	if cfg.EndpointSpec.Host != "city.decodo.com" {
		t.Fatalf("Host = %q, want %q", cfg.EndpointSpec.Host, "city.decodo.com")
	}

	if cfg.EndpointSpec.RotatingPort != 21000 {
		t.Fatalf("RotatingPort = %d, want %d", cfg.EndpointSpec.RotatingPort, 21000)
	}
}

func TestApplyPresetDoesNothingWhenNoPreset(t *testing.T) {
	cfg := decodo.Config{
		Auth: decodo.Auth{
			Username: "username",
			Password: "password",
		},
		Targeting: decodo.Targeting{
			Country: "xx",
		},
	}

	cfg.ApplyPreset()

	if !cfg.EndpointSpec.IsZero() {
		t.Fatal("expected EndpointSpec to remain zero when no preset matches")
	}
}

func TestConfigPresetReturnsStateEndpoint(t *testing.T) {
	cfg := decodo.Config{
		Auth: decodo.Auth{
			Username: "username",
			Password: "password",
		},
		Targeting: decodo.Targeting{
			Country: "us",
			State:   "us_california",
		},
	}

	preset, ok := cfg.Preset()
	if !ok {
		t.Fatal("Preset() returned false, want true for california state")
	}

	if preset.Host != "state.decodo.com" {
		t.Fatalf("preset.Host = %q, want %q", preset.Host, "state.decodo.com")
	}

	if preset.RotatingPort != 10000 {
		t.Fatalf("preset.RotatingPort = %d, want %d", preset.RotatingPort, 10000)
	}

	if preset.StickyPortRange.Start != 10001 || preset.StickyPortRange.End != 10999 {
		t.Fatalf("sticky port range = %v, want {Start: 10001, End: 10999}", preset.StickyPortRange)
	}
}

func TestConfigValidationRejectsInvalidZIP(t *testing.T) {
	cfg := decodo.Config{
		Auth: decodo.Auth{
			Username: "username",
			Password: "password",
		},
		Targeting: decodo.Targeting{
			Country: "us",
			ZIP:     "1234a",
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() error = nil, want error for non-numeric ZIP")
	}
}

func TestPortRangeValidateRejectsEndLessThanStart(t *testing.T) {
	err := decodo.PortRange{Start: 100, End: 50}.Validate()
	if err == nil {
		t.Fatal("Validate() error = nil, want error when End < Start")
	}
}

func TestPortRangeContains(t *testing.T) {
	pr := decodo.PortRange{Start: 10001, End: 10010}

	if !pr.Contains(10005) {
		t.Fatal("Contains(10005) = false, want true")
	}

	if pr.Contains(10000) {
		t.Fatal("Contains(10000) = true, want false (below range)")
	}

	if pr.Contains(10011) {
		t.Fatal("Contains(10011) = true, want false (above range)")
	}
}

func TestPortRangeContainsFalseForZeroRange(t *testing.T) {
	pr := decodo.PortRange{}

	if pr.Contains(100) {
		t.Fatal("Contains(100) = true for zero range, want false")
	}
}

func TestSessionTTLReturnsZeroForRotating(t *testing.T) {
	s := decodo.Session{
		Type:            decodo.SessionTypeRotating,
		DurationMinutes: 30,
	}

	if s.TTL() != 0 {
		t.Fatalf("TTL() = %v, want 0 for rotating session", s.TTL())
	}
}

func TestSessionTTLReturnsZeroForZeroDuration(t *testing.T) {
	s := decodo.Session{
		Type:            decodo.SessionTypeSticky,
		DurationMinutes: 0,
	}

	if s.TTL() != 0 {
		t.Fatalf("TTL() = %v, want 0 for zero duration", s.TTL())
	}
}

func TestEndpointSpecValidateRejectsZeroHost(t *testing.T) {
	spec := decodo.EndpointSpec{
		Host:         "   ",
		RotatingPort: 10000,
	}

	err := spec.Validate()
	if err == nil {
		t.Fatal("Validate() error = nil, want error for whitespace-only host")
	}
}

func TestEndpointSpecValidateRejectsNegativePort(t *testing.T) {
	spec := decodo.EndpointSpec{
		Host:         "gate.decodo.com",
		RotatingPort: -1,
	}

	err := spec.Validate()
	if err == nil {
		t.Fatal("Validate() error = nil, want error for negative rotating port")
	}
}

func TestEndpointSpecValidateAcceptsZeroStickyPortRange(t *testing.T) {
	spec := decodo.EndpointSpec{
		Host:         "gate.decodo.com",
		RotatingPort: 7000,
	}

	err := spec.Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v, want nil for zero sticky port range", err)
	}
}

func TestEndpointSpecValidateRejectsInvalidStickyPortRange(t *testing.T) {
	spec := decodo.EndpointSpec{
		Host:            "gate.decodo.com",
		RotatingPort:    7000,
		StickyPortRange: decodo.PortRange{Start: 100, End: 50},
	}

	err := spec.Validate()
	if err == nil {
		t.Fatal("Validate() error = nil, want error for invalid sticky port range")
	}
}
