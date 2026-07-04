package url

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResourceWrapperAliasesAndDefaultProviders(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.txt")
	if err := os.WriteFile(path, []byte("hello"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	fileURL, err := FileURL(path)
	if err != nil {
		t.Fatalf("FileURL: %v", err)
	}

	r, err := Open(fileURL.String())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	data, err := io.ReadAll(r)
	_ = r.Close()
	if err != nil || string(data) != "hello" {
		t.Fatalf("Open body = %q, %v", data, err)
	}
	if length, err := ContentLength(path); err != nil || length != 5 {
		t.Fatalf("ContentLength = %d, %v", length, err)
	}
	if size, err := Size(path); err != nil || size != 5 {
		t.Fatalf("Size = %d, %v", size, err)
	}
	if size, err := SizeWithOptions(path); err != nil || size != 5 {
		t.Fatalf("SizeWithOptions = %d, %v", size, err)
	}

	defaultReader, err := defaultOpenFile(path)
	if err != nil {
		t.Fatalf("defaultOpenFile: %v", err)
	}
	_ = defaultReader.Close()
	if ips, err := defaultLookupIP(context.Background(), "127.0.0.1"); err != nil || len(ips) == 0 {
		t.Fatalf("defaultLookupIP = %#v, %v", ips, err)
	}
}

func TestSafeContentLengthAndHeadersWithInjectedClient(t *testing.T) {
	ctx := context.WithValue(context.Background(), contextKeyForURLTest{}, "marker")
	client := &http.Client{Transport: urlRoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if req.Context().Value(contextKeyForURLTest{}) != "marker" {
			t.Fatalf("request context missing marker")
		}
		if got := req.Header.Values("X-Token"); len(got) != 2 || got[0] != "a" || got[1] != "b" {
			t.Fatalf("headers = %#v", req.Header)
		}
		return &http.Response{
			StatusCode:    http.StatusOK,
			ContentLength: 12,
			Body:          io.NopCloser(strings.NewReader("")),
			Header:        make(http.Header),
			Request:       req,
		}, nil
	})}
	length, err := ContentLengthSafeWithOptions("http://example.com/data",
		WithContext(ctx),
		WithHTTPClient(client),
		WithHeaders(http.Header{"X-Token": []string{"a", "b"}}),
		WithAllowedHosts("example.com"),
		WithLookupIP(func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("93.184.216.34")}, nil
		}),
	)
	if err != nil || length != 12 {
		t.Fatalf("ContentLengthSafeWithOptions = %d, %v", length, err)
	}
	if _, err := ContentLengthSafe("file:///tmp/secret.txt"); err == nil {
		t.Fatal("ContentLengthSafe should reject file URLs")
	}
}

func TestResourceOptionFallbacksAndSafeDialValidation(t *testing.T) {
	var nilCtx context.Context
	cfg := applyResourceOptions([]ResourceOption{
		WithContext(nilCtx),
		WithHTTPClient(nil),
		WithOpenFile(nil),
		WithStat(nil),
		WithRequestFactory(nil),
		WithLookupIP(nil),
		WithHeaders(nil),
	})
	if cfg.ctx == nil || cfg.client == nil || cfg.openFile == nil || cfg.stat == nil || cfg.requestFactory == nil || cfg.lookupIP == nil {
		t.Fatalf("applyResourceOptions did not restore defaults: %#v", cfg)
	}

	client := &http.Client{}
	openFile := func(string) (io.ReadCloser, error) { return io.NopCloser(strings.NewReader("")), nil }
	stat := func(string) (os.FileInfo, error) { return nil, os.ErrNotExist }
	requestFactory := func(ctx context.Context, method, raw string) (*http.Request, error) {
		return http.NewRequestWithContext(ctx, method, raw, nil)
	}
	lookupIP := func(context.Context, string) ([]net.IP, error) { return []net.IP{net.ParseIP("93.184.216.34")}, nil }
	cfg = applyResourceOptions([]ResourceOption{
		WithHTTPClient(client), WithHTTPClient(nil),
		WithOpenFile(openFile), WithOpenFile(nil),
		WithStat(stat), WithStat(nil),
		WithRequestFactory(requestFactory), WithRequestFactory(nil),
		WithLookupIP(lookupIP), WithLookupIP(nil),
	})
	if cfg.client != client || cfg.openFile == nil || cfg.stat == nil || cfg.requestFactory == nil || cfg.lookupIP == nil {
		t.Fatalf("nil provider option overwrote configured provider: %#v", cfg)
	}

	dial := safeDialContext(resourceConfig{})
	if _, err := dial(context.Background(), "tcp", "missing-port"); err == nil {
		t.Fatal("safeDialContext missing port error = nil")
	}
	if _, err := dial(context.Background(), "tcp", ":80"); err == nil {
		t.Fatal("safeDialContext blank host error = nil")
	}

	public, err := publicHostIPs(nilCtx, resourceConfig{}, "93.184.216.34")
	if err != nil || len(public) != 1 || public[0].String() != "93.184.216.34" {
		t.Fatalf("publicHostIPs direct = %#v, %v", public, err)
	}
	if _, err := publicHostIPs(context.Background(), resourceConfig{}, "localhost"); err == nil {
		t.Fatal("publicHostIPs localhost error = nil")
	}
	if _, err := publicHostIPs(context.Background(), resourceConfig{lookupIP: func(context.Context, string) ([]net.IP, error) {
		return nil, nil
	}}, "example.com"); err == nil {
		t.Fatal("publicHostIPs empty lookup error = nil")
	}
	wantErr := errors.New("lookup failed")
	if _, err := publicHostIPs(context.Background(), resourceConfig{lookupIP: func(context.Context, string) ([]net.IP, error) {
		return nil, wantErr
	}}, "example.com"); !errors.Is(err, wantErr) {
		t.Fatalf("publicHostIPs lookup error = %v", err)
	}
}

type contextKeyForURLTest struct{}
