package http

import (
	"crypto/tls"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
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

	resp := PostJSON(srv.URL, `{"a":1}`)
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

func TestRequestOptionTLSConfig(t *testing.T) {
	client := Get("https://example.com", WithTLSConfig(&tls.Config{ServerName: "example.com"})).buildClient()
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("transport type = %T", client.Transport)
	}
	if transport.TLSClientConfig == nil || transport.TLSClientConfig.ServerName != "example.com" {
		t.Fatalf("TLS config = %#v", transport.TLSClientConfig)
	}
	client = Get("https://example.com", WithTLSConfig(&tls.Config{ServerName: "example.com"}), WithSkipTLSVerify(true)).buildClient()
	transport = client.Transport.(*http.Transport)
	if !transport.TLSClientConfig.InsecureSkipVerify || transport.TLSClientConfig.ServerName != "example.com" {
		t.Fatalf("TLS config with skip verify = %#v", transport.TLSClientConfig)
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

func TestSkipTLSVerifyUsesClonedTransport(t *testing.T) {
	client := Get("https://example.com").SkipTLSVerify(true).buildClient()
	shared := getDefaultTransport()
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("transport type = %T, want *http.Transport", client.Transport)
	}
	if transport == shared {
		t.Fatal("SkipTLSVerify should clone the default transport instead of mutating it")
	}
	if transport.TLSClientConfig == nil || !transport.TLSClientConfig.InsecureSkipVerify {
		t.Fatal("SkipTLSVerify should enable InsecureSkipVerify on the cloned transport")
	}
	if shared.TLSClientConfig != nil && shared.TLSClientConfig.InsecureSkipVerify {
		t.Fatal("default transport must not be mutated by SkipTLSVerify")
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

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
