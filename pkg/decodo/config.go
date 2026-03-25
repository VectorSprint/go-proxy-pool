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
	SessionTypeRotating SessionType = "rotating"
	SessionTypeSticky   SessionType = "sticky"
)

type Config struct {
	Auth      Auth
	Endpoint  string
	Port      int
	Targeting Targeting
	Session   Session
}

type Auth struct {
	Username string
	Password string
}

type Targeting struct {
	Country   string
	City      string
	State     string
	ZIP       string
	Continent string
	ASN       int
}

type Session struct {
	Type            SessionType
	ID              string
	DurationMinutes int
}

func (s Session) TTL() time.Duration {
	if s.Type != SessionTypeSticky || s.DurationMinutes <= 0 {
		return 0
	}

	return time.Duration(s.DurationMinutes) * time.Minute
}

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

func (c Config) Validate() error {
	normalized, err := c.Normalized()
	if err != nil {
		return err
	}

	if err := normalized.Auth.Validate(); err != nil {
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

	return nil
}

func (c Config) Normalized() (Config, error) {
	normalized := c

	normalized.Endpoint = strings.TrimSpace(normalized.Endpoint)
	if normalized.Endpoint == "" {
		normalized.Endpoint = defaultEndpoint
	}

	if normalized.Port == 0 {
		normalized.Port = defaultPort
	}

	normalized.Auth.Username = strings.TrimSpace(normalized.Auth.Username)
	normalized.Auth.Password = strings.TrimSpace(normalized.Auth.Password)

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

	if err := normalized.ValidateShallow(); err != nil {
		return Config{}, err
	}

	return normalized, nil
}

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
