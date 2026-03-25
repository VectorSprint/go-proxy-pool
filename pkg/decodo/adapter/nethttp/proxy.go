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
