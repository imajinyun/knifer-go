package vconf_test

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	knifer "github.com/imajinyun/knifer-go"
	"github.com/imajinyun/knifer-go/vconf"
)

func TestFacadeRemoteSafeWrappers(t *testing.T) {
	trustedServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("trusted=ok\n"))
	}))
	defer trustedServer.Close()
	trusted, err := vconf.LoadRemote(trustedServer.URL + "/app.setting")
	if err != nil {
		t.Fatalf("LoadRemote() error = %v", err)
	}
	if got := trusted.Get("trusted"); got != "ok" {
		t.Fatalf("LoadRemote trusted = %q", got)
	}

	client := &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if got := req.Header.Get("X-Remote-Token"); got != "token" {
			t.Fatalf("remote header = %q, want token", got)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"text/plain"}},
			Body:       io.NopCloser(strings.NewReader("remote=ok\n")),
			Request:    req,
		}, nil
	})}
	lookupPublic := func(context.Context, string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("8.8.8.8")}, nil
	}
	opts := vconf.LoadOptions{
		RemoteClient:       client,
		Headers:            http.Header{"X-Remote-Token": []string{"token"}},
		RemoteAllowedHosts: []string{"config.example"},
		LookupIP:           lookupPublic,
		Timeout:            time.Second,
		MaxBytes:           64,
	}

	remote, err := vconf.LoadRemoteWithOptions("http://config.example/app.setting", opts)
	if err != nil {
		t.Fatalf("LoadRemoteWithOptions() error = %v", err)
	}
	if got := remote.Get("remote"); got != "ok" {
		t.Fatalf("LoadRemoteWithOptions remote = %q", got)
	}
	safe, err := vconf.LoadRemoteSafeWithOptions("http://config.example/app.setting", opts)
	if err != nil {
		t.Fatalf("LoadRemoteSafeWithOptions() error = %v", err)
	}
	if got := safe.Get("remote"); got != "ok" {
		t.Fatalf("LoadRemoteSafeWithOptions remote = %q", got)
	}
	if _, err := vconf.LoadRemoteSafe("http://127.0.0.1/app.setting"); err == nil {
		t.Fatal("LoadRemoteSafe private host error = nil")
	}
}

func TestFacadeLoadRemoteSafeRejectsBoundaryFailures(t *testing.T) {
	publicLookup := func(context.Context, string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("93.184.216.34")}, nil
	}

	t.Run("max bytes", func(t *testing.T) {
		client := &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("name=too-large\n")),
				Request:    req,
			}, nil
		})}
		_, err := vconf.LoadRemoteSafeWithOptions("http://config.example/app.setting", vconf.LoadOptions{
			RemoteClient:       client,
			RemoteAllowedHosts: []string{"config.example"},
			LookupIP:           publicLookup,
			MaxBytes:           4,
		})
		if !errors.Is(err, knifer.ErrCodeInvalidInput) {
			t.Fatalf("LoadRemoteSafeWithOptions max bytes error = %v, want invalid input classification", err)
		}
	})

	t.Run("allowlisted private host", func(t *testing.T) {
		client := &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			t.Fatal("private host must be rejected before request execution")
			return nil, nil
		})}
		_, err := vconf.LoadRemoteSafeWithOptions("http://config.example/app.setting", vconf.LoadOptions{
			RemoteClient:       client,
			RemoteAllowedHosts: []string{"config.example"},
			LookupIP: func(context.Context, string) ([]net.IP, error) {
				return []net.IP{net.ParseIP("10.0.0.1")}, nil
			},
		})
		if !errors.Is(err, knifer.ErrCodeInvalidInput) {
			t.Fatalf("LoadRemoteSafeWithOptions private error = %v, want invalid input", err)
		}
	})

	t.Run("unsafe redirect", func(t *testing.T) {
		client := &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusFound,
				Header:     http.Header{"Location": []string{"http://127.0.0.1/app.setting"}},
				Body:       io.NopCloser(strings.NewReader("")),
				Request:    req,
			}, nil
		})}
		_, err := vconf.LoadRemoteSafeWithOptions("http://config.example/app.setting", vconf.LoadOptions{
			RemoteClient:       client,
			RemoteAllowedHosts: []string{"config.example"},
			LookupIP:           publicLookup,
		})
		if !errors.Is(err, knifer.ErrCodeInvalidInput) {
			t.Fatalf("LoadRemoteSafeWithOptions redirect error = %v, want invalid input classification", err)
		}
	})

	t.Run("non 2xx status", func(t *testing.T) {
		client := &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusTeapot,
				Body:       io.NopCloser(strings.NewReader("nope")),
				Request:    req,
			}, nil
		})}
		_, err := vconf.LoadRemoteSafeWithOptions("http://config.example/app.setting", vconf.LoadOptions{
			RemoteClient:       client,
			RemoteAllowedHosts: []string{"config.example"},
			LookupIP:           publicLookup,
		})
		if !errors.Is(err, knifer.ErrCodeInvalidInput) {
			t.Fatalf("LoadRemoteSafeWithOptions status error = %v, want invalid input", err)
		}
	})

	t.Run("redacts url query in error", func(t *testing.T) {
		secret := "sk-test-secret"
		client := &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusTeapot,
				Body:       io.NopCloser(strings.NewReader("nope")),
				Request:    req,
			}, nil
		})}
		_, err := vconf.LoadRemoteSafeWithOptions("http://config.example/app.setting?token="+secret, vconf.LoadOptions{
			RemoteClient:       client,
			RemoteAllowedHosts: []string{"config.example"},
			LookupIP:           publicLookup,
		})
		if !errors.Is(err, knifer.ErrCodeInvalidInput) {
			t.Fatalf("LoadRemoteSafeWithOptions query error = %v, want invalid input", err)
		}
		if strings.Contains(err.Error(), secret) || strings.Contains(err.Error(), "token=") {
			t.Fatalf("LoadRemoteSafeWithOptions error leaked query secret: %v", err)
		}
	})
}
