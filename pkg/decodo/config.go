package decodo

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	defaultEndpoint              = "gate.decodo.com"
	defaultPort                  = 7000
	defaultStickyDurationMinutes = 10
)

type SessionType string

const (
	// SessionTypeRotating requests a new residential IP on each proxy request.
	SessionTypeRotating SessionType = "rotating"
	// SessionTypeSticky keeps the same residential IP for the configured session duration.
	SessionTypeSticky SessionType = "sticky"
)

// Config describes a Decodo user:pass backconnect proxy configuration.
type Config struct {
	Auth         Auth
	EndpointSpec EndpointSpec
	Endpoint     string
	Port         int
	Targeting    Targeting
	Session      Session
}

// Auth stores the raw Decodo proxy username and password from the dashboard.
type Auth struct {
	Username string
	Password string
}

// Targeting describes optional Decodo location and carrier targeting parameters.
type Targeting struct {
	Country   string
	City      string
	State     string
	ZIP       string
	Continent string
	ASN       int
}

// Session describes whether requests should rotate IPs or reuse a sticky session.
type Session struct {
	Type            SessionType
	ID              string
	DurationMinutes int
}

// EndpointSpec describes a Decodo endpoint together with its rotating port and sticky port range.
type EndpointSpec struct {
	Host            string
	RotatingPort    int
	StickyPortRange PortRange
}

// PortRange describes an inclusive port range.
type PortRange struct {
	Start int
	End   int
}

// TTL returns the sticky-session lifetime as a time.Duration.
func (s Session) TTL() time.Duration {
	if s.Type != SessionTypeSticky || s.DurationMinutes <= 0 {
		return 0
	}

	return time.Duration(s.DurationMinutes) * time.Minute
}

// NewEndpointSpec validates and returns a Decodo endpoint specification.
func NewEndpointSpec(host string, rotatingPort int, stickyPortRange PortRange) (EndpointSpec, error) {
	spec := EndpointSpec{
		Host:            strings.TrimSpace(strings.ToLower(host)),
		RotatingPort:    rotatingPort,
		StickyPortRange: stickyPortRange,
	}

	if err := spec.Validate(); err != nil {
		return EndpointSpec{}, err
	}

	return spec, nil
}

// Validate checks whether the endpoint specification is structurally valid.
func (e EndpointSpec) Validate() error {
	if e.IsZero() {
		return nil
	}

	if strings.TrimSpace(e.Host) == "" {
		return errors.New("endpoint spec host is required")
	}

	if e.RotatingPort <= 0 {
		return errors.New("endpoint spec rotating port must be positive")
	}

	return e.StickyPortRange.Validate()
}

// IsZero reports whether the endpoint specification is unset.
func (e EndpointSpec) IsZero() bool {
	return strings.TrimSpace(e.Host) == "" && e.RotatingPort == 0 && e.StickyPortRange.IsZero()
}

// Validate checks whether the port range is structurally valid.
func (r PortRange) Validate() error {
	if r.IsZero() {
		return nil
	}

	if r.Start <= 0 || r.End <= 0 {
		return errors.New("port range values must be positive")
	}

	if r.End < r.Start {
		return errors.New("port range end must be greater than or equal to start")
	}

	return nil
}

// IsZero reports whether the port range is unset.
func (r PortRange) IsZero() bool {
	return r.Start == 0 && r.End == 0
}

// Contains reports whether the range includes the provided port.
func (r PortRange) Contains(port int) bool {
	if r.IsZero() {
		return false
	}

	return port >= r.Start && port <= r.End
}

func (r PortRange) size() int {
	if r.IsZero() {
		return 0
	}

	return r.End - r.Start + 1
}

// NewAuth validates and normalizes raw Decodo dashboard credentials.
func NewAuth(username, password string) (Auth, error) {
	auth := Auth{
		Username: strings.TrimSpace(username),
		Password: strings.TrimSpace(password),
	}

	if err := auth.Validate(); err != nil {
		return Auth{}, err
	}

	return auth, nil
}

// Validate checks whether the credentials can be used to build a Decodo proxy username.
func (a Auth) Validate() error {
	if strings.TrimSpace(a.Username) == "" {
		return errors.New("username is required")
	}

	if strings.HasPrefix(strings.TrimSpace(a.Username), "user-") {
		return errors.New("username must be the raw decodo proxy username without the user- prefix")
	}

	if strings.TrimSpace(a.Password) == "" {
		return errors.New("password is required")
	}

	return nil
}

// Validate checks whether the configuration satisfies Decodo parameter constraints.
func (c Config) Validate() error {
	normalized, err := c.Normalized()
	if err != nil {
		return err
	}

	if err := normalized.Auth.Validate(); err != nil {
		return err
	}

	if err := normalized.EndpointSpec.Validate(); err != nil {
		return err
	}

	if normalized.Targeting.ASN > 0 {
		if normalized.Targeting.Country != "" || normalized.Targeting.City != "" || normalized.Targeting.State != "" || normalized.Targeting.ZIP != "" || normalized.Targeting.Continent != "" {
			return errors.New("asn cannot be combined with other targeting parameters")
		}
	}

	if normalized.Targeting.Continent != "" && (normalized.Targeting.Country != "" || normalized.Targeting.City != "" || normalized.Targeting.State != "" || normalized.Targeting.ZIP != "") {
		return errors.New("continent targeting cannot be combined with country, state, city, or zip")
	}

	if normalized.Targeting.City != "" && normalized.Targeting.Country == "" {
		return errors.New("city targeting requires country")
	}

	if normalized.Targeting.State != "" {
		if normalized.Targeting.Country != "us" {
			return errors.New("state targeting requires country us")
		}
		if !strings.HasPrefix(normalized.Targeting.State, "us_") {
			return errors.New("state targeting must use us_ prefix")
		}
	}

	if normalized.Targeting.ZIP != "" {
		if normalized.Targeting.Country != "us" {
			return errors.New("zip targeting requires country us")
		}
		if len(normalized.Targeting.ZIP) != 5 || !allDigits(normalized.Targeting.ZIP) {
			return errors.New("zip targeting requires a 5-digit zip")
		}
		if normalized.Targeting.City != "" || normalized.Targeting.State != "" || normalized.Targeting.Continent != "" {
			return errors.New("zip targeting cannot be combined with city, state, or continent")
		}
	}

	switch normalized.Session.Type {
	case "", SessionTypeRotating:
		if normalized.Session.DurationMinutes != 0 {
			return errors.New("rotating session cannot set duration")
		}
		if normalized.Session.ID != "" {
			return errors.New("rotating session cannot set session id")
		}
	case SessionTypeSticky:
		if normalized.Session.ID == "" {
			return errors.New("sticky session requires session id")
		}
		if normalized.Session.DurationMinutes < 1 || normalized.Session.DurationMinutes > 1440 {
			return errors.New("sticky session duration must be between 1 and 1440 minutes")
		}
	default:
		return fmt.Errorf("unsupported session type %q", normalized.Session.Type)
	}

	if normalized.Session.Type == SessionTypeSticky && !normalized.EndpointSpec.IsZero() && !normalized.EndpointSpec.StickyPortRange.IsZero() && !normalized.EndpointSpec.StickyPortRange.Contains(normalized.Port) {
		return errors.New("sticky session port must be inside the endpoint sticky port range")
	}

	return nil
}

// Normalized returns a copy of the configuration with defaults and normalized tokens applied.
func (c Config) Normalized() (Config, error) {
	normalized := c

	normalized.Auth.Username = strings.TrimSpace(normalized.Auth.Username)
	normalized.Auth.Password = strings.TrimSpace(normalized.Auth.Password)
	normalized.Endpoint = strings.TrimSpace(strings.ToLower(normalized.Endpoint))
	normalized.EndpointSpec.Host = strings.TrimSpace(strings.ToLower(normalized.EndpointSpec.Host))

	normalized.Targeting.Country = normalizeToken(normalized.Targeting.Country)
	normalized.Targeting.City = normalizeToken(normalized.Targeting.City)
	normalized.Targeting.State = normalizeToken(normalized.Targeting.State)
	normalized.Targeting.ZIP = strings.TrimSpace(normalized.Targeting.ZIP)
	normalized.Targeting.Continent = normalizeToken(normalized.Targeting.Continent)

	normalized.Session.ID = strings.TrimSpace(normalized.Session.ID)
	if normalized.Session.Type == "" {
		if normalized.Session.ID != "" {
			normalized.Session.Type = SessionTypeSticky
		} else {
			normalized.Session.Type = SessionTypeRotating
		}
	}

	if normalized.Session.Type == SessionTypeSticky && normalized.Session.DurationMinutes == 0 {
		normalized.Session.DurationMinutes = defaultStickyDurationMinutes
	}

	if normalized.EndpointSpec.IsZero() {
		if normalized.Endpoint == "" {
			normalized.Endpoint = defaultEndpoint
		}
		if normalized.Port == 0 {
			normalized.Port = defaultPort
		}
	} else {
		if normalized.Endpoint == "" {
			normalized.Endpoint = normalized.EndpointSpec.Host
		}
		if normalized.Port == 0 {
			if normalized.Session.Type == SessionTypeSticky && !normalized.EndpointSpec.StickyPortRange.IsZero() {
				normalized.Port = normalized.EndpointSpec.StickyPortRange.Start
			} else {
				normalized.Port = normalized.EndpointSpec.RotatingPort
			}
		}
	}

	if err := normalized.ValidateShallow(); err != nil {
		return Config{}, err
	}

	return normalized, nil
}

// ProxyUsername builds the Decodo proxy username, including targeting and session parameters.
func (c Config) ProxyUsername() (string, error) {
	normalized, err := c.Normalized()
	if err != nil {
		return "", err
	}

	if err := normalized.Validate(); err != nil {
		return "", err
	}

	parts := []string{"user", normalized.Auth.Username}

	if normalized.Targeting.Continent != "" {
		parts = append(parts, "continent", normalized.Targeting.Continent)
	}
	if normalized.Targeting.Country != "" {
		parts = append(parts, "country", normalized.Targeting.Country)
	}
	if normalized.Targeting.State != "" {
		parts = append(parts, "state", normalized.Targeting.State)
	}
	if normalized.Targeting.City != "" {
		parts = append(parts, "city", normalized.Targeting.City)
	}
	if normalized.Targeting.ZIP != "" {
		parts = append(parts, "zip", normalized.Targeting.ZIP)
	}
	if normalized.Targeting.ASN > 0 {
		parts = append(parts, "asn", strconv.Itoa(normalized.Targeting.ASN))
	}
	if normalized.Session.Type == SessionTypeSticky {
		parts = append(parts, "session", normalized.Session.ID, "sessionduration", strconv.Itoa(normalized.Session.DurationMinutes))
	}

	return strings.Join(parts, "-"), nil
}

// ProxyURL builds an authenticated Decodo proxy URL suitable for HTTP proxy clients.
func (c Config) ProxyURL() (*url.URL, error) {
	normalized, err := c.Normalized()
	if err != nil {
		return nil, err
	}

	username, err := normalized.ProxyUsername()
	if err != nil {
		return nil, err
	}

	return &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", normalized.Endpoint, normalized.Port),
		User:   url.UserPassword(username, normalized.Auth.Password),
	}, nil
}

// ValidateShallow checks lightweight structural constraints before full validation.
func (c Config) ValidateShallow() error {
	if c.Port < 0 {
		return errors.New("port must be positive")
	}
	return nil
}

func normalizeToken(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.ReplaceAll(value, " ", "_")
	return value
}

func allDigits(value string) bool {
	for _, r := range value {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
