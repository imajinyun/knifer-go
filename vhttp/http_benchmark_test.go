package vhttp_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/imajinyun/knifer-go/vhttp"
)

type benchmarkHTTPPayload struct {
	OK bool `json:"ok"`
}

func BenchmarkGetStringE(b *testing.B) {
	server := newBenchmarkHTTPServer(`{"ok":true}`)
	defer server.Close()

	b.ReportAllocs()
	for b.Loop() {
		body, err := vhttp.GetStringE(server.URL)
		if err != nil || body == "" {
			b.Fatalf("GetStringE body=%q err=%v", body, err)
		}
	}
}

func BenchmarkGetStringEJSONDecode(b *testing.B) {
	server := newBenchmarkHTTPServer(`{"ok":true}`)
	defer server.Close()

	b.ReportAllocs()
	for b.Loop() {
		body, err := vhttp.GetStringE(server.URL)
		if err != nil {
			b.Fatalf("GetStringE: %v", err)
		}
		var payload benchmarkHTTPPayload
		if err := json.Unmarshal([]byte(body), &payload); err != nil || !payload.OK {
			b.Fatalf("json decode payload=%+v err=%v", payload, err)
		}
	}
}

func BenchmarkDownloadBytesEWithOptionsBounded(b *testing.B) {
	server := newBenchmarkHTTPServer(`{"ok":true}`)
	defer server.Close()

	b.ReportAllocs()
	for b.Loop() {
		body, err := vhttp.DownloadBytesEWithOptions(server.URL, vhttp.WithMaxResponseBytes(64))
		if err != nil || len(body) == 0 {
			b.Fatalf("DownloadBytesEWithOptions len=%d err=%v", len(body), err)
		}
	}
}

func BenchmarkGetStringSafeE(b *testing.B) {
	server := newBenchmarkHTTPServer(`{"ok":true}`)
	defer server.Close()

	policy := vhttp.URLPolicy{AllowedSchemes: []string{"http"}, RejectPrivate: false}
	b.ReportAllocs()
	for b.Loop() {
		body, err := vhttp.GetStringSafeE(server.URL, vhttp.WithURLPolicy(policy))
		if err != nil || body == "" {
			b.Fatalf("GetStringSafeE body=%q err=%v", body, err)
		}
	}
}

func BenchmarkDownloadFileSafeWithOptions(b *testing.B) {
	server := newBenchmarkHTTPServer(`safe-file`)
	defer server.Close()

	dest := filepath.Join(b.TempDir(), "download.txt")
	policy := vhttp.URLPolicy{AllowedSchemes: []string{"http"}, RejectPrivate: false}
	b.ReportAllocs()
	for b.Loop() {
		n, err := vhttp.DownloadFileSafeWithOptions(
			server.URL,
			dest,
			[]vhttp.RequestOption{vhttp.WithURLPolicy(policy)},
			vhttp.WithSaveOverwrite(true),
		)
		if err != nil || n == 0 {
			b.Fatalf("DownloadFileSafeWithOptions n=%d err=%v", n, err)
		}
	}
	_ = os.Remove(dest)
}

func newBenchmarkHTTPServer(body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
}
