package vresty_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/imajinyun/go-knifer/vresty"
	grestry "resty.dev/v3"
)

func TestFacadeRestyClientFactoryProvider(t *testing.T) {
	vresty.ResetDefaultRestyClientProvider()
	t.Cleanup(vresty.ResetDefaultRestyClientProvider)

	called := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("X-Factory")))
	}))
	defer server.Close()

	resp := vresty.Get(server.URL,
		vresty.WithRestyClientFactory(func() *grestry.Client {
			called++
			return grestry.New().SetHeader("X-Factory", "per-call")
		}),
	).Execute()
	if resp.Err() != nil {
		t.Fatal(resp.Err())
	}
	if called != 1 || resp.Body() != "per-call" {
		t.Fatalf("factory called=%d body=%q", called, resp.Body())
	}

	vresty.ConfigureDefaultRestyClientProvider(func() *grestry.Client {
		called++
		return grestry.New().SetHeader("X-Factory", "default")
	})
	resp = vresty.Get(server.URL).Execute()
	if resp.Err() != nil {
		t.Fatal(resp.Err())
	}
	if called != 2 || resp.Body() != "default" {
		t.Fatalf("default provider called=%d body=%q", called, resp.Body())
	}
}

func TestFacadeRequestOptionWrappers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "/final", http.StatusFound)
			return
		}
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		_, _ = w.Write([]byte(r.Method + ":" + string(body) + ":" + r.Header.Get("X-A") + ":" + r.Header.Get("X-B")))
	}))
	defer server.Close()

	resp := vresty.Post(server.URL,
		vresty.WithTimeout(time.Second),
		vresty.WithHeaders(map[string]string{"X-A": "one", "X-B": "two"}),
		vresty.WithContentType(string(vresty.ContentTypeTextPlain)),
		vresty.WithCharset("utf-8"),
		vresty.WithCookieDisabled(true),
		vresty.WithMaxResponseBytes(1024),
		vresty.WithMaxDecodeBytes(1024),
		vresty.WithJSONMarshalFunc(func(v any) ([]byte, error) { return []byte(`"custom"`), nil }),
		vresty.WithJSONUnmarshalFunc(func([]byte, any) error { return nil }),
		vresty.WithJSONDecodeReadAllFunc(io.ReadAll),
	).Body([]byte("payload")).Execute()
	if resp.Err() != nil {
		t.Fatalf("Post Execute: %v", resp.Err())
	}
	if got := resp.Body(); got != "POST:payload:one:two" {
		t.Fatalf("option body = %q", got)
	}

	redirect := vresty.Get(server.URL+"/redirect", vresty.WithMaxRedirects(0), vresty.WithFollowRedirects(false)).Execute()
	if redirect.Err() != nil {
		t.Fatalf("redirect Execute: %v", redirect.Err())
	}
	if got := redirect.Status(); got != http.StatusFound {
		t.Fatalf("redirect status = %d, want 302", got)
	}

	restyClient := grestry.New().SetHeader("X-A", "resty")
	withClient := vresty.Get(server.URL, vresty.WithRestyClient(restyClient)).Execute()
	if withClient.Err() != nil {
		t.Fatalf("WithRestyClient Execute: %v", withClient.Err())
	}
	if !strings.Contains(withClient.Body(), ":resty:") {
		t.Fatalf("WithRestyClient body = %q", withClient.Body())
	}
}
