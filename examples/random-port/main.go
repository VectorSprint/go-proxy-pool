package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/VectorSprint/go-proxy-pool/pkg/decodo"
)

// This example demonstrates:
// 1. Using ApplyPreset() to automatically select the correct endpoint based on targeting
// 2. Supporting country, city, and state targeting
// 3. Using RandomPort in pool to randomly select sticky ports from the available range

func main() {
	auth, err := decodo.NewAuth("my-proxy-user", "my-proxy-password")
	if err != nil {
		panic(err)
	}

	// Example 1: Country-level targeting
	fmt.Println("=== Country-level targeting (US) ===")
	cfg := decodo.Config{
		Auth: auth,
		Targeting: decodo.Targeting{
			Country: "us",
		},
		Session: decodo.Session{
			Type:            decodo.SessionTypeSticky,
			ID:              "session-us",
			DurationMinutes: 30,
		},
	}
	cfg.ApplyPreset()
	proxyURL, err := cfg.ProxyURL()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Proxy URL: %s\n", proxyURL.String())
	fmt.Printf("Host: %s\n\n", proxyURL.Host)

	// Example 2: City-level targeting
	fmt.Println("=== City-level targeting (Los Angeles) ===")
	cfg2 := decodo.Config{
		Auth: auth,
		Targeting: decodo.Targeting{
			Country: "us",
			City:    "los_angeles",
		},
		Session: decodo.Session{
			Type:            decodo.SessionTypeSticky,
			ID:              "session-la",
			DurationMinutes: 30,
		},
	}
	cfg2.ApplyPreset()
	proxyURL2, err := cfg2.ProxyURL()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Proxy URL: %s\n", proxyURL2.String())
	fmt.Printf("Host: %s\n\n", proxyURL2.Host)

	// Example 3: State-level targeting
	fmt.Println("=== State-level targeting (Texas) ===")
	cfg3 := decodo.Config{
		Auth: auth,
		Targeting: decodo.Targeting{
			Country: "us",
			State:   "us_texas",
		},
		Session: decodo.Session{
			Type:            decodo.SessionTypeSticky,
			ID:              "session-tx",
			DurationMinutes: 30,
		},
	}
	cfg3.ApplyPreset()
	proxyURL3, err := cfg3.ProxyURL()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Proxy URL: %s\n", proxyURL3.String())
	fmt.Printf("Host: %s\n\n", proxyURL3.Host)

	// Example 4: Random port pool
	fmt.Println("=== Random sticky port allocation (California) ===")
	cfg4 := decodo.Config{
		Auth: auth,
		Targeting: decodo.Targeting{
			Country: "us",
			State:   "us_california",
		},
		Session: decodo.Session{
			Type:            decodo.SessionTypeSticky,
			DurationMinutes: 30,
		},
	}
	cfg4.ApplyPreset()

	pool, err := decodo.NewPool(decodo.PoolOptions{
		Config:           cfg4,
		FailureThreshold: 3,
		RandomPort:       true,
		Rand:             rand.New(rand.NewSource(time.Now().UnixNano())),
	})
	if err != nil {
		panic(err)
	}

	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("account-%d", i)
		lease, err := pool.Get(key)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s: port %d\n", key, lease.Port)
	}
}
