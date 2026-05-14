// Package simpleproxy supports plain authenticated HTTP proxy entries.
package simpleproxy

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

// Proxy is a plain authenticated HTTP proxy endpoint.
type Proxy struct {
	Host     string
	Port     int
	Username string
	Password string
}

// Parse parses ip:port:username:password proxy entries.
func Parse(line string) (Proxy, error) {
	parts := strings.Split(strings.TrimSpace(line), ":")
	if len(parts) != 4 {
		return Proxy{}, errors.New("proxy line must use ip:port:username:password format")
	}

	host := strings.TrimSpace(parts[0])
	if ip := net.ParseIP(host); ip == nil || ip.To4() == nil {
		return Proxy{}, errors.New("proxy host must be an IPv4 address")
	}

	port, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil || port <= 0 {
		return Proxy{}, errors.New("proxy port must be positive")
	}

	proxy := Proxy{
		Host:     host,
		Port:     port,
		Username: strings.TrimSpace(parts[2]),
		Password: strings.TrimSpace(parts[3]),
	}
	if proxy.Username == "" {
		return Proxy{}, errors.New("proxy username is required")
	}
	if proxy.Password == "" {
		return Proxy{}, errors.New("proxy password is required")
	}

	return proxy, nil
}

// ParseLines parses newline-separated proxy entries, ignoring blank lines.
func ParseLines(lines string) ([]Proxy, error) {
	var proxies []Proxy
	for _, line := range strings.Split(lines, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}

		proxy, err := Parse(line)
		if err != nil {
			return nil, err
		}
		proxies = append(proxies, proxy)
	}

	return proxies, nil
}

// URL returns an authenticated HTTP proxy URL.
func (p Proxy) URL() *url.URL {
	return &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", p.Host, p.Port),
		User:   url.UserPassword(p.Username, p.Password),
	}
}

// IPv4Range expands a proxy by replacing its final IPv4 octet with each value
// in the inclusive range [start, end].
func IPv4Range(proxy Proxy, start, end int) ([]Proxy, error) {
	if start < 1 || end > 255 {
		return nil, errors.New("last octet range must be between 1 and 255")
	}
	if end < start {
		return nil, errors.New("last octet range end must be greater than or equal to start")
	}

	ip := net.ParseIP(proxy.Host).To4()
	if ip == nil {
		return nil, errors.New("proxy host must be an IPv4 address")
	}

	proxies := make([]Proxy, 0, end-start+1)
	for last := start; last <= end; last++ {
		next := proxy
		next.Host = fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], last)
		proxies = append(proxies, next)
	}

	return proxies, nil
}
