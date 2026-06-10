package vurl

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	urlimpl "github.com/imajinyun/go-knifer/internal/url"
)

const (
	// ClasspathURLPrefix is the pseudo classpath URL prefix.
	ClasspathURLPrefix = urlimpl.ClasspathURLPrefix
	// FileURLPrefix is the file URL prefix.
	FileURLPrefix = urlimpl.FileURLPrefix
	// JarURLPrefix is the jar URL prefix.
	JarURLPrefix = urlimpl.JarURLPrefix
	// WarURLPrefix is the war URL prefix.
	WarURLPrefix = urlimpl.WarURLPrefix
	// URLProtocolFile is the file URL scheme.
	URLProtocolFile = urlimpl.URLProtocolFile
	// URLProtocolJar is the jar URL scheme.
	URLProtocolJar = urlimpl.URLProtocolJar
	// URLProtocolZip is the zip URL scheme.
	URLProtocolZip = urlimpl.URLProtocolZip
	// URLProtocolWSJar is the WebSphere jar URL scheme.
	URLProtocolWSJar = urlimpl.URLProtocolWSJar
	// URLProtocolVFSZip is the JBoss VFS zip URL scheme.
	URLProtocolVFSZip = urlimpl.URLProtocolVFSZip
	// URLProtocolVFSFile is the JBoss VFS file URL scheme.
	URLProtocolVFSFile = urlimpl.URLProtocolVFSFile
	// URLProtocolVFS is the JBoss VFS URL scheme.
	URLProtocolVFS = urlimpl.URLProtocolVFS
	// JarURLSeparator separates an archive URL from an entry path.
	JarURLSeparator = urlimpl.JarURLSeparator
	// WarURLSeparator separates a war URL from an entry path.
	WarURLSeparator = urlimpl.WarURLSeparator
	// DefaultMaxBytes is the default response body limit used by OpenSafe.
	DefaultMaxBytes = urlimpl.DefaultMaxBytes
)

// URLBuilder builds URLs from scheme, host, path, query, and fragment parts.
type URLBuilder = urlimpl.URLBuilder

// ResourceOption customizes URL resource helpers.
type ResourceOption = urlimpl.ResourceOption

// NormalizeOption customizes URL normalization.
type NormalizeOption = urlimpl.NormalizeOption

// DecodeOption customizes DecodeWithOptions.
type DecodeOption = urlimpl.DecodeOption

// EncodeOption customizes URL encoding helpers.
type EncodeOption = urlimpl.EncodeOption

// NewURLBuilder creates an empty URL builder.
func NewURLBuilder() *URLBuilder { return urlimpl.NewURLBuilder() }

// WithContext sets the context used by HTTP resource requests.
func WithContext(ctx context.Context) ResourceOption { return urlimpl.WithContext(ctx) }

// WithHTTPClient sets the HTTP client used by HTTP resource requests.
func WithHTTPClient(client *http.Client) ResourceOption { return urlimpl.WithHTTPClient(client) }

// WithHeader adds an HTTP header to HTTP resource requests.
func WithHeader(name, value string) ResourceOption { return urlimpl.WithHeader(name, value) }

// WithHeaders adds HTTP headers to HTTP resource requests.
func WithHeaders(headers http.Header) ResourceOption { return urlimpl.WithHeaders(headers) }

// WithTimeout bounds HTTP resource requests.
func WithTimeout(timeout time.Duration) ResourceOption { return urlimpl.WithTimeout(timeout) }

// WithCheckStatus makes HTTP resource helpers reject non-2xx responses.
func WithCheckStatus(check bool) ResourceOption { return urlimpl.WithCheckStatus(check) }

// WithOpenFile sets the file opener used by local file resource helpers.
func WithOpenFile(openFile func(string) (io.ReadCloser, error)) ResourceOption {
	return urlimpl.WithOpenFile(openFile)
}

// WithStat sets the stat provider used by local file resource helpers.
func WithStat(stat func(string) (os.FileInfo, error)) ResourceOption { return urlimpl.WithStat(stat) }

// WithRequestFactory sets the HTTP request factory used by resource helpers.
func WithRequestFactory(factory func(context.Context, string, string) (*http.Request, error)) ResourceOption {
	return urlimpl.WithRequestFactory(factory)
}

// WithLookupIP sets the host resolver used by SSRF-oriented URL validation and safe dialing.
func WithLookupIP(lookupIP func(context.Context, string) ([]net.IP, error)) ResourceOption {
	return urlimpl.WithLookupIP(lookupIP)
}

// WithMaxBytes limits how many response body bytes OpenWithOptions may read.
func WithMaxBytes(n int64) ResourceOption { return urlimpl.WithMaxBytes(n) }

// WithAllowedSchemes restricts resource helpers to the provided URL schemes.
func WithAllowedSchemes(schemes ...string) ResourceOption {
	return urlimpl.WithAllowedSchemes(schemes...)
}

// WithAllowedHosts restricts HTTP(S) resource helpers to the provided host names.
func WithAllowedHosts(hosts ...string) ResourceOption { return urlimpl.WithAllowedHosts(hosts...) }

// WithRejectPrivateHosts rejects localhost, loopback, private, and link-local HTTP(S) hosts unless explicitly allowed.
func WithRejectPrivateHosts(reject bool) ResourceOption {
	return urlimpl.WithRejectPrivateHosts(reject)
}

// WithAllowLocalFiles controls whether file URLs and plain filesystem paths are allowed.
func WithAllowLocalFiles(allow bool) ResourceOption { return urlimpl.WithAllowLocalFiles(allow) }

// WithDefaultScheme sets the scheme used when NormalizeWithOptions receives a URL without scheme.
func WithDefaultScheme(scheme string) NormalizeOption { return urlimpl.WithDefaultScheme(scheme) }

// WithEncodePath controls whether NormalizeUsingOptions escapes the normalized path.
func WithEncodePath(encode bool) NormalizeOption { return urlimpl.WithEncodePath(encode) }

// WithReplaceSlash controls whether NormalizeUsingOptions collapses repeated slashes in the path.
func WithReplaceSlash(replace bool) NormalizeOption { return urlimpl.WithReplaceSlash(replace) }

// WithPlusAsSpace controls whether plus signs are decoded as spaces.
func WithPlusAsSpace(plusToSpace bool) DecodeOption { return urlimpl.WithPlusAsSpace(plusToSpace) }

// WithQueryEscapeFunc sets the query/form escaping provider.
func WithQueryEscapeFunc(escape func(string) string) EncodeOption {
	return urlimpl.WithQueryEscapeFunc(escape)
}

// WithPathEscapeFunc sets the path segment escaping provider.
func WithPathEscapeFunc(escape func(string) string) EncodeOption {
	return urlimpl.WithPathEscapeFunc(escape)
}

// NewHTTPURLBuilder creates an HTTP URL builder.
func NewHTTPURLBuilder(host string) *URLBuilder { return urlimpl.NewHTTPURLBuilder(host) }

// ParseURLBuilder parses raw into a URL builder.
func ParseURLBuilder(raw string) (*URLBuilder, error) { return urlimpl.ParseURLBuilder(raw) }

// Parse parses raw into a URL. Empty input returns nil without error.
func Parse(raw string) (*url.URL, error) { return urlimpl.Parse(raw) }

// ParseHTTP parses raw after encoding blank characters.
func ParseHTTP(raw string) (*url.URL, error) { return urlimpl.ParseHTTP(raw) }

// StringURI returns a string-scheme URI for content.
func StringURI(content string) string { return urlimpl.StringURI(content) }

// EncodeBlank encodes all Unicode blank characters as %20.
func EncodeBlank(raw string) string { return urlimpl.EncodeBlank(raw) }

// FileURL converts a filesystem path to a file URL.
func FileURL(path string) (*url.URL, error) { return urlimpl.FileURL(path) }

// FileURLs converts filesystem paths to file URLs.
func FileURLs(paths ...string) ([]*url.URL, error) { return urlimpl.FileURLs(paths...) }

// Host returns a URL that keeps only scheme and host.
func Host(u *url.URL) *url.URL { return urlimpl.Host(u) }

// Complete resolves relativePath against baseURL and returns the absolute URL string.
func Complete(baseURL, relativePath string) (string, error) {
	return urlimpl.Complete(baseURL, relativePath)
}

// Encode escapes a string for URL query components.
func Encode(s string) string { return urlimpl.Encode(s) }

// EncodeWithOptions escapes a string for URL query components with custom providers.
func EncodeWithOptions(s string, opts ...EncodeOption) string {
	return urlimpl.EncodeWithOptions(s, opts...)
}

// URLEncode escapes a string for URL query components.
func URLEncode(s string) string { return urlimpl.URLEncode(s) }

// URLEncodeWithOptions escapes a string for URL query components with custom providers.
func URLEncodeWithOptions(s string, opts ...EncodeOption) string {
	return urlimpl.URLEncodeWithOptions(s, opts...)
}

// Decode unescapes a URL query component and converts plus signs to spaces.
func Decode(s string) (string, error) { return urlimpl.Decode(s) }

// URLDecode unescapes a URL query component and converts plus signs to spaces.
func URLDecode(s string) (string, error) { return urlimpl.URLDecode(s) }

// DecodePlus unescapes percent-encoded text and controls whether plus signs become spaces.
func DecodePlus(s string, plusToSpace bool) (string, error) {
	return urlimpl.DecodePlus(s, plusToSpace)
}

// DecodeWithOptions unescapes percent-encoded text with custom decoding behavior.
func DecodeWithOptions(s string, opts ...DecodeOption) (string, error) {
	return urlimpl.DecodeWithOptions(s, opts...)
}

// DecodeForPath unescapes percent-encoded path text without converting plus signs to spaces.
func DecodeForPath(s string) (string, error) { return urlimpl.DecodeForPath(s) }

// EncodeAll percent-encodes every non-unreserved character.
func EncodeAll(s string) string { return urlimpl.EncodeAll(s) }

// EncodeQuery escapes text for query/form usage. Spaces are encoded as '+'.
func EncodeQuery(s string) string { return urlimpl.EncodeQuery(s) }

// EncodeQueryWithOptions escapes text for query/form usage with custom providers.
func EncodeQueryWithOptions(s string, opts ...EncodeOption) string {
	return urlimpl.EncodeQueryWithOptions(s, opts...)
}

// EncodePathSegment escapes one path segment, including slash characters.
func EncodePathSegment(s string) string { return urlimpl.EncodePathSegment(s) }

// EncodePathSegmentWithOptions escapes one path segment with custom providers.
func EncodePathSegmentWithOptions(s string, opts ...EncodeOption) string {
	return urlimpl.EncodePathSegmentWithOptions(s, opts...)
}

// EncodePath escapes each path segment and keeps slash separators.
func EncodePath(s string) string { return urlimpl.EncodePath(s) }

// EncodeFragment escapes URL fragment text.
func EncodeFragment(s string) string { return urlimpl.EncodeFragment(s) }

// FormURLEncode escapes text for application/x-www-form-urlencoded usage.
func FormURLEncode(s string) string { return urlimpl.FormURLEncode(s) }

// FormURLEncodeWithOptions escapes text for application/x-www-form-urlencoded usage with custom providers.
func FormURLEncodeWithOptions(s string, opts ...EncodeOption) string {
	return urlimpl.FormURLEncodeWithOptions(s, opts...)
}

// Path returns the decoded path part of raw.
func Path(raw string) (string, error) { return urlimpl.Path(raw) }

// DecodedPath returns u's path after percent-decoding.
func DecodedPath(u *url.URL) string { return urlimpl.DecodedPath(u) }

// ToURI parses location as a URI. If encode is true, blank characters are encoded first.
func ToURI(location string, encode bool) (*url.URL, error) { return urlimpl.ToURI(location, encode) }

// IsFileURL reports whether u uses a file-like scheme.
func IsFileURL(u *url.URL) bool { return urlimpl.IsFileURL(u) }

// IsJarURL reports whether u uses an archive-like scheme.
func IsJarURL(u *url.URL) bool { return urlimpl.IsJarURL(u) }

// IsJarFileURL reports whether u is a file URL ending with .jar.
func IsJarFileURL(u *url.URL) bool { return urlimpl.IsJarFileURL(u) }

// Open opens a URL resource. It supports http, https, file URLs, and plain file paths.
//
// Security: Open is for trusted resource locations only because it may access
// local files and private network addresses. Use OpenSafe for resource locations
// that may cross a user, configuration, or network trust boundary.
func Open(raw string) (io.ReadCloser, error) { return urlimpl.Open(raw) }

// OpenWithOptions opens a URL resource with per-call options.
//
// Security: OpenWithOptions is for trusted resource locations unless options
// explicitly restrict schemes, local files, redirects, and private hosts. Prefer
// OpenSafeWithOptions for untrusted input.
func OpenWithOptions(raw string, opts ...ResourceOption) (io.ReadCloser, error) {
	return urlimpl.OpenWithOptions(raw, opts...)
}

// OpenSafe opens an HTTP(S) URL with secure defaults for untrusted input.
func OpenSafe(raw string) (io.ReadCloser, error) { return urlimpl.OpenSafe(raw) }

// OpenSafeWithOptions opens an HTTP(S) URL with secure defaults for untrusted input.
func OpenSafeWithOptions(raw string, opts ...ResourceOption) (io.ReadCloser, error) {
	return urlimpl.OpenSafeWithOptions(raw, opts...)
}

// ContentLength returns the resource content length. Unknown lengths return -1.
//
// Security: ContentLength is for trusted resource locations only because it may
// access local files and private network addresses. Use ContentLengthSafe for
// untrusted input.
func ContentLength(raw string) (int64, error) { return urlimpl.ContentLength(raw) }

// ContentLengthWithOptions returns the resource content length with per-call options.
//
// Security: ContentLengthWithOptions is for trusted resource locations unless
// options explicitly restrict schemes, local files, redirects, and private hosts.
// Prefer ContentLengthSafeWithOptions for untrusted input.
func ContentLengthWithOptions(raw string, opts ...ResourceOption) (int64, error) {
	return urlimpl.ContentLengthWithOptions(raw, opts...)
}

// ContentLengthSafe returns an HTTP(S) resource content length with secure defaults for untrusted input.
func ContentLengthSafe(raw string) (int64, error) { return urlimpl.ContentLengthSafe(raw) }

// ContentLengthSafeWithOptions returns an HTTP(S) resource content length with secure defaults for untrusted input.
func ContentLengthSafeWithOptions(raw string, opts ...ResourceOption) (int64, error) {
	return urlimpl.ContentLengthSafeWithOptions(raw, opts...)
}

// Size returns the resource size.
//
// Security: Size follows ContentLength and is for trusted resource locations
// only. Use ContentLengthSafe for untrusted input.
func Size(raw string) (int64, error) { return urlimpl.Size(raw) }

// SizeWithOptions returns the resource size with per-call options.
//
// Security: SizeWithOptions follows ContentLengthWithOptions and is for trusted
// resource locations unless safe options are enabled.
func SizeWithOptions(raw string, opts ...ResourceOption) (int64, error) {
	return urlimpl.SizeWithOptions(raw, opts...)
}

// Normalize normalizes a URL string by adding a default scheme and cleaning slashes.
func Normalize(raw string, encodePath, replaceSlash bool) string {
	return urlimpl.Normalize(raw, encodePath, replaceSlash)
}

// NormalizeWithOptions normalizes a URL string with per-call options.
func NormalizeWithOptions(raw string, encodePath, replaceSlash bool, opts ...NormalizeOption) string {
	return urlimpl.NormalizeWithOptions(raw, encodePath, replaceSlash, opts...)
}

// NormalizeUsingOptions normalizes a URL string using only functional options for optional behavior.
func NormalizeUsingOptions(raw string, opts ...NormalizeOption) string {
	return urlimpl.NormalizeUsingOptions(raw, opts...)
}

// BuildQuery converts a map to a URL query string.
func BuildQuery(paramMap map[string]any) string { return urlimpl.BuildQuery(paramMap) }

// EncodeQueryMap converts a map to a URL query string.
func EncodeQueryMap(m map[string]any) string { return urlimpl.EncodeQueryMap(m) }

// EncodeParams encodes the query part of rawURL and leaves URLs without query unchanged.
func EncodeParams(rawURL string) string { return urlimpl.EncodeParams(rawURL) }

// DecodeQueryFirst parses a query string into a single-value map.
func DecodeQueryFirst(paramsStr string) map[string]string { return urlimpl.DecodeQueryFirst(paramsStr) }

// DecodeQuery parses a query string into a multi-value map.
func DecodeQuery(paramsStr string) map[string][]string { return urlimpl.DecodeQuery(paramsStr) }

// AppendQuery appends form values to rawURL.
func AppendQuery(rawURL string, form map[string]any) string { return urlimpl.AppendQuery(rawURL, form) }

// IsHTTP reports whether raw uses the http scheme prefix.
func IsHTTP(raw string) bool { return urlimpl.IsHTTP(raw) }

// IsHTTPS reports whether raw uses the https scheme prefix.
func IsHTTPS(raw string) bool { return urlimpl.IsHTTPS(raw) }

// IsHTTPURL reports whether raw is an absolute http URL with a host.
func IsHTTPURL(raw string) bool { return urlimpl.IsHTTPURL(raw) }

// IsHTTPSURL reports whether raw is an absolute https URL with a host.
func IsHTTPSURL(raw string) bool { return urlimpl.IsHTTPSURL(raw) }

// IsWebURL reports whether raw is an absolute http or https URL with a host.
func IsWebURL(raw string) bool { return urlimpl.IsWebURL(raw) }

// IsAbsoluteURL reports whether raw is an absolute URL with scheme and host.
func IsAbsoluteURL(raw string) bool { return urlimpl.IsAbsoluteURL(raw) }

// DataURIBase64 builds a base64 Data URI string.
func DataURIBase64(mimeType, data string) string { return urlimpl.DataURIBase64(mimeType, data) }

// DataURI builds a Data URI string.
func DataURI(mimeType, charset, encoding, data string) string {
	return urlimpl.DataURI(mimeType, charset, encoding, data)
}
