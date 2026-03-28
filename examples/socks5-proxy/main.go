package main

import (
	"fmt"

	httpcloakadapter "github.com/VectorSprint/go-proxy-pool/pkg/decodo/adapter/httpcloak"
	nethttpadapter "github.com/VectorSprint/go-proxy-pool/pkg/decodo/adapter/nethttp"
	"github.com/VectorSprint/go-proxy-pool/pkg/decodo"
)

// This example demonstrates SOCKS5 proxy support in both httpcloak and nethttp adapters.
// SOCKS5 uses gate.decodo.com:7000 exclusively - targeting is done via username parameters.

func main() {
	auth, err := decodo.NewAuth("my-proxy-user", "my-proxy-password")
	if err != nil {
		panic(err)
	}

	cfg := decodo.Config{
		Auth: auth,
		Targeting: decodo.Targeting{
			Country: "us",
			City:    "new_york",
		},
		Session: decodo.Session{
			Type:            decodo.SessionTypeSticky,
			ID:              "session-1",
			DurationMinutes: 30,
		},
	}

	// Example 1: httpcloak with SOCKS5
	fmt.Println("=== httpcloak SOCKS5 ===")
	socks5, err := httpcloakadapter.ProxyStringSOCKS5(cfg)
	if err != nil {
		panic(err)
	}
	fmt.Println(socks5)
	fmt.Println()

	// Example 2: net/http with SOCKS5
	fmt.Println("=== net/http SOCKS5 ===")
	socks5URL, err := nethttpadapter.ProxyURLSOCKS5(cfg)
	if err != nil {
		panic(err)
	}
	fmt.Printf("URL: %s\n", socks5URL.String())
	fmt.Println()

	// Example 3: From lease conversion
	fmt.Println("=== Convert HTTP lease to SOCKS5 ===")
	lease := decodo.Lease{
		ProxyURL: "http://user-my-proxy-user-country-us-city-new_york-session-s1-sessionduration-30:my-proxy-password@us.decodo.com:10001",
	}
	socks5FromLease, err := httpcloakadapter.ProxyStringSOCKS5FromLease(lease)
	if err != nil {
		panic(err)
	}
	fmt.Println(socks5FromLease)
}
