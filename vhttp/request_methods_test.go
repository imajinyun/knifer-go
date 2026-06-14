package vhttp_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/imajinyun/go-knifer/vhttp"
)

func TestFacadeAdditionalHTTPMethods(t *testing.T) {
	var lastMethod string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastMethod = r.Method
		w.Header().Set("X-Method", r.Method)
		if r.Method != http.MethodHead {
			_, _ = w.Write([]byte(r.Method))
		}
	}))
	defer server.Close()

	tests := []struct {
		name   string
		method string
		req    *vhttp.Request
	}{
		{name: "put", method: http.MethodPut, req: vhttp.Put(server.URL)},
		{name: "delete", method: http.MethodDelete, req: vhttp.Delete(server.URL)},
		{name: "patch", method: http.MethodPatch, req: vhttp.Patch(server.URL)},
		{name: "head", method: http.MethodHead, req: vhttp.Head(server.URL)},
		{name: "options", method: http.MethodOptions, req: vhttp.Options(server.URL)},
		{name: "new request", method: http.MethodTrace, req: vhttp.NewRequest(vhttp.MethodTrace, server.URL)},
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
