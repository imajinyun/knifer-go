package vhttp_test

import (
	"net/http"
	"testing"

	"github.com/imajinyun/go-knifer/vhttp"
)

func TestFacadeSharedConstants(t *testing.T) {
	_ = []vhttp.Method{vhttp.MethodTrace, vhttp.MethodConnect}
	_ = []vhttp.Header{vhttp.HeaderContentType, vhttp.HeaderUserAgent, vhttp.HeaderLocation}
	_ = []vhttp.ContentType{vhttp.ContentTypeJSON, vhttp.ContentTypeEventStream}

	if vhttp.MethodTrace.String() != http.MethodTrace {
		t.Fatalf("MethodTrace = %q", vhttp.MethodTrace.String())
	}
	if got := vhttp.ContentTypeJSON.WithCharset("UTF-8"); got != "application/json;charset=UTF-8" {
		t.Fatalf("ContentTypeJSON.WithCharset = %q", got)
	}
}
