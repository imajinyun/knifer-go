package http

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Covers the utility toolkit-http DownloadTest.

func TestDownloadString(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("plain"))
	}))
	defer srv.Close()

	got, err := DownloadStringE(srv.URL, "")
	if err != nil {
		t.Fatalf("DownloadStringE() error = %v", err)
	}
	if got != "plain" {
		t.Fatalf("body: %q", got)
	}
}

func TestDownloadStringWithOptions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Token") != "secret" {
			http.Error(w, "missing option header", http.StatusTeapot)
			return
		}
		_, _ = w.Write([]byte("with-options"))
	}))
	defer srv.Close()

	got, err := DownloadStringEWithOptions(srv.URL, "", WithHeader("X-Token", "secret"))
	if err != nil {
		t.Fatalf("DownloadStringEWithOptions() error = %v", err)
	}
	if got != "with-options" {
		t.Fatalf("body: %q", got)
	}
}

func TestDownloadBytes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte{0x01, 0x02, 0x03})
	}))
	defer srv.Close()

	got, err := DownloadBytesE(srv.URL)
	if err != nil {
		t.Fatalf("DownloadBytesE() error = %v", err)
	}
	if !bytes.Equal(got, []byte{0x01, 0x02, 0x03}) {
		t.Fatalf("bytes: %v", got)
	}
}

func TestDownloadBytesWithOptions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Mode") != "bytes" {
			http.Error(w, "missing option header", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte{0x04, 0x05, 0x06})
	}))
	defer srv.Close()

	got, err := DownloadBytesEWithOptions(srv.URL, WithHeader("X-Mode", "bytes"))
	if err != nil {
		t.Fatalf("DownloadBytesEWithOptions() error = %v", err)
	}
	if !bytes.Equal(got, []byte{0x04, 0x05, 0x06}) {
		t.Fatalf("bytes: %v", got)
	}
}

func TestDownloadToWriter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("write-me"))
	}))
	defer srv.Close()

	buf := &bytes.Buffer{}
	n, err := Download(srv.URL, buf)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n != int64(len("write-me")) || buf.String() != "write-me" {
		t.Fatalf("got %d bytes %q", n, buf.String())
	}
}

func TestDownloadWithOptions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Mode") != "writer" {
			http.Error(w, "missing option header", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte("write-options"))
	}))
	defer srv.Close()

	buf := &bytes.Buffer{}
	n, err := DownloadWithOptions(srv.URL, buf, WithHeader("X-Mode", "writer"))
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n != int64(len("write-options")) || buf.String() != "write-options" {
		t.Fatalf("got %d bytes %q", n, buf.String())
	}
}
