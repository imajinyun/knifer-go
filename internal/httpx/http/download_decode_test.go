package http

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	knifer "github.com/imajinyun/go-knifer"
)

func TestSaveAsHonorsMaxResponseBytesAfterDecode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		_, _ = gz.Write([]byte("too-large-after-decode"))
		_ = gz.Close()
	}))
	defer srv.Close()

	target := filepath.Join(t.TempDir(), "limited-gzip.txt")
	_, err := Get(srv.URL, WithMaxResponseBytes(4)).Execute().SaveAs(target)
	if !errors.Is(err, knifer.ErrCodeUnsupported) {
		t.Fatalf("SaveAs() gzip error = %v, want unsupported", err)
	}
}

func TestDownloadGzipDecode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		_, _ = gz.Write([]byte("gzipped"))
		_ = gz.Close()
	}))
	defer srv.Close()

	body := Get(srv.URL).Execute().Body()
	if body != "gzipped" {
		t.Fatalf("decoded body: %q", body)
	}
}

func TestDownloadGzipDecodeCanBeDisabled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		_, _ = gz.Write([]byte("gzipped"))
		_ = gz.Close()
	}))
	defer srv.Close()

	data := Get(srv.URL, WithAutoDecodeResponse(false)).Execute().Bytes()
	if bytes.Contains(data, []byte("gzipped")) || len(data) == 0 {
		t.Fatalf("body should remain compressed, got %q", data)
	}
}

func TestDownloadCustomContentDecoder(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "reverse")
		_, _ = w.Write([]byte("olleh"))
	}))
	defer srv.Close()

	decoder := func(r io.Reader) (io.ReadCloser, error) {
		data, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}
		for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
			data[i], data[j] = data[j], data[i]
		}
		return io.NopCloser(bytes.NewReader(data)), nil
	}
	body := Get(srv.URL, WithContentDecoder("reverse", decoder)).Execute().Body()
	if body != "hello" {
		t.Fatalf("custom decoded body: %q", body)
	}
}
