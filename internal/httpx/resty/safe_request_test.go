package resty

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	grestry "resty.dev/v3"
)

func TestSafeRequestRejectsPrivateAndUnsafeRedirects(t *testing.T) {
	if err := GetSafe("file:///tmp/secret.txt").Execute().Err(); err == nil {
		t.Fatal("GetSafe should reject non-HTTP schemes")
	}
	if err := GetSafe("http://127.0.0.1/config.yaml").Execute().Err(); err == nil {
		t.Fatal("GetSafe should reject loopback hosts by default")
	}
	if err := GetSafe("http://224.0.0.1/config.yaml").Execute().Err(); err == nil {
		t.Fatal("GetSafe should reject multicast hosts by default")
	}
	if err := GetSafe("http://0.0.0.0/config.yaml").Execute().Err(); err == nil {
		t.Fatal("GetSafe should reject unspecified hosts by default")
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "http://127.0.0.1/private", http.StatusFound)
			return
		}
		_, _ = w.Write([]byte("safe"))
	}))
	defer srv.Close()
	serverURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	resp := GetSafe(
		srv.URL,
		WithURLPolicy(URLPolicy{AllowedSchemes: []string{"http", "https"}, AllowedHosts: []string{serverURL.Hostname()}}),
	).Execute()
	if resp.Err() != nil || resp.Body() != "safe" {
		t.Fatalf("GetSafe allowed public policy host body=%q err=%v", resp.Body(), resp.Err())
	}
	if err := GetSafe(
		srv.URL+"/redirect",
		WithURLPolicy(URLPolicy{AllowedSchemes: []string{"http", "https"}, AllowedHosts: []string{serverURL.Hostname()}}),
	).Execute().Err(); err == nil {
		t.Fatal("GetSafe should reject unsafe redirect targets")
	}
}

func TestSafeRequestAllowedHostsDoesNotBypassPrivateRejection(t *testing.T) {
	if err := GetSafe("http://127.0.0.1/config.yaml", WithAllowedHosts("127.0.0.1")).Execute().Err(); err == nil {
		t.Fatal("GetSafe should reject allowlisted private hosts when RejectPrivate is enabled")
	}

	client := grestry.New().SetTransport(restyRoundTripperFunc(func(*http.Request) (*http.Response, error) {
		t.Fatal("unsafe request reached base transport")
		return nil, nil
	}))
	resp := GetSafe(
		"http://example.com/config.yaml",
		WithAllowedHosts("example.com"),
		WithRestyClient(client),
		WithLookupIP(func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		}),
	).Execute()
	if resp.Err() == nil {
		t.Fatal("GetSafe should reject allowlisted hosts that resolve private during RoundTrip")
	}
}

func TestSafeRequestRevalidatesHostAtRoundTrip(t *testing.T) {
	lookups := [][]net.IP{{net.ParseIP("93.184.216.34")}, {net.ParseIP("127.0.0.1")}}
	lookupCount := 0
	client := grestry.New().SetTransport(restyRoundTripperFunc(func(*http.Request) (*http.Response, error) {
		t.Fatal("unsafe request reached base transport")
		return nil, nil
	}))
	resp := GetSafe(
		"http://example.com/config.yaml",
		WithRestyClient(client),
		WithLookupIP(func(context.Context, string) ([]net.IP, error) {
			if lookupCount >= len(lookups) {
				return lookups[len(lookups)-1], nil
			}
			ips := lookups[lookupCount]
			lookupCount++
			return ips, nil
		}),
	).Execute()
	if resp.Err() == nil {
		t.Fatal("GetSafe should reject a host that resolves private during RoundTrip")
	}
	if lookupCount != 2 {
		t.Fatalf("lookup count = %d, want 2", lookupCount)
	}
}
