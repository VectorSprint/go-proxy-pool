package httpcloak

import "github.com/VectorSprint/go-proxy-pool/pkg/decodo"

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
