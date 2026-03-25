package main

import (
	"fmt"

	"github.com/VectorSprint/go-proxy-pool/pkg/decodo"
)

func main() {
	auth, err := decodo.NewAuth("my-proxy-user", "my-proxy-password")
	if err != nil {
		panic(err)
	}

	caEndpoint, err := decodo.NewEndpointSpec("ca.decodo.com", 20000, decodo.PortRange{
		Start: 20001,
		End:   29999,
	})
	if err != nil {
		panic(err)
	}

	cfg := decodo.Config{
		Auth:         auth,
		EndpointSpec: caEndpoint,
		Session: decodo.Session{
			Type:            decodo.SessionTypeSticky,
			ID:              "account-1",
			DurationMinutes: 30,
		},
	}

	proxyURL, err := cfg.ProxyURL()
	if err != nil {
		panic(err)
	}

	fmt.Println("sticky dedicated endpoint:", proxyURL.String())
}
