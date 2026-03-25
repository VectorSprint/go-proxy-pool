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
