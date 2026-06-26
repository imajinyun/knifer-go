package vhttp_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/imajinyun/knifer-go/vhttp"
)

const facadeDownloadText = "download-text"

func newFacadeDownloadServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Method", r.Method)
		if r.Method != http.MethodHead {
			_, _ = w.Write([]byte(facadeDownloadText))
		}
	}))
}

func allowLocalURLPolicy() vhttp.RequestOption {
	return vhttp.WithURLPolicy(vhttp.URLPolicy{AllowedSchemes: []string{"http", "https"}, RejectPrivate: false})
}
