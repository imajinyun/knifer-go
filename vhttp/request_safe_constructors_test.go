package vhttp_test

import (
	"net/http"
	"testing"

	"github.com/imajinyun/knifer-go/vhttp"
)

func TestFacadeSafeRequestConstructors(t *testing.T) {
	server := newFacadeDownloadServer(t)
	defer server.Close()

	allowLocal := allowLocalURLPolicy()
	tests := []struct {
		name   string
		method string
		req    *vhttp.Request
	}{
		{name: "get safe", method: http.MethodGet, req: vhttp.GetSafe(server.URL, allowLocal)},
		{name: "post safe", method: http.MethodPost, req: vhttp.PostSafe(server.URL, allowLocal)},
		{name: "put safe", method: http.MethodPut, req: vhttp.PutSafe(server.URL, allowLocal)},
		{name: "delete safe", method: http.MethodDelete, req: vhttp.DeleteSafe(server.URL, allowLocal)},
		{name: "patch safe", method: http.MethodPatch, req: vhttp.PatchSafe(server.URL, allowLocal)},
		{name: "head safe", method: http.MethodHead, req: vhttp.HeadSafe(server.URL, allowLocal)},
		{name: "options safe", method: http.MethodOptions, req: vhttp.OptionsSafe(server.URL, allowLocal)},
		{name: "new safe", method: http.MethodTrace, req: vhttp.NewSafeRequest(vhttp.MethodTrace, server.URL, allowLocal)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := tt.req.Execute()
			if resp.Err() != nil {
				t.Fatalf("Execute: %v", resp.Err())
			}
			if got := resp.Header("X-Method"); got != tt.method {
				t.Fatalf("method header = %q, want %q", got, tt.method)
			}
		})
	}
}
