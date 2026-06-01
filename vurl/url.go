package vurl

import (
	"io"
	"net/url"

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
)

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
// Direct query escaping convenience is also available via vcodec.URLEncode.
func Encode(s string) string { return urlimpl.Encode(s) }

// Decode unescapes a URL query component and converts plus signs to spaces.
// Direct query unescaping convenience is also available via vcodec.URLDecode.
func Decode(s string) (string, error) { return urlimpl.Decode(s) }

// DecodePlus unescapes percent-encoded text and controls whether plus signs become spaces.
func DecodePlus(s string, plusToSpace bool) (string, error) {
	return urlimpl.DecodePlus(s, plusToSpace)
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
func Open(raw string) (io.ReadCloser, error) { return urlimpl.Open(raw) }

// ContentLength returns the resource content length. Unknown lengths return -1.
func ContentLength(raw string) (int64, error) { return urlimpl.ContentLength(raw) }

// Size returns the resource size.
func Size(raw string) (int64, error) { return urlimpl.Size(raw) }

// Normalize normalizes a URL string by adding a default scheme and cleaning slashes.
func Normalize(raw string, encodePath, replaceSlash bool) string {
	return urlimpl.Normalize(raw, encodePath, replaceSlash)
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
