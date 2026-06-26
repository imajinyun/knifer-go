package vresty_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/imajinyun/knifer-go/vresty"
)

type benchmarkRestyPayload struct {
	OK bool `json:"ok"`
}

func BenchmarkGetStringE(b *testing.B) {
	server := newBenchmarkRestyServer(`{"ok":true}`)
	defer server.Close()

	b.ReportAllocs()
	for b.Loop() {
		body, err := vresty.GetStringE(server.URL)
		if err != nil || body == "" {
			b.Fatalf("GetStringE body=%q err=%v", body, err)
		}
	}
}

func BenchmarkResponseResultJSONDecode(b *testing.B) {
	server := newBenchmarkRestyServer(`{"ok":true}`)
	defer server.Close()

	b.ReportAllocs()
	for b.Loop() {
		resp := vresty.Get(server.URL).Result(&benchmarkRestyPayload{}).Execute()
		result, ok := resp.Result().(*benchmarkRestyPayload)
		if resp.Err() != nil || !ok || !result.OK {
			b.Fatalf("Result payload=%+v ok=%v err=%v", result, ok, resp.Err())
		}
	}
}

func BenchmarkDownloadBytesEWithOptionsBounded(b *testing.B) {
	server := newBenchmarkRestyServer(`{"ok":true}`)
	defer server.Close()

	b.ReportAllocs()
	for b.Loop() {
		body, err := vresty.DownloadBytesEWithOptions(server.URL, vresty.WithMaxResponseBytes(64))
		if err != nil || len(body) == 0 {
			b.Fatalf("DownloadBytesEWithOptions len=%d err=%v", len(body), err)
		}
	}
}

func BenchmarkGetStringSafeE(b *testing.B) {
	server := newBenchmarkRestyServer(`{"ok":true}`)
	defer server.Close()

	policy := vresty.URLPolicy{AllowedSchemes: []string{"http"}, RejectPrivate: false}
	b.ReportAllocs()
	for b.Loop() {
		body, err := vresty.GetStringSafeE(server.URL, vresty.WithURLPolicy(policy))
		if err != nil || body == "" {
			b.Fatalf("GetStringSafeE body=%q err=%v", body, err)
		}
	}
}

func BenchmarkDownloadFileSafeWithOptions(b *testing.B) {
	server := newBenchmarkRestyServer(`safe-file`)
	defer server.Close()

	dest := filepath.Join(b.TempDir(), "download.txt")
	policy := vresty.URLPolicy{AllowedSchemes: []string{"http"}, RejectPrivate: false}
	b.ReportAllocs()
	for b.Loop() {
		n, err := vresty.DownloadFileSafeWithOptions(
			server.URL,
			dest,
			[]vresty.RequestOption{vresty.WithURLPolicy(policy)},
			vresty.WithSaveOverwrite(true),
		)
		if err != nil || n == 0 {
			b.Fatalf("DownloadFileSafeWithOptions n=%d err=%v", n, err)
		}
	}
	_ = os.Remove(dest)
}

func newBenchmarkRestyServer(body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
}
