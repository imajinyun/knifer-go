package vhttp

import (
	"time"

	httpx "github.com/imajinyun/go-knifer/internal/httpx/http"
)

// Get creates a GET request.
func Get(rawURL string, opts ...RequestOption) *Request { return httpx.Get(rawURL, opts...) }

// Post creates a POST request.
func Post(rawURL string, opts ...RequestOption) *Request { return httpx.Post(rawURL, opts...) }

// Put creates a PUT request.
func Put(rawURL string, opts ...RequestOption) *Request { return httpx.Put(rawURL, opts...) }

// Delete creates a DELETE request.
func Delete(rawURL string, opts ...RequestOption) *Request { return httpx.Delete(rawURL, opts...) }

// Patch creates a PATCH request.
func Patch(rawURL string, opts ...RequestOption) *Request { return httpx.Patch(rawURL, opts...) }

// Head creates a HEAD request.
func Head(rawURL string, opts ...RequestOption) *Request { return httpx.Head(rawURL, opts...) }

// Options delegates to the internal httpx implementation.
func Options(rawURL string, opts ...RequestOption) *Request {
	return httpx.Options(rawURL, opts...)
}

// NewRequest creates a request by method.
func NewRequest(method Method, rawURL string, opts ...RequestOption) *Request {
	return httpx.NewRequest(method, rawURL, opts...)
}

// CreateRequest delegates to the internal httpx implementation.
func CreateRequest(method Method, rawURL string, opts ...RequestOption) *Request {
	return httpx.NewRequest(method, rawURL, opts...)
}

// CreateGet delegates to the internal httpx implementation.
func CreateGet(rawURL string, followRedirects bool) *Request {
	return httpx.CreateGet(rawURL, followRedirects)
}

// CreatePost delegates to the internal httpx implementation.
func CreatePost(rawURL string) *Request {
	return httpx.CreatePost(rawURL)
}

// GetString sends a GET request and returns response body as string.
func GetString(rawURL string) string { return httpx.GetString(rawURL) }

// GetWithTimeout delegates to the internal httpx implementation.
func GetWithTimeout(rawURL string, timeout time.Duration) string {
	return httpx.GetWithTimeout(rawURL, timeout)
}

// GetWithParams delegates to the internal httpx implementation.
func GetWithParams(rawURL string, params map[string]any) string {
	return httpx.GetWithParams(rawURL, params)
}

// PostForm posts form parameters and returns response body as string.
func PostForm(rawURL string, params map[string]any) string { return httpx.PostForm(rawURL, params) }

// PostJSON posts JSON body and returns response body as string.
func PostJSON(rawURL, jsonStr string) string { return httpx.PostJSON(rawURL, jsonStr) }

// PostString delegates to the internal httpx implementation.
func PostString(rawURL, body string) string {
	return httpx.PostString(rawURL, body)
}
