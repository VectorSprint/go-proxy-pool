package simpleproxy_test

import (
	"testing"

	"github.com/VectorSprint/go-proxy-pool/pkg/simpleproxy"
)

func TestParseProxyLineBuildsHTTPProxyURL(t *testing.T) {
	proxy, err := simpleproxy.Parse("40.27.182.2:3128:tgkoroke:tgkorfedwoksokfed")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if proxy.Host != "40.27.182.2" {
		t.Fatalf("Host = %q, want %q", proxy.Host, "40.27.182.2")
	}
	if proxy.Port != 3128 {
		t.Fatalf("Port = %d, want %d", proxy.Port, 3128)
	}
	if proxy.Username != "tgkoroke" {
		t.Fatalf("Username = %q, want %q", proxy.Username, "tgkoroke")
	}
	if proxy.Password != "tgkorfedwoksokfed" {
		t.Fatalf("Password = %q, want %q", proxy.Password, "tgkorfedwoksokfed")
	}

	if got := proxy.URL().String(); got != "http://tgkoroke:tgkorfedwoksokfed@40.27.182.2:3128" {
		t.Fatalf("URL() = %q", got)
	}
}

func TestParseLinesSkipsBlankLines(t *testing.T) {
	proxies, err := simpleproxy.ParseLines(`
40.27.182.2:3128:user:pass

40.27.182.3:3128:user:pass
`)
	if err != nil {
		t.Fatalf("ParseLines() error = %v", err)
	}

	if len(proxies) != 2 {
		t.Fatalf("len(proxies) = %d, want %d", len(proxies), 2)
	}
	if proxies[1].Host != "40.27.182.3" {
		t.Fatalf("second Host = %q, want %q", proxies[1].Host, "40.27.182.3")
	}
}

func TestParseLinesReturnsParseError(t *testing.T) {
	_, err := simpleproxy.ParseLines("40.27.182.2:3128:user:pass\nbad")
	if err == nil {
		t.Fatal("ParseLines() error = nil, want error")
	}
}

func TestIPv4RangeBuildsProxiesForVariableLastOctet(t *testing.T) {
	proxy, err := simpleproxy.Parse("40.27.182.2:3128:user:pass")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	proxies, err := simpleproxy.IPv4Range(proxy, 1, 3)
	if err != nil {
		t.Fatalf("IPv4Range() error = %v", err)
	}

	wantHosts := []string{"40.27.182.1", "40.27.182.2", "40.27.182.3"}
	if len(proxies) != len(wantHosts) {
		t.Fatalf("len(proxies) = %d, want %d", len(proxies), len(wantHosts))
	}
	for i, want := range wantHosts {
		if proxies[i].Host != want {
			t.Fatalf("proxies[%d].Host = %q, want %q", i, proxies[i].Host, want)
		}
		if proxies[i].URL().String() != "http://user:pass@"+want+":3128" {
			t.Fatalf("proxies[%d].URL() = %q", i, proxies[i].URL().String())
		}
	}
}

func TestIPv4RangeRejectsNonIPv4ProxyHost(t *testing.T) {
	_, err := simpleproxy.IPv4Range(simpleproxy.Proxy{
		Host:     "proxy.example.com",
		Port:     3128,
		Username: "user",
		Password: "pass",
	}, 1, 3)
	if err == nil {
		t.Fatal("IPv4Range() error = nil, want error")
	}
}

func TestParseRejectsInvalidProxyLines(t *testing.T) {
	testCases := []string{
		"",
		"40.27.182.2:3128:user",
		"not-an-ip:3128:user:pass",
		"40.27.182.2:not-a-port:user:pass",
		"40.27.182.2:0:user:pass",
		"40.27.182.2:3128::pass",
		"40.27.182.2:3128:user:",
	}

	for _, line := range testCases {
		t.Run(line, func(t *testing.T) {
			_, err := simpleproxy.Parse(line)
			if err == nil {
				t.Fatal("Parse() error = nil, want error")
			}
		})
	}
}

func TestIPv4RangeRejectsInvalidRanges(t *testing.T) {
	proxy, err := simpleproxy.Parse("40.27.182.2:3128:user:pass")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	testCases := []struct {
		name  string
		start int
		end   int
	}{
		{name: "zero start", start: 0, end: 3},
		{name: "above last octet", start: 1, end: 256},
		{name: "end before start", start: 3, end: 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := simpleproxy.IPv4Range(proxy, tc.start, tc.end)
			if err == nil {
				t.Fatal("IPv4Range() error = nil, want error")
			}
		})
	}
}
