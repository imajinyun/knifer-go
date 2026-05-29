package vhttp

import (
	"io"
	"net/http"
	"time"

	httpx "github.com/imajinyun/go-knifer/internal/http"
)

// Request is a chainable HTTP request builder.
type Request = httpx.HTTPRequest

// Response wraps an HTTP response.
type Response = httpx.HTTPResponse

// Method represents an HTTP method.
type Method = httpx.Method

// Header represents an HTTP header name.
type Header = httpx.Header

// ContentType represents an HTTP content type.
type ContentType = httpx.ContentType

// Error is the HTTP module error type.
type Error = httpx.HTTPError

// SimpleServer is a small HTTP server helper.
type SimpleServer = httpx.SimpleServer

// UserAgent describes parsed User-Agent information.
type UserAgent = httpx.UserAgent

const (
	// MethodGet is GET.
	MethodGet Method = httpx.MethodGet
	// MethodPost is POST.
	MethodPost Method = httpx.MethodPost
	// MethodPut is PUT.
	MethodPut Method = httpx.MethodPut
	// MethodDelete is DELETE.
	MethodDelete Method = httpx.MethodDelete
	// MethodPatch is PATCH.
	MethodPatch Method = httpx.MethodPatch
	// MethodHead is HEAD.
	MethodHead Method = httpx.MethodHead
	// MethodOptions is OPTIONS.
	MethodOptions Method = httpx.MethodOptions
)

// Get creates a GET request.
func Get(rawURL string) *Request { return httpx.Get(rawURL) }

// Post creates a POST request.
func Post(rawURL string) *Request { return httpx.Post(rawURL) }

// Put creates a PUT request.
func Put(rawURL string) *Request { return httpx.Put(rawURL) }

// Delete creates a DELETE request.
func Delete(rawURL string) *Request { return httpx.Delete(rawURL) }

// Patch creates a PATCH request.
func Patch(rawURL string) *Request { return httpx.Patch(rawURL) }

// Head creates a HEAD request.
func Head(rawURL string) *Request { return httpx.Head(rawURL) }

// NewRequest creates a request by method.
func NewRequest(method Method, rawURL string) *Request {
	return httpx.NewRequest(method, rawURL)
}

// GetString sends a GET request and returns response body as string.
func GetString(rawURL string) string { return httpx.GetString(rawURL) }

// PostForm posts form parameters and returns response body as string.
func PostForm(rawURL string, params map[string]any) string { return httpx.PostForm(rawURL, params) }

// PostJSON posts JSON body and returns response body as string.
func PostJSON(rawURL, jsonStr string) string { return httpx.PostJSON(rawURL, jsonStr) }

// Download downloads rawURL into w.
func Download(rawURL string, w io.Writer) (int64, error) { return httpx.Download(rawURL, w) }

// DownloadFile downloads rawURL to dest.
func DownloadFile(rawURL, dest string) (int64, error) { return httpx.DownloadFile(rawURL, dest) }

// SetGlobalTimeout sets the global HTTP timeout.
func SetGlobalTimeout(d time.Duration) { httpx.SetGlobalTimeout(d) }

// GetGlobalTimeout returns the global HTTP timeout.
func GetGlobalTimeout() time.Duration { return httpx.GetGlobalTimeout() }

// SetGlobalHeader sets a global HTTP header.
func SetGlobalHeader(name, value string) { httpx.SetGlobalHeader(name, value) }

// AddGlobalHeader adds a global HTTP header value.
func AddGlobalHeader(name, value string) { httpx.AddGlobalHeader(name, value) }

// RemoveGlobalHeader removes a global HTTP header.
func RemoveGlobalHeader(name string) { httpx.RemoveGlobalHeader(name) }

// CloneGlobalHeaders returns cloned global headers.
func CloneGlobalHeaders() http.Header { return httpx.CloneGlobalHeaders() }

// BuildBasicAuth builds a Basic authorization value.
func BuildBasicAuth(user, pass string) string { return httpx.BuildBasicAuth(user, pass) }

// ToParams converts a map to query parameters.
func ToParams(m map[string]any) string { return httpx.ToParams(m) }

// ParseUserAgent parses a User-Agent string.
func ParseUserAgent(ua string) *UserAgent { return httpx.ParseUserAgent(ua) }

// NewSimpleServer creates a simple HTTP server on port.
func NewSimpleServer(port int) *SimpleServer { return httpx.NewSimpleServer(port) }
