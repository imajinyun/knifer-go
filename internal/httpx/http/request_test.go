package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"
)

// Covers the utility toolkit-http HttpRequestTest.

func TestRequestGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/hello" || r.URL.Query().Get("name") != "world" {
			http.Error(w, "bad", 400)
			return
		}
		w.Header().Set("Content-Type", "text/plain;charset=UTF-8")
		_, _ = w.Write([]byte("hi world"))
	}))
	defer srv.Close()

	resp := Get(srv.URL+"/hello").Query("name", "world").Execute()
	if resp.Err() != nil {
		t.Fatalf("err: %v", resp.Err())
	}
	if !resp.IsOK() {
		t.Fatalf("status: %d", resp.Status())
	}
	if got := resp.Body(); got != "hi world" {
		t.Fatalf("body: %q", got)
	}
	if cs := resp.Charset(); strings.ToUpper(cs) != "UTF-8" {
		t.Fatalf("charset: %q", cs)
	}
}

func TestRequestQueryMap(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.URL.Query().Get("a") + "," + r.URL.Query().Get("b")))
	}))
	defer srv.Close()

	body := Get(srv.URL).QueryMap(map[string]any{"a": 1, "b": "x"}).Execute().Body()
	if body != "1,x" {
		t.Fatalf("body: %q", body)
	}
}

func TestRequestPostForm(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		_, _ = w.Write([]byte(r.PostForm.Get("k")))
	}))
	defer srv.Close()

	body := Post(srv.URL).Form(map[string]any{"k": "v"}).Execute().Body()
	if body != "v" {
		t.Fatalf("body: %q", body)
	}
}

func TestRequestPostJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
			http.Error(w, "bad ct", 400)
			return
		}
		b, _ := io.ReadAll(r.Body)
		_, _ = w.Write(b)
	}))
	defer srv.Close()

	resp, err := PostJSONE(srv.URL, `{"a":1}`)
	if err != nil {
		t.Fatalf("PostJSONE() error = %v", err)
	}
	if resp != `{"a":1}` {
		t.Fatalf("body: %q", resp)
	}
}

func TestRequestBodyStringAutoContentType(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("Content-Type")))
	}))
	defer srv.Close()

	ct := Post(srv.URL).BodyString(`{"x":1}`).Execute().Body()
	if !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("expected json content-type detected, got %q", ct)
	}

	ct2 := Post(srv.URL).BodyString(`<x/>`).Execute().Body()
	if !strings.HasPrefix(ct2, "application/xml") {
		t.Fatalf("expected xml content-type detected, got %q", ct2)
	}
}

func TestRequestHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("X-Token")))
	}))
	defer srv.Close()

	body := Get(srv.URL).Header("X-Token", "abc").Execute().Body()
	if body != "abc" {
		t.Fatalf("body: %q", body)
	}
}

func TestAdditionalClientFactoriesAndSafeWrappers(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Method", r.Method)
		w.Header().Add("Set-Cookie", "sid=abc; Path=/")
		_, _ = w.Write([]byte(r.Method + ":" + r.Header.Get("X-Client")))
	}))
	defer srv.Close()

	cfg := SnapshotGlobalConfig()
	cfg.Headers.Set("X-Client", "cfg")
	client := NewClientWithConfig(cfg, WithHeader("X-Client", "opt"))
	if got := client.Get(srv.URL).Execute().Body(); got != "GET:opt" {
		t.Fatalf("client.Get body = %q", got)
	}
	if got := client.Post(srv.URL).Execute().Body(); got != "POST:opt" {
		t.Fatalf("client.Post body = %q", got)
	}
	if got := NewIsolatedClient(WithClientGlobalConfig(cfg), WithClientRequestOptions(WithHeader("X-Client", "isolated"))).NewRequest(MethodPut, srv.URL).Execute().Body(); got != "PUT:isolated" {
		t.Fatalf("NewIsolatedClient body = %q", got)
	}
	if got := (*Client)(nil).NewRequest(MethodDelete, srv.URL).Execute().Header("X-Method"); got != string(MethodDelete) {
		t.Fatalf("nil client NewRequest method = %q", got)
	}

	allowLocal := WithURLPolicy(URLPolicy{AllowedSchemes: []string{"http", "https"}, RejectPrivate: false})
	requests := []*HTTPRequest{
		PostSafe(srv.URL, allowLocal),
		Put(srv.URL),
		PutSafe(srv.URL, allowLocal),
		DeleteSafe(srv.URL, allowLocal),
		PatchSafe(srv.URL, allowLocal),
		Head(srv.URL),
		HeadSafe(srv.URL, allowLocal),
		Options(srv.URL),
		OptionsSafe(srv.URL, allowLocal),
		NewSafeRequest(MethodTrace, srv.URL, allowLocal),
		client.NewSafeRequest(MethodOptions, srv.URL, allowLocal),
	}
	for _, req := range requests {
		resp := req.Execute()
		if resp.Err() != nil {
			t.Fatalf("safe wrapper Execute: %v", resp.Err())
		}
		if resp.Status() == 0 {
			t.Fatal("safe wrapper status = 0")
		}
	}

	resp := Get(srv.URL).Cookie(&http.Cookie{Name: "k", Value: "v"}).Execute()
	if resp.Err() != nil {
		t.Fatalf("Get cookie Execute: %v", resp.Err())
	}
	if got := resp.Headers()["X-Method"]; len(got) != 1 || got[0] != http.MethodGet {
		t.Fatalf("Headers()[X-Method] = %v", got)
	}
	if cookies := resp.Cookies(); len(cookies) != 1 || cookies[0].Name != "sid" {
		t.Fatalf("Cookies = %#v", cookies)
	}
	var out bytes.Buffer
	if n, err := resp.WriteTo(&out); err != nil || n != int64(out.Len()) || !strings.Contains(out.String(), "GET") {
		t.Fatalf("WriteTo n=%d body=%q err=%v", n, out.String(), err)
	}
	if raw := resp.Raw(); raw == nil {
		t.Fatal("Raw returned nil")
	}
	if err := resp.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestAdditionalRequestOptionsAndAccessors(t *testing.T) {
	requestFactoryCalled := false
	readAllCalled := false
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode:    http.StatusOK,
			ContentLength: int64(len(req.Method + ":" + req.Header.Get("X-A"))),
			Header:        http.Header{"Content-Type": []string{"text/plain"}},
			Body:          io.NopCloser(strings.NewReader(req.Method + ":" + req.Header.Get("X-A"))),
			Request:       req,
		}, nil
	})
	cfg := SnapshotGlobalConfig()
	cfg.Headers.Set("X-A", "cfg")
	req := NewIsolatedRequest(MethodPost, "https://example.com/upload",
		WithGlobalConfig(cfg),
		WithHeaders(map[string]string{"X-A": "one", "X-B": "two"}),
		WithClient(&http.Client{Transport: transport}),
		WithTransport(transport),
		WithTLSConfig(&tls.Config{MinVersion: tls.VersionTLS12}),
		WithResponseReadAllFunc(func(r io.Reader) ([]byte, error) {
			readAllCalled = true
			return io.ReadAll(r)
		}),
		WithRequestFactory(func(method, rawURL string, body io.Reader) (*http.Request, error) {
			requestFactoryCalled = true
			return http.NewRequest(method, rawURL, body)
		}),
		WithMultipartWriterFactory(func(w io.Writer) MultipartWriter {
			return multipart.NewWriter(w)
		}),
	)
	req.Method(MethodPatch).URL("https://example.com/changed").AddHeader("X-A", "extra").Headers(map[string]string{"X-B": "two"}).CookieString("raw=cookie")
	req.Client(&http.Client{Transport: transport}).URLPolicy(URLPolicy{AllowedSchemes: []string{"https"}, RejectPrivate: false})
	resp := req.FormFileReader("file", "a.txt", strings.NewReader("file-data")).Execute()
	if resp.Err() != nil {
		t.Fatalf("multipart Execute: %v", resp.Err())
	}
	if got := resp.Body(); got != "PATCH:one" {
		t.Fatalf("response body = %q", got)
	}
	if !requestFactoryCalled || !readAllCalled {
		t.Fatalf("providers called request=%v readAll=%v", requestFactoryCalled, readAllCalled)
	}
	if got := NewRequestWithConfig(MethodGet, "https://example.com", cfg, WithTransport(transport)).Execute().Body(); got != "GET:cfg" {
		t.Fatalf("NewRequestWithConfig body = %q", got)
	}
}

func TestAdditionalGlobalHTMLAndUtilWrappers(t *testing.T) {
	previous := SnapshotGlobalConfig()
	defer ConfigureGlobalConfig(previous)
	SetGlobalBoundary("boundary-extra")
	SetGlobalDecodeURL(true)
	AddGlobalHeader("X-Extra", "one")
	defer RemoveGlobalHeader("X-Extra")
	if GetGlobalBoundary() != "boundary-extra" || !IsGlobalDecodeURL() {
		t.Fatalf("global boundary/decode = %q/%v", GetGlobalBoundary(), IsGlobalDecodeURL())
	}
	if got := CloneGlobalHeaders()["X-Extra"]; len(got) != 1 || got[0] != "one" {
		t.Fatalf("CloneGlobalHeaders X-Extra = %v", got)
	}
	if got := CleanHTMLWithOptions("a[drop]b", WithHTMLTagRegexp(regexp.MustCompile(`\[.*?\]`)), WithHTMLCommentRegexp(regexp.MustCompile(`$^`))); got != "ab" {
		t.Fatalf("CleanHTMLWithOptions = %q", got)
	}
	if got := FilterHTMLTagWithOptions("<custom>drop</custom><p>keep</p>", []string{"custom"}, WithHTMLFilterCompileFunc(regexp.Compile)); got != "<p>keep</p>" {
		t.Fatalf("FilterHTMLTagWithOptions = %q", got)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_, _ = w.Write([]byte(r.Method + ":" + string(body)))
	}))
	defer srv.Close()
	if got, err := GetStringSafeE(srv.URL, WithURLPolicy(URLPolicy{AllowedSchemes: []string{"http", "https"}, RejectPrivate: false})); err != nil || got != "GET:" {
		t.Fatalf("GetStringSafeE = %q, %v", got, err)
	}
	if got, err := GetWithTimeoutE(srv.URL, time.Second); err != nil || got != "GET:" {
		t.Fatalf("GetWithTimeoutE = %q, %v", got, err)
	}
	if got, err := PostStringE(srv.URL, "body"); err != nil || got != "POST:body" {
		t.Fatalf("PostStringE = %q, %v", got, err)
	}
	if got, err := PostStringSafeE(srv.URL, "safe", WithURLPolicy(URLPolicy{AllowedSchemes: []string{"http", "https"}, RejectPrivate: false})); err != nil || got != "POST:safe" {
		t.Fatalf("PostStringSafeE = %q, %v", got, err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestRequestBasicAuth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("Authorization")))
	}))
	defer srv.Close()

	body := Get(srv.URL).BasicAuth("aladdin", "opensesame").Execute().Body()
	if body != "Basic YWxhZGRpbjpvcGVuc2VzYW1l" {
		t.Fatalf("body: %q", body)
	}
}

func TestRequestBearerAuth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("Authorization")))
	}))
	defer srv.Close()

	body := Get(srv.URL).BearerAuth("xyz.token").Execute().Body()
	if body != "Bearer xyz.token" {
		t.Fatalf("body: %q", body)
	}
}

func TestRequestPatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Method))
	}))
	defer srv.Close()

	body := Patch(srv.URL).Execute().Body()
	if body != http.MethodPatch {
		t.Fatalf("method: %q", body)
	}
}

func TestRequestDelete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Method))
	}))
	defer srv.Close()

	body := Delete(srv.URL).Execute().Body()
	if body != http.MethodDelete {
		t.Fatalf("method: %q", body)
	}
}

func TestRequestTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	resp := Get(srv.URL).Timeout(50 * time.Millisecond).Execute()
	if resp.Err() == nil {
		t.Fatal("expected timeout error")
	}
}

func TestRequestNoFollowRedirects(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/start" {
			http.Redirect(w, r, "/end", http.StatusFound)
			return
		}
		_, _ = w.Write([]byte("end"))
	}))
	defer srv.Close()

	resp := Get(srv.URL + "/start").FollowRedirects(false).Execute()
	if resp.Status() != http.StatusFound {
		t.Fatalf("expected 302, got %d", resp.Status())
	}

	body := Get(srv.URL + "/start").FollowRedirects(true).Execute().Body()
	if body != "end" {
		t.Fatalf("redirected body: %q", body)
	}
}

func TestRequestMultipartUpload(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		f, fh, err := r.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		defer func() {
			if err := f.Close(); err != nil {
				t.Errorf("close multipart file: %v", err)
			}
		}()
		data, err := io.ReadAll(f)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		_, _ = w.Write([]byte(fh.Filename + ":" + string(data) + ":" + r.FormValue("k")))
	}))
	defer srv.Close()

	body := Post(srv.URL).
		Form(map[string]any{"k": "v"}).
		FormFile("file", "hello.txt", []byte("hi")).
		Execute().Body()
	if body != "hello.txt:hi:v" {
		t.Fatalf("body: %q", body)
	}
}

func TestRequestFormFileReader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		f, fh, err := r.FormFile("f")
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		defer func() {
			if err := f.Close(); err != nil {
				t.Errorf("close multipart file: %v", err)
			}
		}()
		data, err := io.ReadAll(f)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		_, _ = w.Write([]byte(fh.Filename + ":" + string(data)))
	}))
	defer srv.Close()

	body := Post(srv.URL).
		FormFileReader("f", "in.txt", strings.NewReader("hello reader")).
		Execute().Body()
	if body != "in.txt:hello reader" {
		t.Fatalf("body: %q", body)
	}
}

func TestRequestReaderBackedBodyIsSingleUse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		_, _ = w.Write(b)
	}))
	defer srv.Close()

	req := Post(srv.URL).BodyReader(strings.NewReader("hello"))
	if got := req.Execute().Body(); got != "hello" {
		t.Fatalf("first body = %q", got)
	}
	resp := req.Execute()
	if resp.Err() == nil {
		t.Fatal("second Execute() should reject reader-backed body reuse")
	}
}

func TestRequestCloneCreatesIndependentBuilder(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.URL.Query().Get("q") + ":" + r.Header.Get("X-Token")))
	}))
	defer srv.Close()

	base := Get(srv.URL).Query("q", "base").Header("X-Token", "base")
	clone := base.Clone().Query("q", "clone").Header("X-Token", "clone")

	if got := base.Execute().Body(); got != "base:base" {
		t.Fatalf("base Body() = %q", got)
	}
	if got := clone.Execute().Body(); got != "base:clone" {
		t.Fatalf("clone Body() = %q", got)
	}
}

func TestRequestCookie(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, _ := r.Cookie("k")
		if c == nil {
			_, _ = w.Write([]byte("no"))
			return
		}
		_, _ = w.Write([]byte(c.Value))
	}))
	defer srv.Close()

	body := Get(srv.URL).Cookie(&http.Cookie{Name: "k", Value: "v"}).Execute().Body()
	if body != "v" {
		t.Fatalf("body: %q", body)
	}
}

func TestRequestCustomTransport(t *testing.T) {
	rt := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		body := io.NopCloser(strings.NewReader("intercepted"))
		return &http.Response{
			StatusCode: 200,
			Body:       body,
			Header:     http.Header{"Content-Type": []string{"text/plain"}},
			Request:    req,
		}, nil
	})
	body := Get("http://will-not-call/").Transport(rt).Execute().Body()
	if body != "intercepted" {
		t.Fatalf("body: %q", body)
	}
}

func TestRequestOptionsOverrideGlobalDefaults(t *testing.T) {
	oldUA := GetGlobalUserAgent()
	oldFollow := GetGlobalFollowRedirects()
	defer SetGlobalUserAgent(oldUA)
	defer SetGlobalFollowRedirects(oldFollow)

	SetGlobalUserAgent("global-agent")
	SetGlobalFollowRedirects(false)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/start" {
			http.Redirect(w, r, "/end", http.StatusFound)
			return
		}
		_, _ = w.Write([]byte(r.Header.Get("X-Req") + ":" + r.Header.Get("User-Agent")))
	}))
	defer srv.Close()

	resp := Get(srv.URL+"/start",
		WithHeader("X-Req", "per-call"),
		WithUserAgent("request-agent"),
		WithFollowRedirects(true),
	).Execute()
	if resp.Err() != nil {
		t.Fatalf("Execute() error = %v", resp.Err())
	}
	if got := resp.Body(); got != "per-call:request-agent" {
		t.Fatalf("Body() = %q, want per-call options to override globals", got)
	}
}

func TestClientUsesCapturedConfig(t *testing.T) {
	oldUA := GetGlobalUserAgent()
	oldFollow := GetGlobalFollowRedirects()
	defer SetGlobalUserAgent(oldUA)
	defer SetGlobalFollowRedirects(oldFollow)

	SetGlobalUserAgent("client-agent")
	SetGlobalFollowRedirects(false)
	client := NewClient()
	SetGlobalUserAgent("mutated-agent")
	SetGlobalFollowRedirects(true)

	req := client.Get("https://example.com")
	if req.userAgent != "client-agent" {
		t.Fatalf("client request userAgent = %q, want captured client-agent", req.userAgent)
	}
	if req.followRedir == nil || *req.followRedir {
		t.Fatalf("client request followRedirects = %v, want captured false", req.followRedir)
	}

	isolated := NewIsolatedClient().Get("https://example.com")
	if isolated.userAgent != "" || isolated.followRedir == nil || !*isolated.followRedir {
		t.Fatalf("isolated client defaults ua=%q follow=%v", isolated.userAgent, isolated.followRedir)
	}
}

func TestRequestOptionContentTypeAndCharset(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("Content-Type")))
	}))
	defer srv.Close()

	got := Post(srv.URL, WithCharset("GBK"), WithContentType("text/custom")).BodyString("hello").Execute().Body()
	if got != "text/custom" {
		t.Fatalf("Content-Type = %q, want text/custom", got)
	}

	got = Post(srv.URL, WithCharset("GBK")).BodyJSON(`{"ok":true}`).Execute().Body()
	if got != "application/json;charset=GBK" {
		t.Fatalf("JSON Content-Type = %q", got)
	}
}

func TestResponseReadOptions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("abcdef"))
	}))
	defer srv.Close()

	limited := Get(srv.URL, WithMaxResponseBytes(3)).Execute()
	if got := limited.Bytes(); len(got) != 0 || limited.Err() == nil {
		t.Fatalf("limited Bytes() = %q err=%v, want max bytes error", string(got), limited.Err())
	}

	readAllCalled := false
	resp := Get(srv.URL, WithResponseReadAllFunc(func(r io.Reader) ([]byte, error) {
		readAllCalled = true
		return []byte("provided"), nil
	})).Execute()
	if got := resp.Body(); got != "provided" || !readAllCalled || resp.Err() != nil {
		t.Fatalf("custom readAll body=%q called=%v err=%v", got, readAllCalled, resp.Err())
	}
}

func TestResponseReadLimitFollowsGlobalConfigSnapshot(t *testing.T) {
	previous := SnapshotGlobalConfig()
	defer ConfigureGlobalConfig(previous)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("abcdef"))
	}))
	defer srv.Close()

	SetGlobalMaxResponseBytes(3)
	resp := Get(srv.URL).Execute()
	SetGlobalMaxResponseBytes(0)
	if got := resp.Bytes(); len(got) != 0 || resp.Err() == nil {
		t.Fatalf("global limited Bytes() = %q err=%v, want max bytes error", string(got), resp.Err())
	}

	unlimited := Get(srv.URL, WithMaxResponseBytes(0)).Execute()
	if got := unlimited.Body(); got != "abcdef" || unlimited.Err() != nil {
		t.Fatalf("unlimited override Body() = %q err=%v", got, unlimited.Err())
	}
}

func TestResponseIgnoreEOFFollowsRequestSnapshot(t *testing.T) {
	old := IsIgnoreEOFError()
	defer SetIgnoreEOFError(old)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("abcdef"))
	}))
	defer srv.Close()

	readUnexpectedEOF := func(io.Reader) ([]byte, error) {
		return []byte("partial"), io.ErrUnexpectedEOF
	}

	ignoreResp := NewRequestWithConfig(MethodGet, srv.URL, GlobalConfig{FollowRedirects: true, MaxRedirects: 10, MaxResponseBytes: defaultGlobalMaxResponseBytes, IgnoreEOFError: true}, WithResponseReadAllFunc(readUnexpectedEOF)).Execute()
	SetIgnoreEOFError(false)
	if got := ignoreResp.Body(); got != "partial" || ignoreResp.Err() != nil {
		t.Fatalf("ignore snapshot body=%q err=%v, want partial without error", got, ignoreResp.Err())
	}

	strictResp := NewRequestWithConfig(MethodGet, srv.URL, GlobalConfig{FollowRedirects: true, MaxRedirects: 10, MaxResponseBytes: defaultGlobalMaxResponseBytes, IgnoreEOFError: false}, WithResponseReadAllFunc(readUnexpectedEOF)).Execute()
	SetIgnoreEOFError(true)
	if got := strictResp.Body(); got != "" || strictResp.Err() == nil {
		t.Fatalf("strict snapshot body=%q err=%v, want read error", got, strictResp.Err())
	}
}

func TestRequestOptionTLSConfig(t *testing.T) {
	client := Get("https://example.com", WithTLSConfig(&tls.Config{ServerName: "example.com"})).buildClient()
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("transport type = %T", client.Transport)
	}
	if transport.TLSClientConfig == nil || transport.TLSClientConfig.ServerName != "example.com" {
		t.Fatalf("TLS config = %#v", transport.TLSClientConfig)
	}
}

func TestRequestOptionCookieJar(t *testing.T) {
	oldJar := GetCookieJar()
	CloseCookie()
	defer SetCookieJar(oldJar)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/set" {
			http.SetCookie(w, &http.Cookie{Name: "sid", Value: "abc", Path: "/"})
			_, _ = w.Write([]byte("set"))
			return
		}
		c, err := r.Cookie("sid")
		if err != nil {
			_, _ = w.Write([]byte("missing"))
			return
		}
		_, _ = w.Write([]byte(c.Value))
	}))
	defer srv.Close()

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookiejar.New() error = %v", err)
	}
	if resp := Get(srv.URL+"/set", WithCookieJar(jar)).Execute(); resp.Err() != nil {
		t.Fatalf("set cookie request error = %v", resp.Err())
	}
	if got := Get(srv.URL+"/get", WithCookieJar(jar)).Execute().Body(); got != "abc" {
		t.Fatalf("cookie jar body = %q, want abc", got)
	}
}

func TestDefaultTransportIsReused(t *testing.T) {
	clientA := Get("https://example.com").buildClient()
	clientB := Post("https://example.com").Timeout(time.Second).buildClient()
	shared := getDefaultTransport()

	if clientA.Transport != shared {
		t.Fatalf("default request transport = %p, want shared default transport %p", clientA.Transport, shared)
	}
	if clientB.Transport != shared {
		t.Fatalf("request with timeout transport = %p, want shared default transport %p", clientB.Transport, shared)
	}
}

func TestTransportProviderEvaluatedWhenBuildingClient(t *testing.T) {
	calls := 0
	custom := &http.Transport{}
	req := Get("https://example.com", WithTransportProvider(func() http.RoundTripper {
		calls++
		return custom
	}))
	if calls != 0 {
		t.Fatalf("transport provider called during construction: %d", calls)
	}
	client := req.buildClient()
	if calls != 1 || client.Transport != custom {
		t.Fatalf("transport provider calls=%d transport=%#v, want custom", calls, client.Transport)
	}
}

func TestDefaultTransportProviderCanBeConfiguredAndReset(t *testing.T) {
	custom := &http.Transport{MaxIdleConnsPerHost: 7}
	ConfigureDefaultTransportProvider(func() *http.Transport { return custom })
	t.Cleanup(ResetDefaultTransport)

	client := Get("https://example.com").buildClient()
	if client.Transport != custom {
		t.Fatalf("configured default transport = %p, want %p", client.Transport, custom)
	}

	ResetDefaultTransport()
	client = Get("https://example.com").buildClient()
	if client.Transport == custom {
		t.Fatal("ResetDefaultTransport should clear configured transport")
	}
	if _, ok := client.Transport.(*http.Transport); !ok {
		t.Fatalf("reset default transport type = %T, want *http.Transport", client.Transport)
	}
}

func TestSafeRequestRejectsPrivateAndUnsafeRedirects(t *testing.T) {
	if err := GetSafe("file:///tmp/secret.txt").Execute().Err(); err == nil {
		t.Fatal("GetSafe should reject non-HTTP schemes")
	}
	if err := GetSafe("http://127.0.0.1/config.yaml").Execute().Err(); err == nil {
		t.Fatal("GetSafe should reject loopback hosts by default")
	}
	if err := GetSafe("http://224.0.0.1/config.yaml").Execute().Err(); err == nil {
		t.Fatal("GetSafe should reject multicast hosts by default")
	}
	if err := GetSafe("http://0.0.0.0/config.yaml").Execute().Err(); err == nil {
		t.Fatal("GetSafe should reject unspecified hosts by default")
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/redirect":
			http.Redirect(w, r, "http://127.0.0.1/private", http.StatusFound)
		default:
			_, _ = w.Write([]byte("ok"))
		}
	}))
	defer srv.Close()

	serverURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parse server url: %v", err)
	}
	resp := GetSafe(srv.URL,
		WithURLPolicy(URLPolicy{AllowedSchemes: []string{"http", "https"}, AllowedHosts: []string{serverURL.Hostname()}}),
	).Execute()
	if body := resp.Body(); body != "ok" || resp.Err() != nil {
		t.Fatalf("GetSafe allowed public policy host body=%q err=%v", body, resp.Err())
	}
	if err := GetSafe(srv.URL+"/redirect",
		WithURLPolicy(URLPolicy{AllowedSchemes: []string{"http", "https"}, AllowedHosts: []string{serverURL.Hostname()}}),
	).Execute().Err(); err == nil {
		t.Fatal("GetSafe should reject unsafe redirect targets")
	}
}

func TestSafeRequestAllowedHostsDoesNotBypassPrivateRejection(t *testing.T) {
	if err := GetSafe("http://127.0.0.1/config.yaml", WithAllowedHosts("127.0.0.1")).Execute().Err(); err == nil {
		t.Fatal("GetSafe should reject allowlisted private hosts when RejectPrivate is enabled")
	}

	req := GetSafe("http://example.com/config.yaml",
		WithAllowedHosts("example.com"),
		WithLookupIP(func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		}),
		WithTransport(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("ok")), Header: make(http.Header)}, nil
		})),
	)
	if err := req.Execute().Err(); err == nil {
		t.Fatal("GetSafe should reject allowlisted hosts that resolve private during RoundTrip")
	}
}

func TestSafeRequestRevalidatesHostAtRoundTrip(t *testing.T) {
	req := GetSafe("http://example.com/config.yaml",
		WithLookupIP(func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		}),
		WithTransport(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("ok")), Header: make(http.Header)}, nil
		})),
	)

	if err := req.Execute().Err(); err == nil {
		t.Fatal("GetSafe should reject a host that resolves private during RoundTrip")
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
