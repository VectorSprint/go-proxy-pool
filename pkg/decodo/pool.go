package decodo

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// FailureCause carries failure information reported by the caller.
type FailureCause struct {
	Err        error
	StatusCode int
}

// Lease represents one resolved sticky-session proxy assignment for a business key.
type Lease struct {
	Key       string
	SessionID string
	ProxyURL  string
	ExpiresAt time.Time
}

// PoolOptions configures how a keyed sticky-session pool behaves.
type PoolOptions struct {
	Config           Config
	FailureThreshold int
	Now              func() time.Time
	NewSessionID     func(key string) string
}

// Pool stores sticky-session leases keyed by caller-defined business identifiers.
type Pool struct {
	mu               sync.Mutex
	config           Config
	failureThreshold int
	now              func() time.Time
	newSessionID     func(key string) string
	entries          map[string]poolEntry
}

type poolEntry struct {
	lease        Lease
	failureCount int
}

// NewPool creates a keyed sticky-session pool from a Decodo configuration.
func NewPool(options PoolOptions) (*Pool, error) {
	config := options.Config
	if config.Session.Type == "" {
		config.Session.Type = SessionTypeSticky
	}
	if config.Session.Type != SessionTypeSticky {
		return nil, errors.New("pool requires sticky session configuration")
	}
	if config.Session.DurationMinutes == 0 {
		config.Session.DurationMinutes = defaultStickyDurationMinutes
	}

	normalized, err := config.Normalized()
	if err != nil {
		return nil, err
	}

	probe := normalized
	probe.Session.ID = "pool-probe"
	if err := probe.Validate(); err != nil {
		return nil, err
	}

	if options.FailureThreshold <= 0 {
		options.FailureThreshold = 1
	}
	if options.Now == nil {
		options.Now = time.Now
	}
	if options.NewSessionID == nil {
		options.NewSessionID = func(key string) string {
			return fmt.Sprintf("%s-%d", normalizeToken(key), options.Now().UnixNano())
		}
	}

	return &Pool{
		config:           normalized,
		failureThreshold: options.FailureThreshold,
		now:              options.Now,
		newSessionID:     options.NewSessionID,
		entries:          make(map[string]poolEntry),
	}, nil
}

// Get returns the active lease for the key or creates a new one when none exists or it expired.
func (p *Pool) Get(key string) (Lease, error) {
	if key == "" {
		return Lease{}, errors.New("key is required")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	current, ok := p.entries[key]
	if ok && !p.isExpired(current.lease) {
		return current.lease, nil
	}

	lease, err := p.newLease(key)
	if err != nil {
		return Lease{}, err
	}

	p.entries[key] = poolEntry{lease: lease}
	return lease, nil
}

// Rotate invalidates the current lease for the key so that the next Get allocates a new session.
func (p *Pool) Rotate(key string) error {
	if key == "" {
		return errors.New("key is required")
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.entries, key)
	return nil
}

// ReportFailure increments failure state for the key and rotates once the threshold is reached.
func (p *Pool) ReportFailure(key string, _ FailureCause) error {
	if key == "" {
		return errors.New("key is required")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	current, ok := p.entries[key]
	if !ok {
		return nil
	}

	current.failureCount++
	if current.failureCount >= p.failureThreshold {
		delete(p.entries, key)
		return nil
	}

	p.entries[key] = current
	return nil
}

// CleanupExpired removes expired leases and returns how many entries were deleted.
func (p *Pool) CleanupExpired() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	removed := 0
	for key, entry := range p.entries {
		if p.isExpired(entry.lease) {
			delete(p.entries, key)
			removed++
		}
	}

	return removed
}

func (p *Pool) isExpired(lease Lease) bool {
	return !lease.ExpiresAt.IsZero() && !p.now().Before(lease.ExpiresAt)
}

func (p *Pool) newLease(key string) (Lease, error) {
	sessionID := p.newSessionID(key)
	expiresAt := p.now().Add(p.config.Session.TTL())

	config := p.config
	config.Session.ID = sessionID

	proxyURL, err := config.ProxyURL()
	if err != nil {
		return Lease{}, err
	}

	return Lease{
		Key:       key,
		SessionID: sessionID,
		ProxyURL:  proxyURL.String(),
		ExpiresAt: expiresAt,
	}, nil
}
