package http

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	knifer "github.com/imajinyun/go-knifer"
)

// Covers the utility toolkit-http DownloadTest.

func TestDownloadString(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("plain"))
	}))
	defer srv.Close()

	if got := DownloadString(srv.URL, ""); got != "plain" {
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

	if got := DownloadStringWithOptions(srv.URL, "", WithHeader("X-Token", "secret")); got != "with-options" {
		t.Fatalf("body: %q", got)
	}
}

func TestDownloadBytes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte{0x01, 0x02, 0x03})
	}))
	defer srv.Close()

	got := DownloadBytes(srv.URL)
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

	got := DownloadBytesWithOptions(srv.URL, WithHeader("X-Mode", "bytes"))
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

func TestDownloadFileToFile(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("file-content"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	target := filepath.Join(dir, "out.txt")
	n, err := DownloadFile(srv.URL, target)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n != int64(len("file-content")) {
		t.Fatalf("size: %d", n)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(data) != "file-content" {
		t.Fatalf("content: %q", string(data))
	}
}

func TestDownloadFileWithOptions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Mode") != "file" {
			http.Error(w, "missing option header", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte("file-options"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	n, err := DownloadFileWithOptions(srv.URL, dir, []RequestOption{WithHeader("X-Mode", "file")}, WithSaveDefaultFilename("out.txt"))
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n != int64(len("file-options")) {
		t.Fatalf("size: %d", n)
	}
	data, err := os.ReadFile(filepath.Join(dir, "out.txt"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(data) != "file-options" {
		t.Fatalf("content: %q", string(data))
	}
}

func TestSaveAsStreamsWithoutCachingBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("x", 64*1024)))
	}))
	defer srv.Close()

	resp := Get(srv.URL).Execute()
	target := filepath.Join(t.TempDir(), "stream.bin")
	n, err := resp.SaveAs(target)
	if err != nil {
		t.Fatalf("SaveAs() error = %v", err)
	}
	if n != 64*1024 {
		t.Fatalf("SaveAs() wrote %d bytes, want %d", n, 64*1024)
	}
	if resp.body != nil {
		t.Fatalf("SaveAs() should stream to file without caching response body, cached %d bytes", len(resp.body))
	}
	if got := resp.Bytes(); got != nil {
		t.Fatalf("Bytes() after streamed SaveAs = %q, want nil", string(got))
	}
	if !errors.Is(resp.Err(), knifer.ErrCodeUnsupported) {
		t.Fatalf("Err() after Bytes() on consumed body = %v, want unsupported", resp.Err())
	}
}

func TestSaveAsUsesCachedBodyAfterBodyRead(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("cached"))
	}))
	defer srv.Close()

	resp := Get(srv.URL).Execute()
	if got := resp.Body(); got != "cached" {
		t.Fatalf("Body() = %q, want cached", got)
	}
	target := filepath.Join(t.TempDir(), "cached.txt")
	if _, err := resp.SaveAs(target); err != nil {
		t.Fatalf("SaveAs() after Body() error = %v", err)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "cached" {
		t.Fatalf("saved content = %q, want cached", data)
	}
}

func TestSaveAsOptions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("saved"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	target := filepath.Join(dir, "out.txt")
	if err := os.WriteFile(target, []byte("old"), 0o644); err != nil {
		t.Fatalf("write old: %v", err)
	}
	if _, err := Get(srv.URL).Execute().SaveAs(target, WithSaveOverwrite(false)); err == nil {
		t.Fatal("SaveAs overwrite false should fail")
	}
	missing := filepath.Join(dir, "missing", "out.txt")
	if _, err := Get(srv.URL).Execute().SaveAs(missing, WithSaveCreateParents(false)); err == nil {
		t.Fatal("SaveAs without parent creation should fail")
	}
	if _, err := DownloadFile(srv.URL, target); err != nil {
		t.Fatalf("DownloadFile overwrite default: %v", err)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(data) != "saved" {
		t.Fatalf("content = %q", data)
	}
}

func TestSaveAsProviderOptions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("provider-save"))
	}))
	defer srv.Close()

	var mkdirPath string
	var mkdirPerm fs.FileMode
	var openPath string
	var openFlag int
	var openPerm fs.FileMode
	var written bytes.Buffer
	n, err := Get(srv.URL).Execute().SaveAs("/virtual/out.txt",
		WithSaveMkdirAll(func(path string, perm fs.FileMode) error {
			mkdirPath, mkdirPerm = path, perm
			return nil
		}),
		WithSaveOpenFile(func(path string, flag int, perm fs.FileMode) (io.WriteCloser, error) {
			openPath, openFlag, openPerm = path, flag, perm
			return nopWriteCloser{Writer: &written}, nil
		}),
		WithSaveDirPerm(0o700), WithSaveFilePerm(0o600),
	)
	if err != nil || n != int64(len("provider-save")) {
		t.Fatalf("SaveAs provider n=%d err=%v", n, err)
	}
	if mkdirPath != "/virtual" || mkdirPerm != 0o700 || openPath != "/virtual/out.txt" || openPerm != 0o600 || openFlag&os.O_CREATE == 0 || written.String() != "provider-save" {
		t.Fatalf("providers mkdir=%q/%v open=%q flag=%#x perm=%v content=%q", mkdirPath, mkdirPerm, openPath, openFlag, openPerm, written.String())
	}
}

type nopWriteCloser struct{ io.Writer }

func (w nopWriteCloser) Close() error { return nil }

func TestSaveAsDefaultFilenameOption(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("fallback"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	if _, err := Get(srv.URL).Execute().SaveAs(dir, WithSaveDefaultFilename("fallback.bin")); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "fallback.bin")); err != nil {
		t.Fatalf("fallback file missing: %v", err)
	}
}

func TestDownloadFileToDirectory(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("D"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	// dest is a directory, so the file name should come from the URL path.
	url := srv.URL + "/foo.bin"
	if _, err := DownloadFile(url, dir); err != nil {
		t.Fatalf("err: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "foo.bin")); err != nil {
		t.Fatalf("file should exist: %v", err)
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

func TestSaveAsViaContentDisposition(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Disposition", `attachment; filename="real.bin"`)
		_, _ = w.Write([]byte("from-cd"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	resp := Get(srv.URL).Execute()
	if _, err := resp.SaveAs(dir); err != nil {
		t.Fatalf("err: %v", err)
	}
	target := filepath.Join(dir, "real.bin")
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("not found: %v", err)
	}
	if !strings.Contains(string(data), "from-cd") {
		t.Fatalf("content: %q", string(data))
	}
}

func TestSaveAsRejectsUnsafeContentDispositionFilename(t *testing.T) {
	tests := []string{
		`attachment; filename="../outside"`,
		`attachment; filename="..\\outside"`,
		`attachment; filename="/tmp/outside"`,
	}
	for _, cd := range tests {
		t.Run(cd, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Disposition", cd)
				_, _ = w.Write([]byte("unsafe"))
			}))
			defer srv.Close()

			dir := t.TempDir()
			_, err := Get(srv.URL).Execute().SaveAs(dir)
			if !errors.Is(err, knifer.ErrCodeInvalidInput) {
				t.Fatalf("SaveAs error = %v, want invalid input", err)
			}
			if _, statErr := os.Stat(filepath.Join(dir, "outside")); !errors.Is(statErr, os.ErrNotExist) {
				t.Fatalf("unsafe file should not be created, stat err = %v", statErr)
			}
		})
	}
}
