package main

import (
	"fmt"

	"github.com/VectorSprint/go-proxy-pool/pkg/decodo"
)

func main() {
	counter := 0

	auth, err := decodo.NewAuth("my-proxy-user", "my-proxy-password")
	if err != nil {
		panic(err)
	}

	pool, err := decodo.NewPool(decodo.PoolOptions{
		Config: decodo.Config{
			Auth: auth,
			Targeting: decodo.Targeting{
				Country: "us",
			},
			Session: decodo.Session{
				Type:            decodo.SessionTypeSticky,
				DurationMinutes: 30,
			},
		},
		FailureThreshold: 2,
		NewSessionID: func(key string) string {
			counter++
			return fmt.Sprintf("session-%s-%d", key, counter)
		},
	})
	if err != nil {
		panic(err)
	}

	lease, err := pool.Get("account-1")
	if err != nil {
		panic(err)
	}

	fmt.Println("session:", lease.SessionID)
	fmt.Println("proxy:", lease.ProxyURL)

	if err := pool.ReportFailure("account-1", decodo.FailureCause{StatusCode: 429}); err != nil {
		panic(err)
	}
	if err := pool.ReportFailure("account-1", decodo.FailureCause{StatusCode: 429}); err != nil {
		panic(err)
	}

	rotatedLease, err := pool.Get("account-1")
	if err != nil {
		panic(err)
	}

	fmt.Println("rotated session:", rotatedLease.SessionID)
}
