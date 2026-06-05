package url

import (
	"context"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"
)

const (
	// ClasspathURLPrefix is the pseudo classpath URL prefix.
	ClasspathURLPrefix = "classpath:"
	// FileURLPrefix is the file URL prefix.
	FileURLPrefix = "file:"
	// JarURLPrefix is the jar URL prefix.
	JarURLPrefix = "jar:"
	// WarURLPrefix is the war URL prefix.
	WarURLPrefix = "war:"

	// URLProtocolFile is the file URL scheme.
	URLProtocolFile = "file"
	// URLProtocolJar is the jar URL scheme.
	URLProtocolJar = "jar"
	// URLProtocolZip is the zip URL scheme.
	URLProtocolZip = "zip"
	// URLProtocolWSJar is the WebSphere jar URL scheme.
	URLProtocolWSJar = "wsjar"
	// URLProtocolVFSZip is the JBoss VFS zip URL scheme.
	URLProtocolVFSZip = "vfszip"
	// URLProtocolVFSFile is the JBoss VFS file URL scheme.
	URLProtocolVFSFile = "vfsfile"
	// URLProtocolVFS is the JBoss VFS URL scheme.
	URLProtocolVFS = "vfs"

	// JarURLSeparator separates an archive URL from an entry path.
	JarURLSeparator = "!/"
	// WarURLSeparator separates a war URL from an entry path.
	WarURLSeparator = "*/"
)

type resourceConfig struct {
	ctx         context.Context
	client      *http.Client
	headers     http.Header
	timeout     time.Duration
	checkStatus bool
}

// ResourceOption customizes URL resource helpers such as OpenWithOptions and ContentLengthWithOptions.
type ResourceOption func(*resourceConfig)

// WithContext sets the context used by HTTP resource requests.
func WithContext(ctx context.Context) ResourceOption { return func(c *resourceConfig) { c.ctx = ctx } }

// WithHTTPClient sets the HTTP client used by HTTP resource requests.
func WithHTTPClient(client *http.Client) ResourceOption {
	return func(c *resourceConfig) { c.client = client }
}

// WithHeader adds an HTTP header to HTTP resource requests.
func WithHeader(name, value string) ResourceOption {
	return func(c *resourceConfig) {
		if c.headers == nil {
			c.headers = make(http.Header)
		}
		c.headers.Add(name, value)
	}
}

// WithHeaders adds HTTP headers to HTTP resource requests.
func WithHeaders(headers http.Header) ResourceOption {
	return func(c *resourceConfig) {
		if c.headers == nil {
			c.headers = make(http.Header)
		}
		for key, values := range headers {
			for _, value := range values {
				c.headers.Add(key, value)
			}
		}
	}
}

// WithTimeout bounds HTTP resource requests. Non-positive values mean no extra timeout.
func WithTimeout(timeout time.Duration) ResourceOption {
	return func(c *resourceConfig) { c.timeout = timeout }
}

// WithCheckStatus makes HTTP resource helpers reject non-2xx responses.
func WithCheckStatus(check bool) ResourceOption {
	return func(c *resourceConfig) { c.checkStatus = check }
}

func applyResourceOptions(opts []ResourceOption) resourceConfig {
	cfg := resourceConfig{ctx: context.Background(), client: http.DefaultClient}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.ctx == nil {
		cfg.ctx = context.Background()
	}
	if cfg.client == nil {
		cfg.client = http.DefaultClient
	}
	return cfg
}

type normalizeConfig struct {
	defaultScheme string
}

// NormalizeOption customizes URL normalization.
type NormalizeOption func(*normalizeConfig)

// WithDefaultScheme sets the scheme used when NormalizeWithOptions receives a URL without scheme.
func WithDefaultScheme(scheme string) NormalizeOption {
	return func(c *normalizeConfig) { c.defaultScheme = scheme }
}

func applyNormalizeOptions(opts []NormalizeOption) normalizeConfig {
	cfg := normalizeConfig{defaultScheme: "http"}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	cfg.defaultScheme = strings.TrimSpace(cfg.defaultScheme)
	if cfg.defaultScheme == "" {
		cfg.defaultScheme = "http"
	}
	cfg.defaultScheme = strings.TrimSuffix(cfg.defaultScheme, "://")
	cfg.defaultScheme = strings.TrimSuffix(cfg.defaultScheme, ":")
	return cfg
}

// Parse parses raw into a URL. Empty input returns nil without error.
func Parse(raw string) (*neturl.URL, error) {
	if raw == "" {
		return nil, nil
	}
	return neturl.Parse(strings.TrimSpace(raw))
}

// ParseHTTP parses raw after encoding blank characters.
func ParseHTTP(raw string) (*neturl.URL, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, fmt.Errorf("url is blank")
	}
	u, err := neturl.Parse(EncodeBlank(raw))
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("invalid absolute url: %s", raw)
	}
	return u, nil
}

// StringURI returns a string-scheme URI for content.
func StringURI(content string) string {
	if content == "" {
		return ""
	}
	if strings.HasPrefix(content, "string:///") {
		return content
	}
	return "string:///" + content
}

// EncodeBlank encodes all Unicode blank characters as %20.
func EncodeBlank(raw string) string {
	if raw == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(raw))
	for _, r := range raw {
		if unicode.IsSpace(r) {
			b.WriteString("%20")
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// FileURL converts a filesystem path to a file URL.
func FileURL(path string) (*neturl.URL, error) {
	if path == "" {
		return nil, fmt.Errorf("file path is blank")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	return &neturl.URL{Scheme: URLProtocolFile, Path: filepath.ToSlash(abs)}, nil
}

// FileURLs converts filesystem paths to file URLs.
func FileURLs(paths ...string) ([]*neturl.URL, error) {
	urls := make([]*neturl.URL, 0, len(paths))
	for _, path := range paths {
		u, err := FileURL(path)
		if err != nil {
			return nil, err
		}
		urls = append(urls, u)
	}
	return urls, nil
}

// Host returns a URL that keeps only scheme and host.
func Host(u *neturl.URL) *neturl.URL {
	if u == nil {
		return nil
	}
	return &neturl.URL{Scheme: u.Scheme, Host: u.Host}
}

// Complete resolves relativePath against baseURL and returns the absolute URL string.
func Complete(baseURL, relativePath string) (string, error) {
	baseURL = Normalize(baseURL, false, false)
	if strings.TrimSpace(baseURL) == "" {
		return "", nil
	}
	base, err := neturl.Parse(baseURL)
	if err != nil {
		return "", err
	}
	rel, err := neturl.Parse(relativePath)
	if err != nil {
		return "", err
	}
	return base.ResolveReference(rel).String(), nil
}

// Encode escapes a string for URL query components.
func Encode(s string) string { return neturl.QueryEscape(s) }

// URLEncode escapes a string for URL query components.
func URLEncode(s string) string { return Encode(s) }

// Decode unescapes a URL query component and converts plus signs to spaces.
func Decode(s string) (string, error) { return DecodePlus(s, true) }

// URLDecode unescapes a URL query component and converts plus signs to spaces.
func URLDecode(s string) (string, error) { return Decode(s) }

// DecodePlus unescapes percent-encoded text and controls whether plus signs become spaces.
func DecodePlus(s string, plusToSpace bool) (string, error) {
	if plusToSpace {
		return neturl.QueryUnescape(s)
	}
	return neturl.PathUnescape(strings.ReplaceAll(s, "+", "%2B"))
}

// Path returns the decoded path part of raw.
func Path(raw string) (string, error) {
	u, err := Parse(raw)
	if err != nil || u == nil {
		return "", err
	}
	return u.Path, nil
}

// DecodedPath returns u's path after percent-decoding.
func DecodedPath(u *neturl.URL) string {
	if u == nil {
		return ""
	}
	path, err := neturl.PathUnescape(u.EscapedPath())
	if err != nil {
		return u.Path
	}
	return path
}

// ToURI parses location as a URI. If encode is true, blank characters are encoded first.
func ToURI(location string, encode bool) (*neturl.URL, error) {
	location = strings.TrimSpace(location)
	if encode {
		location = EncodeBlank(location)
	}
	return neturl.Parse(location)
}

// IsFileURL reports whether u uses a file-like scheme.
func IsFileURL(u *neturl.URL) bool {
	if u == nil {
		return false
	}
	scheme := strings.ToLower(u.Scheme)
	return scheme == URLProtocolFile || scheme == URLProtocolVFSFile || scheme == URLProtocolVFS
}

// IsJarURL reports whether u uses an archive-like scheme.
func IsJarURL(u *neturl.URL) bool {
	if u == nil {
		return false
	}
	scheme := strings.ToLower(u.Scheme)
	return scheme == URLProtocolJar || scheme == URLProtocolZip || scheme == URLProtocolVFSZip || scheme == URLProtocolWSJar
}

// IsJarFileURL reports whether u is a file URL ending with .jar.
func IsJarFileURL(u *neturl.URL) bool {
	return IsFileURL(u) && strings.HasSuffix(strings.ToLower(u.Path), ".jar")
}

// Open opens a URL resource. It supports http, https, file URLs, and plain file paths.
func Open(raw string) (io.ReadCloser, error) {
	return OpenWithOptions(raw)
}

// OpenWithOptions opens a URL resource with per-call options.
func OpenWithOptions(raw string, opts ...ResourceOption) (io.ReadCloser, error) {
	u, err := neturl.Parse(raw)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "http" || u.Scheme == "https" {
		resp, err := doResourceRequest(raw, http.MethodGet, applyResourceOptions(opts))
		if err != nil {
			return nil, err
		}
		return resp.Body, nil
	}
	path := raw
	if IsFileURL(u) {
		path = u.Path
	}
	// #nosec G304 -- URL utility intentionally opens the caller-provided file URL or path.
	return os.Open(path)
}

// ContentLength returns the resource content length. Unknown lengths return -1.
func ContentLength(raw string) (int64, error) {
	return ContentLengthWithOptions(raw)
}

// ContentLengthWithOptions returns the resource content length with per-call options.
func ContentLengthWithOptions(raw string, opts ...ResourceOption) (int64, error) {
	if raw == "" {
		return -1, nil
	}
	u, err := neturl.Parse(raw)
	if err != nil {
		return -1, err
	}
	if u.Scheme == "http" || u.Scheme == "https" {
		resp, err := doResourceRequest(raw, http.MethodHead, applyResourceOptions(opts))
		if err != nil {
			return -1, err
		}
		defer func() { _ = resp.Body.Close() }()
		return resp.ContentLength, nil
	}
	path := raw
	if IsFileURL(u) {
		path = u.Path
	}
	info, err := os.Stat(path)
	if err != nil {
		return -1, err
	}
	return info.Size(), nil
}

// Size returns the resource size.
func Size(raw string) (int64, error) { return ContentLength(raw) }

// SizeWithOptions returns the resource size with per-call options.
func SizeWithOptions(raw string, opts ...ResourceOption) (int64, error) {
	return ContentLengthWithOptions(raw, opts...)
}

func doResourceRequest(raw, method string, cfg resourceConfig) (*http.Response, error) {
	ctx := cfg.ctx
	if cfg.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.timeout)
		defer cancel()
	}
	req, err := http.NewRequestWithContext(ctx, method, raw, nil)
	if err != nil {
		return nil, err
	}
	for key, values := range cfg.headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	resp, err := cfg.client.Do(req)
	if err != nil {
		return nil, err
	}
	if cfg.checkStatus && (resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices) {
		defer func() { _ = resp.Body.Close() }()
		return nil, fmt.Errorf("unexpected http status %d for %s", resp.StatusCode, raw)
	}
	return resp, nil
}

// Normalize normalizes a URL string by adding a default scheme and cleaning slashes.
func Normalize(raw string, encodePath, replaceSlash bool) string {
	return NormalizeWithOptions(raw, encodePath, replaceSlash)
}

// NormalizeWithOptions normalizes a URL string with per-call options.
func NormalizeWithOptions(raw string, encodePath, replaceSlash bool, opts ...NormalizeOption) string {
	if strings.TrimSpace(raw) == "" {
		return raw
	}
	cfg := applyNormalizeOptions(opts)
	sep := strings.Index(raw, "://")
	protocol := cfg.defaultScheme + "://"
	body := raw
	if sep > 0 {
		protocol = raw[:sep+3]
		body = raw[sep+3:]
	}
	params := ""
	if idx := strings.Index(body, "?"); idx >= 0 {
		params = body[idx:]
		body = body[:idx]
	}
	body = strings.TrimLeft(body, "\\/")
	body = strings.ReplaceAll(body, "\\", "/")
	if replaceSlash {
		for strings.Contains(body, "//") {
			body = strings.ReplaceAll(body, "//", "/")
		}
	}
	domain, path := body, ""
	if idx := strings.Index(body, "/"); idx > 0 {
		domain = body[:idx]
		path = body[idx:]
	}
	if encodePath && path != "" {
		path = encodePathKeepSlash(path)
	}
	return protocol + domain + path + params
}

// BuildQuery converts a map to a URL query string.
func BuildQuery(paramMap map[string]any) string { return EncodeQueryMap(paramMap) }

// EncodeQueryMap converts a map to a URL query string.
func EncodeQueryMap(m map[string]any) string {
	values := neturl.Values{}
	for k, v := range m {
		if k == "" {
			continue
		}
		values.Set(k, valueToString(v))
	}
	return values.Encode()
}

// EncodeParams encodes the query part of rawURL and leaves URLs without query unchanged.
func EncodeParams(rawURL string) string {
	idx := strings.Index(rawURL, "?")
	if idx < 0 {
		return rawURL
	}
	pre := rawURL[:idx]
	q := rawURL[idx+1:]
	values, err := neturl.ParseQuery(q)
	if err != nil {
		return rawURL
	}
	return pre + "?" + values.Encode()
}

// DecodeQueryFirst parses a query string into a single-value map.
func DecodeQueryFirst(paramsStr string) map[string]string {
	out := map[string]string{}
	values, err := neturl.ParseQuery(paramsStr)
	if err != nil {
		return out
	}
	for k, vs := range values {
		if len(vs) > 0 {
			out[k] = vs[0]
		}
	}
	return out
}

// DecodeQuery parses a query string into a multi-value map.
func DecodeQuery(paramsStr string) map[string][]string {
	values, err := neturl.ParseQuery(paramsStr)
	if err != nil {
		return map[string][]string{}
	}
	out := map[string][]string{}
	for k, v := range values {
		out[k] = v
	}
	return out
}

// AppendQuery appends form values to rawURL.
func AppendQuery(rawURL string, form map[string]any) string {
	encoded := EncodeQueryMap(form)
	if encoded == "" {
		return rawURL
	}
	if strings.Contains(rawURL, "?") {
		if strings.HasSuffix(rawURL, "&") || strings.HasSuffix(rawURL, "?") {
			return rawURL + encoded
		}
		return rawURL + "&" + encoded
	}
	return rawURL + "?" + encoded
}

// IsHTTP reports whether raw uses the http scheme prefix.
func IsHTTP(raw string) bool { return strings.HasPrefix(strings.ToLower(raw), "http:") }

// IsHTTPS reports whether raw uses the https scheme prefix.
func IsHTTPS(raw string) bool { return strings.HasPrefix(strings.ToLower(raw), "https:") }

// IsHTTPURL reports whether raw is an absolute http URL with a host.
func IsHTTPURL(raw string) bool { return isAbsoluteURLWithScheme(raw, "http") }

// IsHTTPSURL reports whether raw is an absolute https URL with a host.
func IsHTTPSURL(raw string) bool { return isAbsoluteURLWithScheme(raw, "https") }

// IsWebURL reports whether raw is an absolute http or https URL with a host.
func IsWebURL(raw string) bool { return IsHTTPURL(raw) || IsHTTPSURL(raw) }

// IsAbsoluteURL reports whether raw is an absolute URL with scheme and host.
func IsAbsoluteURL(raw string) bool {
	if raw == "" || strings.TrimSpace(raw) != raw {
		return false
	}
	u, err := neturl.Parse(raw)
	return err == nil && u.IsAbs() && u.Host != ""
}

// DataURIBase64 builds a base64 Data URI string.
func DataURIBase64(mimeType, data string) string { return DataURI(mimeType, "", "base64", data) }

// DataURI builds a Data URI string.
func DataURI(mimeType, charset, encoding, data string) string {
	var b strings.Builder
	b.WriteString("data:")
	if strings.TrimSpace(mimeType) != "" {
		b.WriteString(mimeType)
	}
	if strings.TrimSpace(charset) != "" {
		b.WriteString(";charset=")
		b.WriteString(charset)
	}
	if strings.TrimSpace(encoding) != "" {
		b.WriteByte(';')
		b.WriteString(encoding)
	}
	b.WriteByte(',')
	b.WriteString(data)
	return b.String()
}

func isAbsoluteURLWithScheme(raw, scheme string) bool {
	if raw == "" || strings.TrimSpace(raw) != raw {
		return false
	}
	u, err := neturl.Parse(raw)
	return err == nil && strings.EqualFold(u.Scheme, scheme) && u.Host != ""
}

func encodePathKeepSlash(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		parts[i] = neturl.PathEscape(part)
	}
	return strings.Join(parts, "/")
}

func valueToString(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprint(v)
}
