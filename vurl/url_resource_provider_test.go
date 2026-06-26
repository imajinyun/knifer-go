package vurl_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/imajinyun/knifer-go/vurl"
)

func TestFacadeResourceProviderOptions(t *testing.T) {
	openedPath := ""
	rc, err := vurl.OpenWithOptions("file:///virtual/data.txt", vurl.WithOpenFile(func(path string) (io.ReadCloser, error) {
		openedPath = path
		return io.NopCloser(strings.NewReader("facade-file")), nil
	}))
	if err != nil {
		t.Fatalf("OpenWithOptions custom open: %v", err)
	}
	data, err := io.ReadAll(rc)
	_ = rc.Close()
	if err != nil || string(data) != "facade-file" || openedPath != "/virtual/data.txt" {
		t.Fatalf("custom open data=%q path=%q err=%v", data, openedPath, err)
	}

	statSource := t.TempDir() + "/stat.txt"
	if err := os.WriteFile(statSource, []byte("12345"), 0o600); err != nil {
		t.Fatal(err)
	}
	statPath := ""
	length, err := vurl.ContentLengthWithOptions("/virtual/stat.txt", vurl.WithStat(func(path string) (os.FileInfo, error) {
		statPath = path
		return os.Stat(statSource)
	}))
	if err != nil || length != 5 || statPath != "/virtual/stat.txt" {
		t.Fatalf("custom stat length=%d path=%q err=%v", length, statPath, err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Factory") != "facade" {
			http.Error(w, "missing factory header", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte("factory"))
	}))
	defer server.Close()
	method := ""
	_, err = vurl.ContentLengthWithOptions(server.URL, vurl.WithRequestFactory(func(ctx context.Context, m, raw string) (*http.Request, error) {
		method = m
		req, err := http.NewRequestWithContext(ctx, m, raw, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("X-Factory", "facade")
		return req, nil
	}), vurl.WithCheckStatus(true))
	if err != nil || method != http.MethodHead {
		t.Fatalf("request factory method=%q err=%v", method, err)
	}
}
