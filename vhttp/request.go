package vhttp

import (
	"time"

	httpx "github.com/imajinyun/knifer-go/internal/httpx/http"
)

// Get creates a GET request.
//
// Security: Get is for trusted URLs. Use GetSafe when the URL is untrusted.
func Get(rawURL string, opts ...RequestOption) *Request { return httpx.Get(rawURL, opts...) }

// NewClient creates a request factory using the current global configuration snapshot.
func NewClient(opts ...ClientOption) *Client { return httpx.NewClient(opts...) }

// NewIsolatedClient creates a request factory without reading package-level global defaults.
func NewIsolatedClient(opts ...ClientOption) *Client { return httpx.NewIsolatedClient(opts...) }

// NewClientWithConfig creates a request factory from an explicit configuration snapshot.
func NewClientWithConfig(cfg GlobalConfig, opts ...RequestOption) *Client {
	return httpx.NewClientWithConfig(cfg, opts...)
}

// WithClientGlobalConfig sets the configuration snapshot used by a Client.
func WithClientGlobalConfig(cfg GlobalConfig) ClientOption { return httpx.WithClientGlobalConfig(cfg) }

// WithClientRequestOptions sets request options applied to every request created by a Client.
func WithClientRequestOptions(opts ...RequestOption) ClientOption {
	return httpx.WithClientRequestOptions(opts...)
}

// GetSafe creates a GET request with SSRF-oriented safety checks enabled.
func GetSafe(rawURL string, opts ...RequestOption) *Request { return httpx.GetSafe(rawURL, opts...) }

// Post creates a POST request.
//
// Security: Post is for trusted URLs. Use PostSafe when the URL is untrusted.
func Post(rawURL string, opts ...RequestOption) *Request { return httpx.Post(rawURL, opts...) }

// PostSafe creates a POST request with SSRF-oriented safety checks enabled.
func PostSafe(rawURL string, opts ...RequestOption) *Request { return httpx.PostSafe(rawURL, opts...) }

// Put creates a PUT request.
//
// Security: Put is for trusted URLs. Use PutSafe when the URL is untrusted.
func Put(rawURL string, opts ...RequestOption) *Request { return httpx.Put(rawURL, opts...) }

// PutSafe creates a PUT request with SSRF-oriented safety checks enabled.
func PutSafe(rawURL string, opts ...RequestOption) *Request { return httpx.PutSafe(rawURL, opts...) }

// Delete creates a DELETE request.
//
// Security: Delete is for trusted URLs. Use DeleteSafe when the URL is untrusted.
func Delete(rawURL string, opts ...RequestOption) *Request { return httpx.Delete(rawURL, opts...) }

// DeleteSafe creates a DELETE request with SSRF-oriented safety checks enabled.
func DeleteSafe(rawURL string, opts ...RequestOption) *Request {
	return httpx.DeleteSafe(rawURL, opts...)
}

// Patch creates a PATCH request.
//
// Security: Patch is for trusted URLs. Use PatchSafe when the URL is untrusted.
func Patch(rawURL string, opts ...RequestOption) *Request { return httpx.Patch(rawURL, opts...) }

// PatchSafe creates a PATCH request with SSRF-oriented safety checks enabled.
func PatchSafe(rawURL string, opts ...RequestOption) *Request {
	return httpx.PatchSafe(rawURL, opts...)
}

// Head creates a HEAD request.
//
// Security: Head is for trusted URLs. Use HeadSafe when the URL is untrusted.
func Head(rawURL string, opts ...RequestOption) *Request { return httpx.Head(rawURL, opts...) }

// HeadSafe creates a HEAD request with SSRF-oriented safety checks enabled.
func HeadSafe(rawURL string, opts ...RequestOption) *Request { return httpx.HeadSafe(rawURL, opts...) }

// Options creates an OPTIONS request.
//
// Security: Options is for trusted URLs. Use OptionsSafe when the URL is untrusted.
func Options(rawURL string, opts ...RequestOption) *Request {
	return httpx.Options(rawURL, opts...)
}

// OptionsSafe creates an OPTIONS request with SSRF-oriented safety checks enabled.
func OptionsSafe(rawURL string, opts ...RequestOption) *Request {
	return httpx.OptionsSafe(rawURL, opts...)
}

// NewRequest creates a request by method.
//
// Security: NewRequest is for trusted URLs. Use NewSafeRequest when the URL is untrusted.
func NewRequest(method Method, rawURL string, opts ...RequestOption) *Request {
	return httpx.NewRequest(method, rawURL, opts...)
}

// NewSafeRequest creates a request with SSRF-oriented safety checks enabled.
func NewSafeRequest(method Method, rawURL string, opts ...RequestOption) *Request {
	return httpx.NewSafeRequest(method, rawURL, opts...)
}

// NewIsolatedRequest creates a request without reading package-level global defaults.
func NewIsolatedRequest(method Method, rawURL string, opts ...RequestOption) *Request {
	return httpx.NewIsolatedRequest(method, rawURL, opts...)
}

// NewRequestWithConfig creates a request from an explicit global configuration snapshot.
//
// Security: NewRequestWithConfig is for trusted URLs. Use NewSafeRequest when the URL is untrusted.
func NewRequestWithConfig(method Method, rawURL string, cfg GlobalConfig, opts ...RequestOption) *Request {
	return httpx.NewRequestWithConfig(method, rawURL, cfg, opts...)
}

// GetStringE sends a GET request and returns response body as string or an error.
func GetStringE(rawURL string) (string, error) { return GetStringEWithOptions(rawURL) }

// GetStringEWithOptions sends a GET request with options and returns response body as string or an error.
func GetStringEWithOptions(rawURL string, opts ...RequestOption) (string, error) {
	return httpx.GetStringEWithOptions(rawURL, opts...)
}

// GetStringSafeE sends a safe GET request and returns response body as string or an error.
func GetStringSafeE(rawURL string, opts ...RequestOption) (string, error) {
	return httpx.GetStringSafeE(rawURL, opts...)
}

// GetWithTimeoutE sends a GET request with a timeout and returns response body or an error.
func GetWithTimeoutE(rawURL string, timeout time.Duration) (string, error) {
	return GetWithTimeoutEWithOptions(rawURL, timeout)
}

// GetWithTimeoutEWithOptions sends a GET request with a timeout and custom options, returning body or error.
func GetWithTimeoutEWithOptions(rawURL string, timeout time.Duration, opts ...RequestOption) (string, error) {
	return httpx.GetWithTimeoutEWithOptions(rawURL, timeout, opts...)
}

// GetWithParamsE sends a GET request with form parameters and returns response body or an error.
func GetWithParamsE(rawURL string, params map[string]any) (string, error) {
	return GetWithParamsEWithOptions(rawURL, params)
}

// GetWithParamsEWithOptions sends a GET request with form parameters and custom options, returning body or error.
func GetWithParamsEWithOptions(rawURL string, params map[string]any, opts ...RequestOption) (string, error) {
	return httpx.GetWithParamsEWithOptions(rawURL, params, opts...)
}

// PostFormE posts form parameters and returns response body or an error.
func PostFormE(rawURL string, params map[string]any) (string, error) {
	return PostFormEWithOptions(rawURL, params)
}

// PostFormEWithOptions posts form parameters with options and returns response body or an error.
func PostFormEWithOptions(rawURL string, params map[string]any, opts ...RequestOption) (string, error) {
	return httpx.PostFormEWithOptions(rawURL, params, opts...)
}

// PostFormSafeE posts form parameters with SSRF-oriented safety checks enabled.
func PostFormSafeE(rawURL string, params map[string]any, opts ...RequestOption) (string, error) {
	return httpx.PostFormSafeE(rawURL, params, opts...)
}

// PostJSONE posts JSON body and returns response body or an error.
func PostJSONE(rawURL, jsonStr string) (string, error) { return PostJSONEWithOptions(rawURL, jsonStr) }

// PostJSONEWithOptions posts JSON body with options and returns response body or an error.
func PostJSONEWithOptions(rawURL, jsonStr string, opts ...RequestOption) (string, error) {
	return httpx.PostJSONEWithOptions(rawURL, jsonStr, opts...)
}

// PostJSONSafeE posts JSON body with SSRF-oriented safety checks enabled.
func PostJSONSafeE(rawURL, jsonStr string, opts ...RequestOption) (string, error) {
	return httpx.PostJSONSafeE(rawURL, jsonStr, opts...)
}

// PostStringE posts a string body and returns response body or an error.
func PostStringE(rawURL, body string) (string, error) { return PostStringEWithOptions(rawURL, body) }

// PostStringEWithOptions posts a string body with options and returns response body or an error.
func PostStringEWithOptions(rawURL, body string, opts ...RequestOption) (string, error) {
	return httpx.PostStringEWithOptions(rawURL, body, opts...)
}

// PostStringSafeE posts a string body with SSRF-oriented safety checks enabled.
func PostStringSafeE(rawURL, body string, opts ...RequestOption) (string, error) {
	return httpx.PostStringSafeE(rawURL, body, opts...)
}
