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

	knifer "github.com/imajinyun/knifer-go"
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
	for _, tt := range []struct {
		name  string
		level int
	}{
		{name: "default", level: gzip.DefaultCompression},
		{name: "stored_blocks", level: gzip.NoCompression},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var encoded bytes.Buffer
			gz, err := gzip.NewWriterLevel(&encoded, tt.level)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := gz.Write([]byte("gzipped")); err != nil {
				t.Fatal(err)
			}
			if err := gz.Close(); err != nil {
				t.Fatal(err)
			}
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Encoding", "gzip")
				_, _ = w.Write(encoded.Bytes())
			}))
			defer srv.Close()

			data := Get(srv.URL, WithAutoDecodeResponse(false)).Execute().Bytes()
			if !bytes.Equal(data, encoded.Bytes()) {
				t.Fatalf("raw body = %q, want original gzip bytes %q", data, encoded.Bytes())
			}
			reader, err := gzip.NewReader(bytes.NewReader(data))
			if err != nil {
				t.Fatal(err)
			}
			defer reader.Close()
			decoded, err := io.ReadAll(reader)
			if err != nil || string(decoded) != "gzipped" {
				t.Fatalf("decoded body = %q, err = %v", decoded, err)
			}
		})
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
