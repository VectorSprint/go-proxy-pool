package httpcloak

import "github.com/VectorSprint/go-proxy-pool/pkg/decodo"

func ProxyString(config decodo.Config) (string, error) {
	proxyURL, err := config.ProxyURL()
	if err != nil {
		return "", err
	}

	return proxyURL.String(), nil
}

func ProxyStringFromLease(lease decodo.Lease) string {
	return lease.ProxyURL
}
