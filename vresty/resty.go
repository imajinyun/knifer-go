package vresty

import (
	"context"
	"crypto/tls"
	"io"
	"io/fs"
	"net"
	"os"
	"regexp"
	"time"

	restyimpl "github.com/imajinyun/knifer-go/internal/httpx/resty"
	grestry "resty.dev/v3"
)

// Request is a chainable HTTP request builder backed by resty.
type Request = restyimpl.HTTPRequest

// Client is an explicit resty request factory with a captured configuration snapshot.
type Client = restyimpl.Client

// RequestOption customizes one HTTP request at construction time.
type RequestOption = restyimpl.RequestOption

// ClientOption customizes a Client.
type ClientOption = restyimpl.ClientOption

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

// GlobalConfig captures resty package-level defaults for explicit request construction.
type GlobalConfig = restyimpl.GlobalConfig

// URLPolicy controls SSRF-oriented request validation for untrusted URLs.
type URLPolicy = restyimpl.URLPolicy

// CharsetOption customizes charset extraction helpers per call.
type CharsetOption = restyimpl.CharsetOption

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

const (
	HeaderAuthorization      Header = restyimpl.HeaderAuthorization
	HeaderProxyAuthorization Header = restyimpl.HeaderProxyAuthorization
	HeaderDate               Header = restyimpl.HeaderDate
	HeaderConnection         Header = restyimpl.HeaderConnection
	HeaderMimeVersion        Header = restyimpl.HeaderMimeVersion
	HeaderTrailer            Header = restyimpl.HeaderTrailer
	HeaderTransferEncoding   Header = restyimpl.HeaderTransferEncoding
	HeaderUpgrade            Header = restyimpl.HeaderUpgrade
	HeaderVia                Header = restyimpl.HeaderVia
	HeaderCacheControl       Header = restyimpl.HeaderCacheControl
	HeaderPragma             Header = restyimpl.HeaderPragma
	HeaderContentType        Header = restyimpl.HeaderContentType
	HeaderHost               Header = restyimpl.HeaderHost
	HeaderReferer            Header = restyimpl.HeaderReferer
	HeaderOrigin             Header = restyimpl.HeaderOrigin
	HeaderUserAgent          Header = restyimpl.HeaderUserAgent
	HeaderAccept             Header = restyimpl.HeaderAccept
	HeaderAcceptLanguage     Header = restyimpl.HeaderAcceptLanguage
	HeaderAcceptEncoding     Header = restyimpl.HeaderAcceptEncoding
	HeaderAcceptCharset      Header = restyimpl.HeaderAcceptCharset
	HeaderCookie             Header = restyimpl.HeaderCookie
	HeaderContentLength      Header = restyimpl.HeaderContentLength
	HeaderWWWAuthenticate    Header = restyimpl.HeaderWWWAuthenticate
	HeaderSetCookie          Header = restyimpl.HeaderSetCookie
	HeaderContentEncoding    Header = restyimpl.HeaderContentEncoding
	HeaderContentDisposition Header = restyimpl.HeaderContentDisposition
	HeaderETag               Header = restyimpl.HeaderETag
	HeaderLocation           Header = restyimpl.HeaderLocation
)

const (
	ContentTypeFormURLEncoded ContentType = restyimpl.ContentTypeFormURLEncoded
	ContentTypeMultipart      ContentType = restyimpl.ContentTypeMultipart
	ContentTypeJSON           ContentType = restyimpl.ContentTypeJSON
	ContentTypeXML            ContentType = restyimpl.ContentTypeXML
	ContentTypeTextPlain      ContentType = restyimpl.ContentTypeTextPlain
	ContentTypeTextXML        ContentType = restyimpl.ContentTypeTextXML
	ContentTypeTextHTML       ContentType = restyimpl.ContentTypeTextHTML
	ContentTypeOctetStream    ContentType = restyimpl.ContentTypeOctetStream
	ContentTypeEventStream    ContentType = restyimpl.ContentTypeEventStream
)

// Get creates a GET request.
//
// Security: Get is for trusted URLs. Use GetSafe for untrusted URLs.
func Get(rawURL string, opts ...RequestOption) *Request { return restyimpl.Get(rawURL, opts...) }

// NewClient creates a request factory using the current global configuration snapshot.
func NewClient(opts ...ClientOption) *Client { return restyimpl.NewClient(opts...) }

// NewIsolatedClient creates a request factory without reading package-level global defaults.
func NewIsolatedClient(opts ...ClientOption) *Client { return restyimpl.NewIsolatedClient(opts...) }

// NewClientWithConfig creates a request factory from an explicit configuration snapshot.
func NewClientWithConfig(cfg GlobalConfig, opts ...RequestOption) *Client {
	return restyimpl.NewClientWithConfig(cfg, opts...)
}

// WithClientGlobalConfig sets the configuration snapshot used by a Client.
func WithClientGlobalConfig(cfg GlobalConfig) ClientOption {
	return restyimpl.WithClientGlobalConfig(cfg)
}

// WithClientRequestOptions sets request options applied to every request created by a Client.
func WithClientRequestOptions(opts ...RequestOption) ClientOption {
	return restyimpl.WithClientRequestOptions(opts...)
}

// GetSafe creates a GET request with SSRF-oriented safety checks enabled.
func GetSafe(rawURL string, opts ...RequestOption) *Request {
	return restyimpl.GetSafe(rawURL, opts...)
}

// Post creates a POST request.
//
// Security: Post is for trusted URLs. Use PostSafe for untrusted URLs.
func Post(rawURL string, opts ...RequestOption) *Request { return restyimpl.Post(rawURL, opts...) }

// PostSafe creates a POST request with SSRF-oriented safety checks enabled.
func PostSafe(rawURL string, opts ...RequestOption) *Request {
	return restyimpl.PostSafe(rawURL, opts...)
}

// Put creates a PUT request.
//
// Security: Put is for trusted URLs. Use PutSafe for untrusted URLs.
func Put(rawURL string, opts ...RequestOption) *Request { return restyimpl.Put(rawURL, opts...) }

// PutSafe creates a PUT request with SSRF-oriented safety checks enabled.
func PutSafe(rawURL string, opts ...RequestOption) *Request {
	return restyimpl.PutSafe(rawURL, opts...)
}

// Delete creates a DELETE request.
//
// Security: Delete is for trusted URLs. Use DeleteSafe for untrusted URLs.
func Delete(rawURL string, opts ...RequestOption) *Request { return restyimpl.Delete(rawURL, opts...) }

// DeleteSafe creates a DELETE request with SSRF-oriented safety checks enabled.
func DeleteSafe(rawURL string, opts ...RequestOption) *Request {
	return restyimpl.DeleteSafe(rawURL, opts...)
}

// Patch creates a PATCH request.
//
// Security: Patch is for trusted URLs. Use PatchSafe for untrusted URLs.
func Patch(rawURL string, opts ...RequestOption) *Request { return restyimpl.Patch(rawURL, opts...) }

// PatchSafe creates a PATCH request with SSRF-oriented safety checks enabled.
func PatchSafe(rawURL string, opts ...RequestOption) *Request {
	return restyimpl.PatchSafe(rawURL, opts...)
}

// Head creates a HEAD request.
//
// Security: Head is for trusted URLs. Use HeadSafe for untrusted URLs.
func Head(rawURL string, opts ...RequestOption) *Request { return restyimpl.Head(rawURL, opts...) }

// HeadSafe creates a HEAD request with SSRF-oriented safety checks enabled.
func HeadSafe(rawURL string, opts ...RequestOption) *Request {
	return restyimpl.HeadSafe(rawURL, opts...)
}

// Options creates an OPTIONS request.
//
// Security: Options is for trusted URLs. Use OptionsSafe for untrusted URLs.
func Options(rawURL string, opts ...RequestOption) *Request {
	return restyimpl.Options(rawURL, opts...)
}

// OptionsSafe creates an OPTIONS request with SSRF-oriented safety checks enabled.
func OptionsSafe(rawURL string, opts ...RequestOption) *Request {
	return restyimpl.OptionsSafe(rawURL, opts...)
}

// NewRequest creates a request by method.
//
// Security: NewRequest is for trusted URLs unless callers provide WithURLPolicy
// with RejectPrivate enabled. Use NewSafeRequest for untrusted URLs.
func NewRequest(method Method, rawURL string, opts ...RequestOption) *Request {
	return restyimpl.NewRequest(method, rawURL, opts...)
}

// NewSafeRequest creates a request with SSRF-oriented safety checks enabled.
func NewSafeRequest(method Method, rawURL string, opts ...RequestOption) *Request {
	return restyimpl.NewSafeRequest(method, rawURL, opts...)
}

// NewIsolatedRequest creates a request without reading package-level global defaults.
func NewIsolatedRequest(method Method, rawURL string, opts ...RequestOption) *Request {
	return restyimpl.NewIsolatedRequest(method, rawURL, opts...)
}

// NewRequestWithConfig creates a request from an explicit global configuration snapshot.
//
// Security: NewRequestWithConfig is for trusted URLs unless callers provide
// WithURLPolicy with RejectPrivate enabled. Use NewSafeRequest for untrusted
// URLs.
func NewRequestWithConfig(method Method, rawURL string, cfg GlobalConfig, opts ...RequestOption) *Request {
	return restyimpl.NewRequestWithConfig(method, rawURL, cfg, opts...)
}

// WithGlobalConfig initializes request defaults from a captured global configuration snapshot.
func WithGlobalConfig(cfg GlobalConfig) RequestOption { return restyimpl.WithGlobalConfig(cfg) }

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

// WithTLSConfig sets a per-request TLS config. It is ignored when WithRestyClient is set.
func WithTLSConfig(cfg *tls.Config) RequestOption { return restyimpl.WithTLSConfig(cfg) }

// WithRestyClient sets a per-request resty client and takes precedence over WithTLSConfig.
func WithRestyClient(c *grestry.Client) RequestOption { return restyimpl.WithRestyClient(c) }

// WithRestyClientFactory sets a per-request resty client factory.
func WithRestyClientFactory(factory func() *grestry.Client) RequestOption {
	return restyimpl.WithRestyClientFactory(factory)
}

// ConfigureDefaultRestyClientProvider sets the provider used to create resty clients when no per-request client is set.
func ConfigureDefaultRestyClientProvider(provider func() *grestry.Client) {
	restyimpl.ConfigureDefaultRestyClientProvider(provider)
}

// ResetDefaultRestyClientProvider restores resty.New as the default client provider.
func ResetDefaultRestyClientProvider() { restyimpl.ResetDefaultRestyClientProvider() }

// WithUserAgent sets a per-request User-Agent.
func WithUserAgent(ua string) RequestOption { return restyimpl.WithUserAgent(ua) }

// WithCookieDisabled sets per-request cookie management behavior.
func WithCookieDisabled(disabled bool) RequestOption { return restyimpl.WithCookieDisabled(disabled) }

// WithContentType sets a per-request Content-Type at construction time.
func WithContentType(ct string) RequestOption { return restyimpl.WithContentType(ct) }

// WithCharset sets a per-request charset at construction time.
func WithCharset(charset string) RequestOption { return restyimpl.WithCharset(charset) }

// WithJSONMarshalFunc sets the JSON marshal provider used by request body encoding.
func WithJSONMarshalFunc(marshal func(any) ([]byte, error)) RequestOption {
	return restyimpl.WithJSONMarshalFunc(marshal)
}

// WithJSONUnmarshalFunc sets the JSON unmarshal provider used by response decoding.
func WithJSONUnmarshalFunc(unmarshal func([]byte, any) error) RequestOption {
	return restyimpl.WithJSONUnmarshalFunc(unmarshal)
}

// WithJSONDecodeReadAllFunc sets the reader used before custom JSON unmarshalling.
func WithJSONDecodeReadAllFunc(readAll func(io.Reader) ([]byte, error)) RequestOption {
	return restyimpl.WithJSONDecodeReadAllFunc(readAll)
}

// WithMaxDecodeBytes limits bytes read before custom JSON unmarshalling. Non-positive means unlimited.
func WithMaxDecodeBytes(maxBytes int64) RequestOption {
	return restyimpl.WithMaxDecodeBytes(maxBytes)
}

// WithMaxResponseBytes limits response bytes read into memory. Non-positive means unlimited.
func WithMaxResponseBytes(maxBytes int64) RequestOption {
	return restyimpl.WithMaxResponseBytes(maxBytes)
}

// WithURLPolicy sets SSRF-oriented validation for the request URL and redirect targets.
func WithURLPolicy(policy URLPolicy) RequestOption { return restyimpl.WithURLPolicy(policy) }

// WithAllowedHosts restricts Safe requests to the provided host names.
func WithAllowedHosts(hosts ...string) RequestOption { return restyimpl.WithAllowedHosts(hosts...) }

// WithLookupIP sets the host resolver used by SSRF-oriented URL validation.
func WithLookupIP(lookupIP func(context.Context, string) ([]net.IP, error)) RequestOption {
	return restyimpl.WithLookupIP(lookupIP)
}

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

// WithSaveStat sets the stat provider used to resolve directory destinations.
func WithSaveStat(stat func(string) (os.FileInfo, error)) SaveOption {
	return restyimpl.WithSaveStat(stat)
}

// WithSaveMkdirAll sets the directory creator used when saving responses.
func WithSaveMkdirAll(mkdirAll func(string, fs.FileMode) error) SaveOption {
	return restyimpl.WithSaveMkdirAll(mkdirAll)
}

// WithSaveOpenFile sets the file opener used when saving responses.
func WithSaveOpenFile(openFile func(string, int, fs.FileMode) (io.WriteCloser, error)) SaveOption {
	return restyimpl.WithSaveOpenFile(openFile)
}

// GetStringE sends a GET request and returns response body as string or an error.
func GetStringE(rawURL string) (string, error) { return GetStringEWithOptions(rawURL) }

// GetStringEWithOptions sends a GET request with options and returns response body as string or an error.
func GetStringEWithOptions(rawURL string, opts ...RequestOption) (string, error) {
	return restyimpl.GetStringEWithOptions(rawURL, opts...)
}

// GetStringSafeE sends a safe GET request and returns response body as string or an error.
func GetStringSafeE(rawURL string, opts ...RequestOption) (string, error) {
	return restyimpl.GetStringSafeE(rawURL, opts...)
}

// GetWithTimeoutE sends a GET request with a timeout and returns response body or an error.
func GetWithTimeoutE(rawURL string, timeout time.Duration) (string, error) {
	return GetWithTimeoutEWithOptions(rawURL, timeout)
}

// GetWithTimeoutEWithOptions sends a GET request with a timeout and custom options, returning body or error.
func GetWithTimeoutEWithOptions(rawURL string, timeout time.Duration, opts ...RequestOption) (string, error) {
	return restyimpl.GetWithTimeoutEWithOptions(rawURL, timeout, opts...)
}

// GetWithParamsE sends a GET request with form parameters and returns response body or an error.
func GetWithParamsE(rawURL string, params map[string]any) (string, error) {
	return GetWithParamsEWithOptions(rawURL, params)
}

// GetWithParamsEWithOptions sends a GET request with form parameters and custom options, returning body or error.
func GetWithParamsEWithOptions(rawURL string, params map[string]any, opts ...RequestOption) (string, error) {
	return restyimpl.GetWithParamsEWithOptions(rawURL, params, opts...)
}

// PostStringE posts a string body and returns response body or an error.
func PostStringE(rawURL, body string) (string, error) { return PostStringEWithOptions(rawURL, body) }

// PostStringEWithOptions posts a string body with options and returns response body or an error.
func PostStringEWithOptions(rawURL, body string, opts ...RequestOption) (string, error) {
	return restyimpl.PostStringEWithOptions(rawURL, body, opts...)
}

// PostStringSafeE posts a string body with SSRF-oriented safety checks enabled.
func PostStringSafeE(rawURL, body string, opts ...RequestOption) (string, error) {
	return restyimpl.PostStringSafeE(rawURL, body, opts...)
}

// PostFormE posts form parameters and returns response body or an error.
func PostFormE(rawURL string, params map[string]any) (string, error) {
	return PostFormEWithOptions(rawURL, params)
}

// PostFormEWithOptions posts form parameters with options and returns response body or an error.
func PostFormEWithOptions(rawURL string, params map[string]any, opts ...RequestOption) (string, error) {
	return restyimpl.PostFormEWithOptions(rawURL, params, opts...)
}

// PostFormSafeE posts form parameters with SSRF-oriented safety checks enabled.
func PostFormSafeE(rawURL string, params map[string]any, opts ...RequestOption) (string, error) {
	return restyimpl.PostFormSafeE(rawURL, params, opts...)
}

// PostJSONE posts JSON body and returns response body or an error.
func PostJSONE(rawURL, jsonStr string) (string, error) { return PostJSONEWithOptions(rawURL, jsonStr) }

// PostJSONEWithOptions posts JSON body with options and returns response body or an error.
func PostJSONEWithOptions(rawURL, jsonStr string, opts ...RequestOption) (string, error) {
	return restyimpl.PostJSONEWithOptions(rawURL, jsonStr, opts...)
}

// PostJSONSafeE posts JSON body with SSRF-oriented safety checks enabled.
func PostJSONSafeE(rawURL, jsonStr string, opts ...RequestOption) (string, error) {
	return restyimpl.PostJSONSafeE(rawURL, jsonStr, opts...)
}

// Download downloads rawURL into w.
func Download(rawURL string, w io.Writer) (int64, error) { return DownloadWithOptions(rawURL, w) }

// DownloadWithOptions downloads rawURL into w with per-request options.
func DownloadWithOptions(rawURL string, w io.Writer, opts ...RequestOption) (int64, error) {
	return restyimpl.DownloadWithOptions(rawURL, w, opts...)
}

// DownloadSafe downloads rawURL into w with SSRF-oriented safety checks enabled.
func DownloadSafe(rawURL string, w io.Writer, opts ...RequestOption) (int64, error) {
	return restyimpl.DownloadSafe(rawURL, w, opts...)
}

// DownloadFile downloads rawURL to dest.
func DownloadFile(rawURL, dest string, opts ...SaveOption) (int64, error) {
	return DownloadFileWithOptions(rawURL, dest, nil, opts...)
}

// DownloadFileWithOptions downloads rawURL to dest with per-request and per-save options.
func DownloadFileWithOptions(rawURL, dest string, requestOpts []RequestOption, saveOpts ...SaveOption) (int64, error) {
	return restyimpl.DownloadFileWithOptions(rawURL, dest, requestOpts, saveOpts...)
}

// DownloadFileSafe downloads rawURL to dest with SSRF-oriented safety checks enabled.
func DownloadFileSafe(rawURL, dest string, opts ...SaveOption) (int64, error) {
	return DownloadFileSafeWithOptions(rawURL, dest, nil, opts...)
}

// DownloadFileSafeWithOptions downloads rawURL to dest with SSRF-oriented safety checks enabled.
func DownloadFileSafeWithOptions(rawURL, dest string, requestOpts []RequestOption, saveOpts ...SaveOption) (int64, error) {
	return restyimpl.DownloadFileSafeWithOptions(rawURL, dest, requestOpts, saveOpts...)
}

// DownloadBytesE downloads and returns bytes or an error.
func DownloadBytesE(rawURL string) ([]byte, error) { return DownloadBytesEWithOptions(rawURL) }

// DownloadBytesEWithOptions downloads and returns bytes with per-request options or an error.
func DownloadBytesEWithOptions(rawURL string, opts ...RequestOption) ([]byte, error) {
	return restyimpl.DownloadBytesEWithOptions(rawURL, opts...)
}

// DownloadBytesSafeE downloads and returns bytes with SSRF-oriented safety checks enabled.
func DownloadBytesSafeE(rawURL string, opts ...RequestOption) ([]byte, error) {
	return restyimpl.DownloadBytesSafeE(rawURL, opts...)
}

// DownloadStringE downloads remote text and returns an error on request failure.
func DownloadStringE(rawURL, customCharset string) (string, error) {
	return DownloadStringEWithOptions(rawURL, customCharset)
}

// DownloadStringEWithOptions downloads remote text with per-request options and returns an error on failure.
func DownloadStringEWithOptions(rawURL, customCharset string, opts ...RequestOption) (string, error) {
	return restyimpl.DownloadStringEWithOptions(rawURL, customCharset, opts...)
}

// DownloadStringSafeE downloads remote text with SSRF-oriented safety checks enabled.
func DownloadStringSafeE(rawURL, customCharset string, opts ...RequestOption) (string, error) {
	return restyimpl.DownloadStringSafeE(rawURL, customCharset, opts...)
}

// SetGlobalTimeout sets the global HTTP timeout.
func SetGlobalTimeout(d time.Duration) { restyimpl.SetGlobalTimeout(d) }

// GetGlobalTimeout returns the global HTTP timeout.
func GetGlobalTimeout() time.Duration { return restyimpl.GetGlobalTimeout() }

// SetGlobalMaxRedirects sets the global maximum redirect count.
func SetGlobalMaxRedirects(n int) { restyimpl.SetGlobalMaxRedirects(n) }

// GetGlobalMaxRedirects returns the global maximum redirect count.
func GetGlobalMaxRedirects() int { return restyimpl.GetGlobalMaxRedirects() }

// SetGlobalMaxResponseBytes sets the global maximum response bytes read into memory.
// Non-positive values disable the limit.
func SetGlobalMaxResponseBytes(n int64) { restyimpl.SetGlobalMaxResponseBytes(n) }

// GetGlobalMaxResponseBytes returns the global maximum response bytes read into memory.
func GetGlobalMaxResponseBytes() int64 { return restyimpl.GetGlobalMaxResponseBytes() }

// SetGlobalFollowRedirects sets whether redirects are followed globally.
func SetGlobalFollowRedirects(b bool) { restyimpl.SetGlobalFollowRedirects(b) }

// GetGlobalFollowRedirects reports whether redirects are followed globally.
func GetGlobalFollowRedirects() bool { return restyimpl.GetGlobalFollowRedirects() }

// SetGlobalUserAgent sets the global default User-Agent.
func SetGlobalUserAgent(ua string) { restyimpl.SetGlobalUserAgent(ua) }

// GetGlobalUserAgent returns the global default User-Agent.
func GetGlobalUserAgent() string { return restyimpl.GetGlobalUserAgent() }

// SnapshotGlobalConfig returns a copy of the package-level resty defaults.
func SnapshotGlobalConfig() GlobalConfig { return restyimpl.SnapshotGlobalConfig() }

// ResetGlobalConfig restores package-level resty defaults, including headers and cookies.
func ResetGlobalConfig() { restyimpl.ResetGlobalConfig() }

// ConfigureGlobalConfig replaces package-level resty defaults with cfg.
func ConfigureGlobalConfig(cfg GlobalConfig) { restyimpl.ConfigureGlobalConfig(cfg) }

// WithScopedGlobalConfig runs fn with cfg installed as package-level resty defaults,
// then restores the previous defaults.
func WithScopedGlobalConfig(cfg GlobalConfig, fn func()) { restyimpl.WithScopedGlobalConfig(cfg, fn) }

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

// WithCharsetRegexp sets the regexp used by GetCharsetFromContentTypeWithOptions.
func WithCharsetRegexp(re *regexp.Regexp) CharsetOption { return restyimpl.WithCharsetRegexp(re) }

// WithMetaCharsetRegexp sets the regexp used by GetCharsetFromHTMLWithOptions.
func WithMetaCharsetRegexp(re *regexp.Regexp) CharsetOption {
	return restyimpl.WithMetaCharsetRegexp(re)
}

// IsHTTPS reports whether the given URL is https.
func IsHTTPS(rawURL string) bool { return restyimpl.IsHTTPS(rawURL) }

// IsHTTP reports whether the given URL is http.
func IsHTTP(rawURL string) bool { return restyimpl.IsHTTP(rawURL) }

// ToParams converts a map to a URL query string.
func ToParams(m map[string]any) string { return restyimpl.ToParams(m) }

// URLWithForm appends form values to a URL.
func URLWithForm(rawURL string, form map[string]any) string {
	return restyimpl.URLWithForm(rawURL, form)
}

// BuildContentType builds a Content-Type string with charset.
func BuildContentType(contentType, charset string) string {
	return restyimpl.BuildContentType(contentType, charset)
}

// GuessContentType guesses Content-Type from the body.
func GuessContentType(body string) ContentType { return restyimpl.GuessContentType(body) }

// IsDefaultContentType reports whether the value is a default Content-Type.
func IsDefaultContentType(contentType string) bool {
	return restyimpl.IsDefaultContentType(contentType)
}

// IsFormURLEncoded reports whether the value is application/x-www-form-urlencoded.
func IsFormURLEncoded(contentType string) bool { return restyimpl.IsFormURLEncoded(contentType) }

// NewHTTPError creates an HTTP error.
func NewHTTPError(msg string, cause error) *Error { return restyimpl.NewHTTPError(msg, cause) }

// HTTPErrorf creates an HTTP error with a formatted message.
func HTTPErrorf(format string, args ...any) *Error { return restyimpl.HTTPErrorf(format, args...) }

// GetCharsetFromContentType extracts charset from Content-Type.
func GetCharsetFromContentType(ct string) string { return restyimpl.GetCharsetFromContentType(ct) }

// GetCharsetFromContentTypeWithOptions extracts charset from Content-Type with options.
func GetCharsetFromContentTypeWithOptions(ct string, opts ...CharsetOption) string {
	return restyimpl.GetCharsetFromContentTypeWithOptions(ct, opts...)
}

// GetCharsetFromHTML extracts charset from HTML meta tags.
func GetCharsetFromHTML(html string) string { return restyimpl.GetCharsetFromHTML(html) }

// GetCharsetFromHTMLWithOptions extracts charset from HTML meta tags with options.
func GetCharsetFromHTMLWithOptions(html string, opts ...CharsetOption) string {
	return restyimpl.GetCharsetFromHTMLWithOptions(html, opts...)
}

// GetMimeType returns the MIME type by file extension.
func GetMimeType(filename string) string { return restyimpl.GetMimeType(filename) }
