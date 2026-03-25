package nethttp

import (
	"net/http"
	"net/url"

	"github.com/VectorSprint/go-proxy-pool/pkg/decodo"
)

func ProxyURL(config decodo.Config) (*url.URL, error) {
	return config.ProxyURL()
}

func ProxyURLFromLease(lease decodo.Lease) (*url.URL, error) {
	return url.Parse(lease.ProxyURL)
}

func ProxyFunc(config decodo.Config) (func(*http.Request) (*url.URL, error), error) {
	proxyURL, err := ProxyURL(config)
	if err != nil {
		return nil, err
	}

	return http.ProxyURL(proxyURL), nil
}
