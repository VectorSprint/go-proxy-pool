package httpcloak

import (
	"net/url"

	"github.com/VectorSprint/go-proxy-pool/pkg/decodo"
)

// ProxyString converts a Decodo config into the proxy string accepted by httpcloak.
func ProxyString(config decodo.Config) (string, error) {
	proxyURL, err := config.ProxyURL()
	if err != nil {
		return "", err
	}

	return proxyURL.String(), nil
}

// ProxyStringFromLease returns the already-built proxy string stored in a lease.
func ProxyStringFromLease(lease decodo.Lease) string {
	return lease.ProxyURL
}

// ProxyStringSOCKS5 converts a Decodo config into a SOCKS5 proxy string for httpcloak.
// SOCKS5 always uses gate.decodo.com:7000 - targeting is done via username parameters.
// The "h" in socks5h:// means hostname resolution occurs on the proxy side.
func ProxyStringSOCKS5(config decodo.Config) (string, error) {
	proxyURL, err := proxyURLSOCKS5(config)
	if err != nil {
		return "", err
	}

	return proxyURL.String(), nil
}

// ProxyStringSOCKS5FromLease converts an HTTP proxy URL stored in a lease to SOCKS5 format.
func ProxyStringSOCKS5FromLease(lease decodo.Lease) (string, error) {
	proxyURL, err := url.Parse(lease.ProxyURL)
	if err != nil {
		return "", err
	}

	username := proxyURL.User.Username()
	password, _ := proxyURL.User.Password()

	socks5URL := &url.URL{
		Scheme: "socks5h",
		Host:   "gate.decodo.com:7000",
		User:   url.UserPassword(username, password),
	}

	return socks5URL.String(), nil
}

// proxyURLSOCKS5 builds a SOCKS5 proxy URL from a Decodo config.
func proxyURLSOCKS5(config decodo.Config) (*url.URL, error) {
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
