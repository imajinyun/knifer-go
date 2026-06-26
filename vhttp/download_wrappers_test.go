package vhttp_test

import (
	"bytes"
	"testing"

	"github.com/imajinyun/knifer-go/vhttp"
)

func TestFacadeDownloadWriterWrappers(t *testing.T) {
	server := newFacadeDownloadServer(t)
	defer server.Close()

	var buf bytes.Buffer
	if n, err := vhttp.Download(server.URL, &buf); err != nil || n != int64(len(facadeDownloadText)) || buf.String() != facadeDownloadText {
		t.Fatalf("Download n=%d body=%q err=%v", n, buf.String(), err)
	}
	buf.Reset()
	if n, err := vhttp.DownloadWithOptions(server.URL, &buf, vhttp.WithMaxResponseBytes(64)); err != nil || n != int64(len(facadeDownloadText)) || buf.String() != facadeDownloadText {
		t.Fatalf("DownloadWithOptions n=%d body=%q err=%v", n, buf.String(), err)
	}
	buf.Reset()
	if n, err := vhttp.DownloadSafe(server.URL, &buf, allowLocalURLPolicy()); err != nil || n != int64(len(facadeDownloadText)) || buf.String() != facadeDownloadText {
		t.Fatalf("DownloadSafe n=%d body=%q err=%v", n, buf.String(), err)
	}
}

func TestFacadeDownloadBytesWrappers(t *testing.T) {
	server := newFacadeDownloadServer(t)
	defer server.Close()

	if b, err := vhttp.DownloadBytesE(server.URL); err != nil || string(b) != facadeDownloadText {
		t.Fatalf("DownloadBytesE = %q, %v", b, err)
	}
	if b, err := vhttp.DownloadBytesEWithOptions(server.URL, vhttp.WithMaxResponseBytes(64)); err != nil || string(b) != facadeDownloadText {
		t.Fatalf("DownloadBytesEWithOptions = %q, %v", b, err)
	}
	if b, err := vhttp.DownloadBytesSafeE(server.URL, allowLocalURLPolicy()); err != nil || string(b) != facadeDownloadText {
		t.Fatalf("DownloadBytesSafeE = %q, %v", b, err)
	}
}

func TestFacadeDownloadStringWrappers(t *testing.T) {
	server := newFacadeDownloadServer(t)
	defer server.Close()

	if got, err := vhttp.DownloadStringE(server.URL, ""); err != nil || got != facadeDownloadText {
		t.Fatalf("DownloadStringE = %q, %v", got, err)
	}
	if got, err := vhttp.DownloadStringEWithOptions(server.URL, "", vhttp.WithMaxResponseBytes(64)); err != nil || got != facadeDownloadText {
		t.Fatalf("DownloadStringEWithOptions = %q, %v", got, err)
	}
	if got, err := vhttp.DownloadStringSafeE(server.URL, "", allowLocalURLPolicy()); err != nil || got != facadeDownloadText {
		t.Fatalf("DownloadStringSafeE = %q, %v", got, err)
	}
}
