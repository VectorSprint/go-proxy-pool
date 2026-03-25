package main

import (
	"fmt"

	"github.com/VectorSprint/go-proxy-pool/pkg/decodo"
	httpcloakadapter "github.com/VectorSprint/go-proxy-pool/pkg/decodo/adapter/httpcloak"
)

func main() {
	auth, err := decodo.NewAuth("my-proxy-user", "my-proxy-password")
	if err != nil {
		panic(err)
	}

	cfg := decodo.Config{
		Auth: auth,
		Targeting: decodo.Targeting{
			Country: "de",
		},
		Session: decodo.Session{
			Type:            decodo.SessionTypeSticky,
			ID:              "job-1",
			DurationMinutes: 10,
		},
	}

	proxy, err := httpcloakadapter.ProxyString(cfg)
	if err != nil {
		panic(err)
	}

	fmt.Println("proxy for httpcloak:", proxy)
	fmt.Println("usage: client.SetProxy(proxy)")
}
