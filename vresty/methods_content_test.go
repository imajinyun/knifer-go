package vresty_test

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/imajinyun/knifer-go/vresty"
)

func TestFacadeAdditionalMethods(t *testing.T) {
	var lastMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastMethod = r.Method
		w.Header().Set("X-Method", r.Method)
		if r.Method != http.MethodHead {
			_, _ = w.Write([]byte(r.Method))
		}
	}))
	defer srv.Close()

	tests := []struct {
		name   string
		method string
		req    *vresty.Request
	}{
		{name: "put", method: http.MethodPut, req: vresty.Put(srv.URL)},
		{name: "delete", method: http.MethodDelete, req: vresty.Delete(srv.URL)},
		{name: "patch", method: http.MethodPatch, req: vresty.Patch(srv.URL)},
		{name: "head", method: http.MethodHead, req: vresty.Head(srv.URL)},
		{name: "options", method: http.MethodOptions, req: vresty.Options(srv.URL)},
		{name: "new request", method: http.MethodTrace, req: vresty.NewRequest(vresty.MethodTrace, srv.URL)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := tt.req.Execute()
			if resp.Err() != nil {
				t.Fatalf("Execute: %v", resp.Err())
			}
			if lastMethod != tt.method {
				t.Fatalf("server method = %q, want %q", lastMethod, tt.method)
			}
			if got := resp.Header("X-Method"); got != tt.method {
				t.Fatalf("response method header = %q, want %q", got, tt.method)
			}
		})
	}
}

func TestFacadeContentHelpers(t *testing.T) {
	if got := vresty.BuildContentType("application/json", "utf-8"); got != "application/json;charset=utf-8" {
		t.Fatalf("BuildContentType = %q", got)
	}
	if got := vresty.GuessContentType("<root/>"); got != vresty.ContentTypeXML {
		t.Fatalf("GuessContentType = %q", got)
	}
	if !vresty.IsDefaultContentType("") || !vresty.IsFormURLEncoded("application/x-www-form-urlencoded; charset=utf-8") {
		t.Fatal("content type predicates returned unexpected result")
	}
	if got := vresty.URLWithForm("https://example.com/path?x=1", map[string]any{"q": "go"}); got != "https://example.com/path?x=1&q=go" {
		t.Fatalf("URLWithForm = %q", got)
	}
	if got := vresty.GetCharsetFromContentTypeWithOptions("text/plain; enc=gbk", vresty.WithCharsetRegexp(regexp.MustCompile(`enc=([a-z0-9-]+)`))); got != "gbk" {
		t.Fatalf("GetCharsetFromContentTypeWithOptions = %q", got)
	}
	if got := vresty.GetCharsetFromHTMLWithOptions(`<meta data-charset="big5">`, vresty.WithMetaCharsetRegexp(regexp.MustCompile(`data-charset="([^"]+)"`))); got != "big5" {
		t.Fatalf("GetCharsetFromHTMLWithOptions = %q", got)
	}
	if got := vresty.GetMimeType("payload.zip"); got != "application/zip" {
		t.Fatalf("GetMimeType = %q", got)
	}
}
