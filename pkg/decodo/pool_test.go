package decodo_test

import (
	"strings"
	"testing"
	"time"

	"github.com/VectorSprint/go-proxy-pool/pkg/decodo"
)

func TestPoolReusesStickySessionByKey(t *testing.T) {
	now := time.Date(2026, 3, 25, 13, 0, 0, 0, time.UTC)
	pool, err := decodo.NewPool(decodo.PoolOptions{
		Config: decodo.Config{
			Auth: decodo.Auth{
				Username: "username",
				Password: "password",
			},
			Session: decodo.Session{
				Type:            decodo.SessionTypeSticky,
				DurationMinutes: 30,
			},
		},
		FailureThreshold: 2,
		Now: func() time.Time {
			return now
		},
		NewSessionID: sequenceSessionIDs("session-1", "session-2"),
	})
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}

	first, err := pool.Get("account-1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	second, err := pool.Get("account-1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if first.SessionID != second.SessionID {
		t.Fatalf("session ids differ: %q vs %q", first.SessionID, second.SessionID)
	}

	if got := first.ExpiresAt; !got.Equal(now.Add(30 * time.Minute)) {
		t.Fatalf("ExpiresAt = %s, want %s", got, now.Add(30*time.Minute))
	}
}

func TestPoolRotateReplacesSessionForKey(t *testing.T) {
	now := time.Date(2026, 3, 25, 13, 0, 0, 0, time.UTC)
	pool, err := decodo.NewPool(decodo.PoolOptions{
		Config: decodo.Config{
			Auth: decodo.Auth{
				Username: "username",
				Password: "password",
			},
			Session: decodo.Session{
				Type:            decodo.SessionTypeSticky,
				DurationMinutes: 30,
			},
		},
		FailureThreshold: 2,
		Now: func() time.Time {
			return now
		},
		NewSessionID: sequenceSessionIDs("session-1", "session-2"),
	})
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}

	first, _ := pool.Get("account-1")

	if err := pool.Rotate("account-1"); err != nil {
		t.Fatalf("Rotate() error = %v", err)
	}

	second, _ := pool.Get("account-1")
	if first.SessionID == second.SessionID {
		t.Fatalf("Rotate() did not replace session id: %q", first.SessionID)
	}
}

func TestPoolRotatesExpiredSession(t *testing.T) {
	now := time.Date(2026, 3, 25, 13, 0, 0, 0, time.UTC)
	pool, err := decodo.NewPool(decodo.PoolOptions{
		Config: decodo.Config{
			Auth: decodo.Auth{
				Username: "username",
				Password: "password",
			},
			Session: decodo.Session{
				Type:            decodo.SessionTypeSticky,
				DurationMinutes: 30,
			},
		},
		FailureThreshold: 2,
		Now: func() time.Time {
			return now
		},
		NewSessionID: sequenceSessionIDs("session-1", "session-2"),
	})
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}

	first, _ := pool.Get("account-1")
	now = now.Add(31 * time.Minute)

	second, err := pool.Get("account-1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if first.SessionID == second.SessionID {
		t.Fatalf("expired session was reused: %q", first.SessionID)
	}
}

func TestPoolRotatesAfterFailureThreshold(t *testing.T) {
	now := time.Date(2026, 3, 25, 13, 0, 0, 0, time.UTC)
	pool, err := decodo.NewPool(decodo.PoolOptions{
		Config: decodo.Config{
			Auth: decodo.Auth{
				Username: "username",
				Password: "password",
			},
			Session: decodo.Session{
				Type:            decodo.SessionTypeSticky,
				DurationMinutes: 30,
			},
		},
		FailureThreshold: 2,
		Now: func() time.Time {
			return now
		},
		NewSessionID: sequenceSessionIDs("session-1", "session-2"),
	})
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}

	first, _ := pool.Get("account-1")

	if err := pool.ReportFailure("account-1", decodo.FailureCause{Err: assertErr("boom-1")}); err != nil {
		t.Fatalf("ReportFailure() error = %v", err)
	}
	if err := pool.ReportFailure("account-1", decodo.FailureCause{StatusCode: 429}); err != nil {
		t.Fatalf("ReportFailure() error = %v", err)
	}

	second, err := pool.Get("account-1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if first.SessionID == second.SessionID {
		t.Fatalf("failure threshold did not rotate session: %q", first.SessionID)
	}
}

func TestPoolLeaseBuildsStickyProxyURL(t *testing.T) {
	now := time.Date(2026, 3, 25, 13, 0, 0, 0, time.UTC)
	pool, err := decodo.NewPool(decodo.PoolOptions{
		Config: decodo.Config{
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
				DurationMinutes: 30,
			},
		},
		FailureThreshold: 2,
		Now: func() time.Time {
			return now
		},
		NewSessionID: sequenceSessionIDs("session-1"),
	})
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}

	lease, err := pool.Get("account-1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if !strings.Contains(lease.ProxyURL, "session-session-1-sessionduration-30") {
		t.Fatalf("ProxyURL = %q", lease.ProxyURL)
	}
}

func TestPoolCleanupExpiredRemovesExpiredEntries(t *testing.T) {
	now := time.Date(2026, 3, 25, 13, 0, 0, 0, time.UTC)
	pool, err := decodo.NewPool(decodo.PoolOptions{
		Config: decodo.Config{
			Auth: decodo.Auth{
				Username: "username",
				Password: "password",
			},
			Session: decodo.Session{
				Type:            decodo.SessionTypeSticky,
				DurationMinutes: 1,
			},
		},
		FailureThreshold: 2,
		Now: func() time.Time {
			return now
		},
		NewSessionID: sequenceSessionIDs("session-1", "session-2"),
	})
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}

	first, err := pool.Get("account-1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	now = now.Add(2 * time.Minute)
	removed := pool.CleanupExpired()
	if removed != 1 {
		t.Fatalf("CleanupExpired() = %d, want %d", removed, 1)
	}

	second, err := pool.Get("account-1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if first.SessionID == second.SessionID {
		t.Fatalf("expired session was not cleaned up: %q", first.SessionID)
	}
}

func TestPoolAssignsStickyPortsFromRange(t *testing.T) {
	spec, err := decodo.NewEndpointSpec("ca.decodo.com", 20000, decodo.PortRange{
		Start: 20001,
		End:   20003,
	})
	if err != nil {
		t.Fatalf("NewEndpointSpec() error = %v", err)
	}

	now := time.Date(2026, 3, 25, 13, 0, 0, 0, time.UTC)
	pool, err := decodo.NewPool(decodo.PoolOptions{
		Config: decodo.Config{
			Auth: decodo.Auth{
				Username: "username",
				Password: "password",
			},
			EndpointSpec: spec,
			Session: decodo.Session{
				Type:            decodo.SessionTypeSticky,
				DurationMinutes: 30,
			},
		},
		Now: func() time.Time {
			return now
		},
		NewSessionID: sequenceSessionIDs("session-1", "session-2"),
	})
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}

	first, err := pool.Get("account-1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	second, err := pool.Get("account-2")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if first.Port != 20001 {
		t.Fatalf("first port = %d, want %d", first.Port, 20001)
	}

	if second.Port != 20002 {
		t.Fatalf("second port = %d, want %d", second.Port, 20002)
	}
}

func TestPoolRandomPortSelection(t *testing.T) {
	spec, err := decodo.NewEndpointSpec("ca.decodo.com", 20000, decodo.PortRange{
		Start: 20001,
		End:   20010,
	})
	if err != nil {
		t.Fatalf("NewEndpointSpec() error = %v", err)
	}

	now := time.Date(2026, 3, 25, 13, 0, 0, 0, time.UTC)
	pool, err := decodo.NewPool(decodo.PoolOptions{
		Config: decodo.Config{
			Auth: decodo.Auth{
				Username: "username",
				Password: "password",
			},
			EndpointSpec: spec,
			Session: decodo.Session{
				Type:            decodo.SessionTypeSticky,
				DurationMinutes: 30,
			},
		},
		RandomPort: true,
		Now: func() time.Time {
			return now
		},
		NewSessionID: sequenceSessionIDs("session-1", "session-2"),
	})
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}

	lease1, err := pool.Get("account-1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if lease1.Port < 20001 || lease1.Port > 20010 {
		t.Fatalf("port %d out of range [20001, 20010]", lease1.Port)
	}

	lease2, err := pool.Get("account-2")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if lease2.Port < 20001 || lease2.Port > 20010 {
		t.Fatalf("port %d out of range [20001, 20010]", lease2.Port)
	}

	if lease1.Port == lease2.Port {
		t.Fatalf("expected different ports, got both %d", lease1.Port)
	}
}

func TestPoolAllocatesSequentialPortsExhaustingRange(t *testing.T) {
	spec, err := decodo.NewEndpointSpec("ca.decodo.com", 20000, decodo.PortRange{
		Start: 20001,
		End:   20003,
	})
	if err != nil {
		t.Fatalf("NewEndpointSpec() error = %v", err)
	}

	now := time.Date(2026, 3, 25, 13, 0, 0, 0, time.UTC)
	pool, err := decodo.NewPool(decodo.PoolOptions{
		Config: decodo.Config{
			Auth: decodo.Auth{
				Username: "username",
				Password: "password",
			},
			EndpointSpec: spec,
			Session: decodo.Session{
				Type:            decodo.SessionTypeSticky,
				DurationMinutes: 30,
			},
		},
		Now: func() time.Time {
			return now
		},
		NewSessionID: sequenceSessionIDs("s1", "s2", "s3", "s4"),
	})
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}

	// Fill all ports
	_, err = pool.Get("a")
	if err != nil {
		t.Fatalf("Get(a) error = %v", err)
	}
	_, err = pool.Get("b")
	if err != nil {
		t.Fatalf("Get(b) error = %v", err)
	}
	_, err = pool.Get("c")
	if err != nil {
		t.Fatalf("Get(c) error = %v", err)
	}

	// Now all ports in range [20001, 20003] should be in use
	// The 4th get should fail since all 3 ports are exhausted
	_, err = pool.Get("d")
	if err == nil {
		t.Fatal("Get(d) error = nil, want error when all ports exhausted")
	}
}

func TestPoolNewPoolRejectsRotatingSession(t *testing.T) {
	_, err := decodo.NewPool(decodo.PoolOptions{
		Config: decodo.Config{
			Auth: decodo.Auth{
				Username: "username",
				Password: "password",
			},
			Session: decodo.Session{
				Type:            decodo.SessionTypeRotating,
				DurationMinutes: 0,
			},
		},
	})
	if err == nil {
		t.Fatal("NewPool() error = nil, want error for rotating session")
	}
}

func TestPoolReportFailureIgnoresUnknownKey(t *testing.T) {
	pool, err := decodo.NewPool(decodo.PoolOptions{
		Config: decodo.Config{
			Auth: decodo.Auth{
				Username: "username",
				Password: "password",
			},
			Session: decodo.Session{
				Type:            decodo.SessionTypeSticky,
				DurationMinutes: 30,
			},
		},
	})
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}

	// Should not panic
	err = pool.ReportFailure("unknown-key", decodo.FailureCause{StatusCode: 500})
	if err != nil {
		t.Fatalf("ReportFailure(unknown) error = %v", err)
	}
}

func TestPoolRotateIgnoresUnknownKey(t *testing.T) {
	pool, err := decodo.NewPool(decodo.PoolOptions{
		Config: decodo.Config{
			Auth: decodo.Auth{
				Username: "username",
				Password: "password",
			},
			Session: decodo.Session{
				Type:            decodo.SessionTypeSticky,
				DurationMinutes: 30,
			},
		},
	})
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}

	// Should not panic
	err = pool.Rotate("unknown-key")
	if err != nil {
		t.Fatalf("Rotate(unknown) error = %v", err)
	}
}

func TestPoolGetRequiresNonEmptyKey(t *testing.T) {
	pool, err := decodo.NewPool(decodo.PoolOptions{
		Config: decodo.Config{
			Auth: decodo.Auth{
				Username: "username",
				Password: "password",
			},
			Session: decodo.Session{
				Type:            decodo.SessionTypeSticky,
				DurationMinutes: 30,
			},
		},
	})
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}

	_, err = pool.Get("")
	if err == nil {
		t.Fatal("Get('') error = nil, want error")
	}
}

// TestPoolRandomPortExhaustedReturnsError is hard to test deterministically
// since random port selection doesn't guarantee unique ports.
// The exhausted case is already covered by TestPoolAllocatesSequentialPortsExhaustingRange.

func sequenceSessionIDs(ids ...string) func(string) string {
	index := 0
	return func(string) string {
		current := ids[index]
		if index < len(ids)-1 {
			index++
		}
		return current
	}
}

type assertErr string

func (e assertErr) Error() string {
	return string(e)
}
