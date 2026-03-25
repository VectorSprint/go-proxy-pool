package decodo_test

import (
	"fmt"

	"github.com/VectorSprint/go-proxy-pool/pkg/decodo"
	decodohttpcloak "github.com/VectorSprint/go-proxy-pool/pkg/decodo/adapter/httpcloak"
)

func ExampleNewAuth() {
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

	proxyURL, err := cfg.ProxyURL()
	if err != nil {
		panic(err)
	}

	fmt.Println(proxyURL.String())
	// Output:
	// http://user-my-proxy-user-country-us-city-new_york-session-account-1-sessionduration-30:my-proxy-password@gate.decodo.com:7000
}

func ExamplePool_Get() {
	auth, err := decodo.NewAuth("my-proxy-user", "my-proxy-password")
	if err != nil {
		panic(err)
	}

	pool, err := decodo.NewPool(decodo.PoolOptions{
		Config: decodo.Config{
			Auth: auth,
			Session: decodo.Session{
				Type:            decodo.SessionTypeSticky,
				DurationMinutes: 30,
			},
		},
		NewSessionID: func(key string) string {
			return "session-" + key
		},
	})
	if err != nil {
		panic(err)
	}

	lease, err := pool.Get("account-1")
	if err != nil {
		panic(err)
	}

	fmt.Println(decodohttpcloak.ProxyStringFromLease(lease))
	// Output:
	// http://user-my-proxy-user-session-session-account-1-sessionduration-30:my-proxy-password@gate.decodo.com:7000
}
