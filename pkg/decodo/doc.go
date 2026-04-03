// Package decodo provides helpers for building Decodo residential proxy
// credentials, generating proxy URLs, and managing keyed sticky-session pools
// that can be reused by Go HTTP clients.
//
// # Credential Configuration
//
// Create an Auth from your Decodo user credentials:
//
//	auth, err := decodo.NewAuth("my-proxy-user", "my-proxy-password")
//
// Build a Config describing your desired proxy session and endpoint:
//
//	config := decodo.Config{
//	    Auth: auth,
//	    Session: decodo.Session{
//	        Type:             decodo.SessionTypeSticky,
//	        DurationMinutes:  10,
//	        Country:           "us",
//	        State:             "ny",  // optional US state filter
//	        City:              "new-york", // optional city filter
//	    },
//	}
//
// # Generating Proxy URLs
//
// Produce a proxy URL string suitable for httpcloak or a SOCKS5 dialer:
//
//	proxyURL, err := config.ProxyURL()
//
// # Sticky Session Pool
//
// For applications handling multiple business keys (e.g., per-user or per-order),
// use a Pool to manage sticky sessions with automatic rotation:
//
//	pool, err := decodo.NewPool(decodo.PoolOptions{Config: config})
//
//	lease, err := pool.Get("user-123")  // returns same proxy for this key
//	// ... use lease.ProxyURL with your HTTP client
//
// When a proxy fails, report it and the pool will rotate automatically:
//
//	pool.ReportFailure("user-123", decodo.FailureCause{Err: err})
//
// See https://pkg.go.dev/github.com/VectorSprint/go-proxy-pool/pkg/decodo/adapter/httpcloak
// and https://pkg.go.dev/github.com/VectorSprint/go-proxy-pool/pkg/decodo/adapter/nethttp
// for adapter packages that convert Config or Lease into proxy strings
// compatible with popular Go HTTP libraries.
package decodo
