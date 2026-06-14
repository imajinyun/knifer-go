package http

import "testing"

func TestNewRequest(t *testing.T) {
	req := NewRequest(MethodPut, "http://example.com")
	if req.method != MethodPut {
		t.Fatalf("method: %v", req.method)
	}
}

func TestGetWithFollowRedirectsOption(t *testing.T) {
	req := Get("http://example.com", WithFollowRedirects(false))
	if req.followRedir == nil || *req.followRedir != false {
		t.Fatalf("followRedir: %v", req.followRedir)
	}
}

func TestNewRequestWithOptionsAppliesRequestOptions(t *testing.T) {
	getReq := Get("http://example.com", WithFollowRedirects(false), WithHeader("X-Create", "get"), WithUserAgent("create-get-agent"))
	if getReq.followRedir == nil || *getReq.followRedir {
		t.Fatalf("followRedir: %v", getReq.followRedir)
	}
	if got := getReq.headers.Get("X-Create"); got != "get" {
		t.Fatalf("Get header = %q, want get", got)
	}
	if got := getReq.userAgent; got != "create-get-agent" {
		t.Fatalf("Get userAgent = %q", got)
	}

	postReq := Post("http://example.com", WithHeader("X-Create", "post"))
	if postReq.method != MethodPost {
		t.Fatalf("Post method = %v, want POST", postReq.method)
	}
	if got := postReq.headers.Get("X-Create"); got != "post" {
		t.Fatalf("Post header = %q, want post", got)
	}
}
