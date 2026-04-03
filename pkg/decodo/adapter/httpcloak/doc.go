// Package httpcloak exposes helpers that adapt decodo configuration and leases
// into proxy strings accepted by github.com/sardanioss/httpcloak.
//
// # Proxy Strings
//
// Convert a decodo Config into a proxy string for httpcloak:
//
//	proxyStr, err := httpcloakadapter.ProxyString(config)
//
// Or use a Lease already obtained from a Pool:
//
//	proxyStr := httpcloakadapter.ProxyStringFromLease(lease)
//
// # SOCKS5 Support
//
// For SOCKS5 proxies, use ProxyStringSOCKS5. Note that SOCKS5 targeting
// is done via username parameters; the scheme socks5h:// means DNS resolution
// occurs on the proxy side:
//
//	socks5Str, err := httpcloakadapter.ProxyStringSOCKS5(config)
package httpcloak
