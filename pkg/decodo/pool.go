package decodo

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// FailureCause represents a proxy failure reported by the caller to Pool.ReportFailure.
type FailureCause struct {
	Err        error
	StatusCode int
}

// Lease represents a resolved sticky-session proxy assignment for a business key.
// Obtain a lease by calling Pool.Get. The lease remains valid until its
// ExpiresAt time has passed, unless explicitly rotated via Pool.Rotate.
type Lease struct {
	Key       string
	SessionID string
	Port      int
	ProxyURL  string
	ExpiresAt time.Time
}

// PoolOptions configures how a keyed sticky-session Pool behaves.
type PoolOptions struct {
	Config           Config
	FailureThreshold int
	Now              func() time.Time
	NewSessionID     func(key string) string
	// RandomPort selects a random sticky port from the available range instead of
	// sequentially allocating ports. This reduces detection risk when using a single
	// endpoint with many sessions.
	RandomPort bool
	// Rand is the random source for port selection. If nil, math/rand is used.
	Rand *rand.Rand
}

// Pool manages sticky-session proxy leases, each keyed by a caller-defined business identifier
// (e.g., user ID, order ID). The Pool ensures that each key consistently uses the same
// residential proxy for the session duration, then rotates automatically.
type Pool struct {
	mu               sync.Mutex
	config           Config
	failureThreshold int
	now              func() time.Time
	newSessionID     func(key string) string
	randomPort       bool
	rand             *rand.Rand
	portExplicit     bool
	nextStickyPort   int
	entries          map[string]poolEntry
}

type poolEntry struct {
	lease        Lease
	failureCount int
}

// NewPool creates a keyed sticky-session Pool from a Decodo Config.
// The Config must have Session.Type set to SessionTypeSticky. NewPool returns
// an error if the config validation fails or if a sticky session is not requested.
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
		randomPort:       options.RandomPort,
		rand:             options.Rand,
		portExplicit:     options.Config.Port != 0,
		nextStickyPort:   normalized.EndpointSpec.StickyPortRange.Start,
		entries:          make(map[string]poolEntry),
	}, nil
}

// Get returns the active Lease for the given business key. If no active lease exists
// or the existing lease has expired, a new one is allocated automatically.
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

// Rotate immediately invalidates the current lease for the key so that the next
// call to Get allocates a fresh session.
func (p *Pool) Rotate(key string) error {
	if key == "" {
		return errors.New("key is required")
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.entries, key)
	return nil
}

// ReportFailure records a failure for the given key. When the number of recorded
// failures reaches FailureThreshold, the lease is rotated automatically.
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

// CleanupExpired removes all expired leases from the pool and returns the number
// of entries deleted. Call this periodically (e.g., via a background goroutine)
// to prevent the pool from accumulating stale entries.
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
	port, err := p.selectPort()
	if err != nil {
		return Lease{}, err
	}

	config := p.config
	config.Session.ID = sessionID
	config.Port = port

	proxyURL, err := config.ProxyURL()
	if err != nil {
		return Lease{}, err
	}

	return Lease{
		Key:       key,
		SessionID: sessionID,
		Port:      port,
		ProxyURL:  proxyURL.String(),
		ExpiresAt: expiresAt,
	}, nil
}

func (p *Pool) selectPort() (int, error) {
	if p.portExplicit || p.config.Session.Type != SessionTypeSticky || p.config.EndpointSpec.IsZero() || p.config.EndpointSpec.StickyPortRange.IsZero() {
		return p.config.Port, nil
	}

	return p.allocateStickyPort()
}

func (p *Pool) allocateStickyPort() (int, error) {
	portRange := p.config.EndpointSpec.StickyPortRange
	if portRange.IsZero() {
		return p.config.Port, nil
	}

	if p.randomPort {
		return p.allocateRandomPort(portRange)
	}

	if p.nextStickyPort < portRange.Start || p.nextStickyPort > portRange.End {
		p.nextStickyPort = portRange.Start
	}

	candidate := p.nextStickyPort
	for attempts := 0; attempts < portRange.size(); attempts++ {
		if !p.portInUse(candidate) {
			p.nextStickyPort = candidate + 1
			if p.nextStickyPort > portRange.End {
				p.nextStickyPort = portRange.Start
			}
			return candidate, nil
		}

		candidate++
		if candidate > portRange.End {
			candidate = portRange.Start
		}
	}

	return 0, errors.New("no sticky ports available in the configured range")
}

func (p *Pool) allocateRandomPort(portRange PortRange) (int, error) {
	size := portRange.size()
	if size == 0 {
		return p.config.Port, nil
	}

	r := p.rand
	if r == nil {
		r = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	usedPorts := make(map[int]bool)
	for _, entry := range p.entries {
		if !p.isExpired(entry.lease) {
			usedPorts[entry.lease.Port] = true
		}
	}

	if len(usedPorts) >= size {
		return 0, errors.New("no sticky ports available in the configured range")
	}

	for attempts := 0; attempts < size; attempts++ {
		candidate := portRange.Start + r.Intn(size)
		if !usedPorts[candidate] {
			return candidate, nil
		}
	}

	return 0, errors.New("no sticky ports available in the configured range")
}

func (p *Pool) portInUse(port int) bool {
	for _, entry := range p.entries {
		if entry.lease.Port == port && !p.isExpired(entry.lease) {
			return true
		}
	}

	return false
}
