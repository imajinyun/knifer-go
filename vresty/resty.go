package vresty

import (
	"crypto/tls"
	"io"
	"io/fs"
	"time"

	restyimpl "github.com/imajinyun/go-knifer/internal/httpx/resty"
	grestry "resty.dev/v3"
)

// Request is a chainable HTTP request builder backed by resty.
type Request = restyimpl.HTTPRequest

// RequestOption customizes one HTTP request at construction time.
type RequestOption = restyimpl.RequestOption

// Response wraps an HTTP response.
type Response = restyimpl.HTTPResponse

// SaveOption customizes response file saving.
type SaveOption = restyimpl.SaveOption

// Method represents an HTTP method.
type Method = restyimpl.Method

// Header represents an HTTP header name.
type Header = restyimpl.Header

// ContentType represents an HTTP content type.
type ContentType = restyimpl.ContentType

// HeaderValues stores HTTP header values.
type HeaderValues = restyimpl.HeaderValues

// Cookie contains a response cookie name and value.
type Cookie = restyimpl.Cookie

// Error is the HTTP module error type.
type Error = restyimpl.HTTPError

const (
	// MethodGet is GET.
	MethodGet Method = restyimpl.MethodGet
	// MethodPost is POST.
	MethodPost Method = restyimpl.MethodPost
	// MethodPut is PUT.
	MethodPut Method = restyimpl.MethodPut
	// MethodDelete is DELETE.
	MethodDelete Method = restyimpl.MethodDelete
	// MethodPatch is PATCH.
	MethodPatch Method = restyimpl.MethodPatch
	// MethodHead is HEAD.
	MethodHead Method = restyimpl.MethodHead
	// MethodOptions is OPTIONS.
	MethodOptions Method = restyimpl.MethodOptions
	// MethodTrace is TRACE.
	MethodTrace Method = restyimpl.MethodTrace
	// MethodConnect is CONNECT.
	MethodConnect Method = restyimpl.MethodConnect
)

// Get creates a GET request.
func Get(rawURL string, opts ...RequestOption) *Request { return restyimpl.Get(rawURL, opts...) }

// Post creates a POST request.
func Post(rawURL string, opts ...RequestOption) *Request { return restyimpl.Post(rawURL, opts...) }

// Put creates a PUT request.
func Put(rawURL string, opts ...RequestOption) *Request { return restyimpl.Put(rawURL, opts...) }

// Delete creates a DELETE request.
func Delete(rawURL string, opts ...RequestOption) *Request { return restyimpl.Delete(rawURL, opts...) }

// Patch creates a PATCH request.
func Patch(rawURL string, opts ...RequestOption) *Request { return restyimpl.Patch(rawURL, opts...) }

// Head creates a HEAD request.
func Head(rawURL string, opts ...RequestOption) *Request { return restyimpl.Head(rawURL, opts...) }

// Options creates an OPTIONS request.
func Options(rawURL string, opts ...RequestOption) *Request {
	return restyimpl.Options(rawURL, opts...)
}

// NewRequest creates a request by method.
func NewRequest(method Method, rawURL string, opts ...RequestOption) *Request {
	return restyimpl.NewRequest(method, rawURL, opts...)
}

// WithTimeout sets a per-request timeout.
func WithTimeout(d time.Duration) RequestOption { return restyimpl.WithTimeout(d) }

// WithHeader sets one per-request header.
func WithHeader(name, value string) RequestOption { return restyimpl.WithHeader(name, value) }

// WithHeaders sets per-request headers in batch.
func WithHeaders(headers map[string]string) RequestOption { return restyimpl.WithHeaders(headers) }

// WithFollowRedirects sets per-request redirect behavior.
func WithFollowRedirects(b bool) RequestOption { return restyimpl.WithFollowRedirects(b) }

// WithMaxRedirects sets the per-request redirect limit.
func WithMaxRedirects(n int) RequestOption { return restyimpl.WithMaxRedirects(n) }

// WithSkipTLSVerify sets per-request TLS verification behavior.
func WithSkipTLSVerify(b bool) RequestOption { return restyimpl.WithSkipTLSVerify(b) }

// WithTLSConfig sets a per-request TLS config. It is ignored when WithRestyClient is set.
func WithTLSConfig(cfg *tls.Config) RequestOption { return restyimpl.WithTLSConfig(cfg) }

// WithRestyClient sets a per-request resty client and takes precedence over WithTLSConfig.
func WithRestyClient(c *grestry.Client) RequestOption { return restyimpl.WithRestyClient(c) }

// WithUserAgent sets a per-request User-Agent.
func WithUserAgent(ua string) RequestOption { return restyimpl.WithUserAgent(ua) }

// WithCookieDisabled sets per-request cookie management behavior.
func WithCookieDisabled(disabled bool) RequestOption { return restyimpl.WithCookieDisabled(disabled) }

// WithContentType sets a per-request Content-Type at construction time.
func WithContentType(ct string) RequestOption { return restyimpl.WithContentType(ct) }

// WithCharset sets a per-request charset at construction time.
func WithCharset(charset string) RequestOption { return restyimpl.WithCharset(charset) }

// WithSaveFilePerm sets the file permission used when creating the destination file.
func WithSaveFilePerm(perm fs.FileMode) SaveOption { return restyimpl.WithSaveFilePerm(perm) }

// WithSaveDirPerm sets the directory permission used when creating parent directories.
func WithSaveDirPerm(perm fs.FileMode) SaveOption { return restyimpl.WithSaveDirPerm(perm) }

// WithSaveOverwrite controls whether an existing destination file may be replaced.
func WithSaveOverwrite(overwrite bool) SaveOption { return restyimpl.WithSaveOverwrite(overwrite) }

// WithSaveCreateParents controls whether parent directories are created automatically.
func WithSaveCreateParents(create bool) SaveOption { return restyimpl.WithSaveCreateParents(create) }

// WithSaveDefaultFilename sets the fallback file name used when dest is a directory.
func WithSaveDefaultFilename(name string) SaveOption { return restyimpl.WithSaveDefaultFilename(name) }

// GetString sends a GET request and returns response body as string.
func GetString(rawURL string) string { return restyimpl.GetString(rawURL) }

// PostForm posts form parameters and returns response body as string.
func PostForm(rawURL string, params map[string]any) string { return restyimpl.PostForm(rawURL, params) }

// PostJSON posts JSON body and returns response body as string.
func PostJSON(rawURL, jsonStr string) string { return restyimpl.PostJSON(rawURL, jsonStr) }

// Download downloads rawURL into w.
func Download(rawURL string, w io.Writer) (int64, error) { return restyimpl.Download(rawURL, w) }

// DownloadWithOptions downloads rawURL into w with per-request options.
func DownloadWithOptions(rawURL string, w io.Writer, opts ...RequestOption) (int64, error) {
	return restyimpl.DownloadWithOptions(rawURL, w, opts...)
}

// DownloadFile downloads rawURL to dest.
func DownloadFile(rawURL, dest string, opts ...SaveOption) (int64, error) {
	return restyimpl.DownloadFile(rawURL, dest, opts...)
}

// DownloadFileWithOptions downloads rawURL to dest with per-request and per-save options.
func DownloadFileWithOptions(rawURL, dest string, requestOpts []RequestOption, saveOpts ...SaveOption) (int64, error) {
	return restyimpl.DownloadFileWithOptions(rawURL, dest, requestOpts, saveOpts...)
}

// DownloadBytes downloads and returns bytes.
func DownloadBytes(rawURL string) []byte { return restyimpl.DownloadBytes(rawURL) }

// DownloadBytesWithOptions downloads and returns bytes with per-request options.
func DownloadBytesWithOptions(rawURL string, opts ...RequestOption) []byte {
	return restyimpl.DownloadBytesWithOptions(rawURL, opts...)
}

// DownloadString downloads remote text.
func DownloadString(rawURL, customCharset string) string {
	return restyimpl.DownloadString(rawURL, customCharset)
}

// DownloadStringWithOptions downloads remote text with per-request options.
func DownloadStringWithOptions(rawURL, customCharset string, opts ...RequestOption) string {
	return restyimpl.DownloadStringWithOptions(rawURL, customCharset, opts...)
}

// SetGlobalTimeout sets the global HTTP timeout.
func SetGlobalTimeout(d time.Duration) { restyimpl.SetGlobalTimeout(d) }

// GetGlobalTimeout returns the global HTTP timeout.
func GetGlobalTimeout() time.Duration { return restyimpl.GetGlobalTimeout() }

// SetGlobalHeader sets a global HTTP header.
func SetGlobalHeader(name, value string) { restyimpl.SetGlobalHeader(name, value) }

// AddGlobalHeader adds a global HTTP header value.
func AddGlobalHeader(name, value string) { restyimpl.AddGlobalHeader(name, value) }

// RemoveGlobalHeader removes a global HTTP header.
func RemoveGlobalHeader(name string) { restyimpl.RemoveGlobalHeader(name) }

// CloneGlobalHeaders returns cloned global headers.
func CloneGlobalHeaders() HeaderValues { return restyimpl.CloneGlobalHeaders() }

// CloseCookie disables global cookie management.
func CloseCookie() { restyimpl.CloseCookie() }

// BuildBasicAuth builds a Basic authorization value.
func BuildBasicAuth(user, pass string) string { return restyimpl.BuildBasicAuth(user, pass) }

// BuildContentType builds a Content-Type string with charset.
func BuildContentType(contentType, charset string) string {
	return restyimpl.BuildContentType(contentType, charset)
}

// GuessContentType guesses Content-Type from the body.
func GuessContentType(body string) ContentType { return restyimpl.GuessContentType(body) }

// GetCharsetFromContentType extracts charset from Content-Type.
func GetCharsetFromContentType(ct string) string { return restyimpl.GetCharsetFromContentType(ct) }

// GetCharsetFromHTML extracts charset from HTML meta tags.
func GetCharsetFromHTML(html string) string { return restyimpl.GetCharsetFromHTML(html) }

// GetMimeType returns the MIME type by file extension.
func GetMimeType(filename string) string { return restyimpl.GetMimeType(filename) }
