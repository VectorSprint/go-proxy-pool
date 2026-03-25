package main

import (
	"fmt"
	"net/http"

	"github.com/VectorSprint/go-proxy-pool/pkg/decodo"
	nethttpadapter "github.com/VectorSprint/go-proxy-pool/pkg/decodo/adapter/nethttp"
)

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
			ID:              "account-1",
			DurationMinutes: 30,
		},
	}

	proxyFunc, err := nethttpadapter.ProxyFunc(cfg)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	if err != nil {
		panic(err)
	}

	proxyURL, err := proxyFunc(req)
	if err != nil {
		panic(err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: proxyFunc,
		},
	}

	fmt.Println("proxy:", proxyURL.String())
	fmt.Printf("client transport configured: %t\n", client.Transport != nil)
}
