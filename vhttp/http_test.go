package vhttp_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	knifer "github.com/imajinyun/go-knifer"
	"github.com/imajinyun/go-knifer/vhttp"
	"github.com/imajinyun/go-knifer/vurl"
)

func TestFacadeUsesNamesWithoutHTTPPrefix(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != string(vhttp.MethodGet) {
			t.Fatalf("method = %q, want GET", r.Method)
		}
		if got := r.URL.Query().Get("lang"); got != "go" {
			t.Fatalf("query lang = %q, want go", got)
		}
		if got := r.Header.Get("X-Client"); got != "go-knifer" {
			t.Fatalf("header X-Client = %q, want go-knifer", got)
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	req := vhttp.Get(server.URL).
		Query("lang", "go").
		Header("X-Client", "go-knifer")

	resp := executeRequest(req)
	if resp.Err() != nil {
		t.Fatal(resp.Err())
	}
	if got := resp.Body(); got != "ok" {
		t.Fatalf("Body() = %q, want ok", got)
	}
}

func TestFacadeRequestOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("X-Opt") + ":" + r.Header.Get("User-Agent")))
	}))
	defer server.Close()

	resp := vhttp.Get(
		server.URL,
		vhttp.WithHeader("X-Opt", "yes"),
		vhttp.WithUserAgent("vhttp-test/1.0"),
	).Execute()
	if resp.Err() != nil {
		t.Fatal(resp.Err())
	}
	if got := resp.Body(); got != "yes:vhttp-test/1.0" {
		t.Fatalf("Body() = %q, want option headers", got)
	}
}

func TestFacadeResponseDecodeOptions(t *testing.T) {
	gzipServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		_, _ = gz.Write([]byte("gzipped"))
		_ = gz.Close()
	}))
	defer gzipServer.Close()

	compressed := vhttp.Get(gzipServer.URL, vhttp.WithAutoDecodeResponse(false)).Execute().Bytes()
	if bytes.Contains(compressed, []byte("gzipped")) || len(compressed) == 0 {
		t.Fatalf("body should remain compressed, got %q", compressed)
	}

	customServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "upper")
		_, _ = w.Write([]byte("hello"))
	}))
	defer customServer.Close()

	decoder := func(r io.Reader) (io.ReadCloser, error) {
		data, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}
		return io.NopCloser(strings.NewReader(strings.ToUpper(string(data)))), nil
	}
	if got := vhttp.Get(customServer.URL, vhttp.WithContentDecoder("upper", decoder)).Execute().Body(); got != "HELLO" {
		t.Fatalf("custom decoded body = %q", got)
	}
}

func TestFacadeSimpleServerOptions(t *testing.T) {
	server := vhttp.NewSimpleServerAddrWithOptions("127.0.0.1:0",
		vhttp.WithReadHeaderTimeout(time.Second),
		vhttp.WithReadTimeout(time.Second),
		vhttp.WithWriteTimeout(time.Second),
		vhttp.WithIdleTimeout(time.Second),
		vhttp.WithHTTPServer(&http.Server{Addr: "127.0.0.1:0"}),
	)
	if server == nil {
		t.Fatal("NewSimpleServerAddrWithOptions returned nil")
	}
	if err := server.StopWithContext(context.Background()); err != nil {
		t.Fatalf("StopWithContext on idle server = %v", err)
	}
}

func TestFacadeHelperNamesWithoutHTTPPrefix(t *testing.T) {
	vhttp.SetGlobalTimeout(2 * time.Second)
	if got := vhttp.GetGlobalTimeout(); got != 2*time.Second {
		t.Fatalf("GetGlobalTimeout() = %v, want 2s", got)
	}

	vhttp.SetGlobalHeader("X-Test", "a")
	vhttp.AddGlobalHeader("X-Test", "b")
	if got := vhttp.CloneGlobalHeaders().Values("X-Test"); len(got) != 2 {
		t.Fatalf("CloneGlobalHeaders().Values(X-Test) = %v, want 2 values", got)
	}
	vhttp.RemoveGlobalHeader("X-Test")
	if got := vhttp.CloneGlobalHeaders().Values("X-Test"); len(got) != 0 {
		t.Fatalf("after RemoveGlobalHeader values = %v, want empty", got)
	}

	if got := vhttp.BuildBasicAuth("aladdin", "opensesame"); got != "Basic YWxhZGRpbjpvcGVuc2VzYW1l" {
		t.Fatalf("BuildBasicAuth() = %q", got)
	}
	if got := vurl.EncodeQueryMap(map[string]any{"q": "go", "page": 1}); !strings.Contains(got, "q=go") || !strings.Contains(got, "page=1") {
		t.Fatalf("EncodeQueryMap() = %q", got)
	}
}

func TestFacadeErrorNamesWithoutHTTPPrefix(t *testing.T) {
	cause := errors.New("closed")
	err := vhttp.NewError("read failed", cause)
	if !errors.Is(err, cause) {
		t.Fatalf("NewError() does not unwrap cause")
	}
	if !errors.Is(err, knifer.ErrCodeInternal) {
		t.Fatalf("NewError() does not match ErrCodeInternal")
	}
	code, ok := knifer.CodeOf(err)
	if !ok || code != knifer.ErrCodeInternal {
		t.Fatalf("CodeOf(NewError()) = %q, %v; want internal", code, ok)
	}

	formatted := vhttp.Errorf("status %d", 500)
	if got := errorString(formatted); got != "status 500" {
		t.Fatalf("Errorf().Error() = %q, want status 500", got)
	}
}

func executeRequest(req *vhttp.Request) *vhttp.Response {
	return req.Execute()
}

func errorString(err *vhttp.Error) string {
	return err.Error()
}
