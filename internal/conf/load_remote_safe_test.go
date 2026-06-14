package conf

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestLoadRemoteSafeRejectsPrivateHostsAndUnsafeRedirects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("app:\n  name: remote"))
	}))
	defer server.Close()

	if _, err := LoadRemoteSafe(server.URL + "/app.yaml"); err == nil {
		t.Fatal("LoadRemoteSafe should reject private hosts by default")
	}
	if _, err := LoadRemoteSafe("http://224.0.0.1/app.yaml"); err == nil {
		t.Fatal("LoadRemoteSafe should reject multicast hosts by default")
	}
	if _, err := LoadRemoteSafe("http://0.0.0.0/app.yaml"); err == nil {
		t.Fatal("LoadRemoteSafe should reject unspecified hosts by default")
	}
	remoteURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := LoadRemoteSafeWithOptions(server.URL+"/app.yaml", LoadOptions{RemoteAllowedHosts: []string{remoteURL.Hostname()}}); err == nil {
		t.Fatal("LoadRemoteSafeWithOptions should reject allowlisted private hosts")
	}

	redirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://127.0.0.1/private.yaml", http.StatusFound)
	}))
	defer redirect.Close()
	redirectURL, err := url.Parse(redirect.URL)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := LoadRemoteSafeWithOptions(redirect.URL+"/app.yaml", LoadOptions{RemoteAllowedHosts: []string{redirectURL.Hostname()}}); err == nil {
		t.Fatal("LoadRemoteSafeWithOptions should reject unsafe redirect target")
	}
}

func TestLoadRemoteSafeAllowedHostsDoesNotBypassPrivateRejection(t *testing.T) {
	if _, err := LoadRemoteSafeWithOptions("http://127.0.0.1/app.yaml", LoadOptions{RemoteAllowedHosts: []string{"127.0.0.1"}}); err == nil {
		t.Fatal("LoadRemoteSafeWithOptions should reject allowlisted loopback host")
	}

	lookupCount := 0
	_, err := LoadRemoteSafeWithOptions("http://config.example/app.yaml", LoadOptions{
		RemoteAllowedHosts: []string{"config.example"},
		LookupIP: func(context.Context, string) ([]net.IP, error) {
			lookupCount++
			return []net.IP{net.ParseIP("10.0.0.1")}, nil
		},
	})
	if err == nil {
		t.Fatal("LoadRemoteSafeWithOptions should reject allowlisted host resolving to private address")
	}
	if lookupCount == 0 {
		t.Fatal("LoadRemoteSafeWithOptions did not resolve allowlisted host for private-address validation")
	}
}

func TestLoadRemoteSafeAllowsAllowedPublicHost(t *testing.T) {
	client := &http.Client{Transport: confRoundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("app:\n  name: remote")),
			Request:    r,
		}, nil
	})}
	c, err := LoadRemoteSafeWithOptions("http://config.example/app.yaml", LoadOptions{
		RemoteClient:       client,
		RemoteAllowedHosts: []string{"config.example"},
		LookupIP: func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("93.184.216.34")}, nil
		},
	})
	if err != nil {
		t.Fatalf("LoadRemoteSafeWithOptions allowed public host: %v", err)
	}
	if got := c.GetByGroup("app", "name"); got != "remote" {
		t.Fatalf("remote app.name = %q", got)
	}
}

func TestLoadRemoteSafeRevalidatesHostAtRoundTrip(t *testing.T) {
	lookups := [][]net.IP{{net.ParseIP("93.184.216.34")}, {net.ParseIP("127.0.0.1")}}
	lookupCount := 0
	client := &http.Client{Transport: confRoundTripperFunc(func(*http.Request) (*http.Response, error) {
		t.Fatal("unsafe request reached base transport")
		return nil, nil
	})}
	_, err := LoadRemoteSafeWithOptions("http://example.com/app.yaml", LoadOptions{
		RemoteClient: client,
		LookupIP: func(context.Context, string) ([]net.IP, error) {
			if lookupCount >= len(lookups) {
				return lookups[len(lookups)-1], nil
			}
			ips := lookups[lookupCount]
			lookupCount++
			return ips, nil
		},
	})
	if err == nil {
		t.Fatal("LoadRemoteSafeWithOptions should reject a host that resolves private during RoundTrip")
	}
	if lookupCount != 2 {
		t.Fatalf("lookup count = %d, want 2", lookupCount)
	}
}

type confRoundTripperFunc func(*http.Request) (*http.Response, error)

func (f confRoundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
