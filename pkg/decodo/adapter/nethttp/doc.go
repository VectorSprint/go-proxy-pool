// Package nethttp exposes helpers that adapt decodo configuration and leases
// into proxy values accepted by the Go standard library net/http package.
//
// # Proxy URL
//
// Convert a decodo Config into a *url.URL:
//
//	proxyURL, err := nethttpadapter.ProxyURL(config)
//
// # ProxyFunc
//
// Obtain a net/http-compatible proxy function for use with http.Transport:
//
//	proxyFunc, err := nethttpadapter.ProxyFunc(config)
//
//	transport := &http.Transport{Proxy: proxyFunc}
//
// # SOCKS5 Support
//
// net/http does not natively support SOCKS5. ProxyURLSOCKS5 returns a *url.URL
// suitable for use with a SOCKS5-capable dialer such as golang.org/x/net/proxy:
//
//	socks5URL, err := nethttpadapter.ProxyURLSOCKS5(config)
//
// Then configure your transport's Dial field with the SOCKS5 dialer.
package nethttp
