package nethttp

import (
	"net/http"
	"net/url"

	"github.com/VectorSprint/go-proxy-pool/pkg/decodo"
)

// ProxyURL converts a Decodo config into a URL usable by net/http proxy helpers.
func ProxyURL(config decodo.Config) (*url.URL, error) {
	return config.ProxyURL()
}

// ProxyURLFromLease parses the proxy URL stored in a lease.
func ProxyURLFromLease(lease decodo.Lease) (*url.URL, error) {
	return url.Parse(lease.ProxyURL)
}

// ProxyFunc returns a net/http-compatible proxy function from a Decodo config.
func ProxyFunc(config decodo.Config) (func(*http.Request) (*url.URL, error), error) {
	proxyURL, err := ProxyURL(config)
	if err != nil {
		return nil, err
	}

	return http.ProxyURL(proxyURL), nil
}

// ProxyURLSOCKS5 converts a Decodo config into a SOCKS5 proxy URL.
// SOCKS5 always uses gate.decodo.com:7000 - targeting is done via username parameters.
func ProxyURLSOCKS5(config decodo.Config) (*url.URL, error) {
	socks5Config := config
	socks5Config.Endpoint = "gate.decodo.com"
	socks5Config.Port = 7000
	socks5Config.EndpointSpec = decodo.EndpointSpec{}

	proxyURL, err := socks5Config.ProxyURL()
	if err != nil {
		return nil, err
	}

	proxyURL.Scheme = "socks5h"

	return proxyURL, nil
}

// ProxyURLSOCKS5FromLease converts an HTTP proxy URL stored in a lease to SOCKS5 format.
func ProxyURLSOCKS5FromLease(lease decodo.Lease) (*url.URL, error) {
	proxyURL, err := url.Parse(lease.ProxyURL)
	if err != nil {
		return nil, err
	}

	username := proxyURL.User.Username()
	password, _ := proxyURL.User.Password()

	return &url.URL{
		Scheme: "socks5h",
		Host:   "gate.decodo.com:7000",
		User:   url.UserPassword(username, password),
	}, nil
}

// ProxyFuncSOCKS5 returns a SOCKS5 proxy URL function for net/http.
// Note: net/http does not natively support SOCKS5. This returns the URL
// for use with a SOCKS5-capable dialer (e.g., golang.org/x/net/proxy).
func ProxyFuncSOCKS5(config decodo.Config) (*url.URL, error) {
	return ProxyURLSOCKS5(config)
}
