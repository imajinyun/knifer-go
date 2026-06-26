package vresty_test

import (
	"errors"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	knifer "github.com/imajinyun/knifer-go"
	"github.com/imajinyun/knifer-go/vresty"
)

func TestFacadeBuildBasicAuth(t *testing.T) {
	if got := vresty.BuildBasicAuth("u", "p"); got != "Basic dTpw" {
		t.Fatalf("BuildBasicAuth() = %q, want Basic dTpw", got)
	}
}

func TestFacadeUtilityWrappers(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			body, _ := io.ReadAll(r.Body)
			_, _ = w.Write([]byte("post:" + string(body) + ":" + r.Header.Get("X-Util")))
			return
		}
		_, _ = w.Write([]byte(r.URL.Query().Get("q") + ":" + r.Header.Get("X-Util")))
	}))
	defer srv.Close()

	got, err := vresty.GetWithParamsEWithOptions(srv.URL, map[string]any{"q": "go"}, vresty.WithHeader("X-Util", "get"))
	if err != nil {
		t.Fatalf("GetWithParamsEWithOptions() error = %v", err)
	}
	if got != "go:get" {
		t.Fatalf("GetWithParamsEWithOptions() = %q, want go:get", got)
	}
	got, err = vresty.PostStringEWithOptions(srv.URL, "body", vresty.WithHeader("X-Util", "post"))
	if err != nil {
		t.Fatalf("PostStringEWithOptions() error = %v", err)
	}
	if got != "post:body:post" {
		t.Fatalf("PostStringEWithOptions() = %q, want post:body:post", got)
	}
	if !vresty.IsHTTP("http://example.com") || !vresty.IsHTTPS("https://example.com") {
		t.Fatal("IsHTTP/IsHTTPS wrappers returned false")
	}
	if got := vresty.ToParams(map[string]any{"q": "go"}); got != "q=go" {
		t.Fatalf("ToParams() = %q, want q=go", got)
	}
}

func TestFacadeErrorsCharsetAndSaveOptions(t *testing.T) {
	cause := errors.New("network closed")
	err := vresty.NewHTTPError("request failed", cause)
	if !errors.Is(err, cause) || !errors.Is(err, knifer.ErrCodeInternal) {
		t.Fatalf("NewHTTPError does not unwrap cause or code: %v", err)
	}
	if got := vresty.HTTPErrorf("status %d", http.StatusBadGateway).Error(); got != "status 502" {
		t.Fatalf("HTTPErrorf = %q", got)
	}
	if got := vresty.GetCharsetFromContentType("text/plain; charset=gb18030"); got != "gb18030" {
		t.Fatalf("GetCharsetFromContentType = %q", got)
	}
	if got := vresty.GetCharsetFromHTML(`<meta charset="utf-8">`); got != "utf-8" {
		t.Fatalf("GetCharsetFromHTML = %q", got)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("stat"))
	}))
	defer server.Close()

	info := fakeFileInfo{isDir: true}
	statCalled := false
	if _, err := vresty.Get(server.URL).Execute().SaveAs("/virtual-dir",
		vresty.WithSaveStat(func(path string) (os.FileInfo, error) {
			statCalled = path == "/virtual-dir"
			return info, nil
		}),
		vresty.WithSaveDefaultFilename("fallback.txt"),
		vresty.WithSaveCreateParents(false),
		vresty.WithSaveOpenFile(func(string, int, fs.FileMode) (io.WriteCloser, error) {
			return nopWriteCloser{Writer: io.Discard}, nil
		}),
	); err != nil {
		t.Fatalf("SaveAs with stat provider: %v", err)
	}
	if !statCalled {
		t.Fatal("WithSaveStat provider was not called")
	}
}

type fakeFileInfo struct{ isDir bool }

func (f fakeFileInfo) Name() string { return "fake" }

func (f fakeFileInfo) Size() int64 { return 0 }

func (f fakeFileInfo) Mode() fs.FileMode { return fs.ModeDir }

func (f fakeFileInfo) ModTime() time.Time { return time.Time{} }

func (f fakeFileInfo) IsDir() bool { return f.isDir }

func (f fakeFileInfo) Sys() any { return nil }
