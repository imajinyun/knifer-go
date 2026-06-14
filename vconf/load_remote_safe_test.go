package vconf_test

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/imajinyun/go-knifer/vconf"
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
