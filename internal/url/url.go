package url

import (
	"context"
	"fmt"
	"io"
	"net"
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

	// DefaultMaxBytes is the default response body limit used by OpenSafe.
	DefaultMaxBytes int64 = 64 << 20
)

type resourceConfig struct {
	ctx            context.Context
	client         *http.Client
	headers        http.Header
	timeout        time.Duration
	checkStatus    bool
	openFile       func(string) (io.ReadCloser, error)
	stat           func(string) (os.FileInfo, error)
	requestFactory func(context.Context, string, string) (*http.Request, error)
	lookupIP       func(context.Context, string) ([]net.IP, error)
	maxBytes       int64
	allowedSchemes []string
	allowedHosts   []string
	rejectPrivate  bool
	allowLocal     bool
	checkRedirect  bool
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

// WithOpenFile sets the file opener used by local file resource helpers.
func WithOpenFile(openFile func(string) (io.ReadCloser, error)) ResourceOption {
	return func(c *resourceConfig) { c.openFile = openFile }
}

// WithStat sets the stat provider used by local file resource helpers.
func WithStat(stat func(string) (os.FileInfo, error)) ResourceOption {
	return func(c *resourceConfig) { c.stat = stat }
}

// WithRequestFactory sets the HTTP request factory used by resource helpers.
func WithRequestFactory(factory func(context.Context, string, string) (*http.Request, error)) ResourceOption {
	return func(c *resourceConfig) { c.requestFactory = factory }
}

// WithLookupIP sets the host resolver used by SSRF-oriented URL validation and safe dialing.
func WithLookupIP(lookupIP func(context.Context, string) ([]net.IP, error)) ResourceOption {
	return func(c *resourceConfig) { c.lookupIP = lookupIP }
}

// WithMaxBytes limits how many response body bytes OpenWithOptions may read.
// Non-positive values mean unlimited unless another option sets a positive limit.
func WithMaxBytes(n int64) ResourceOption {
	return func(c *resourceConfig) { c.maxBytes = n }
}

// WithAllowedSchemes restricts resource helpers to the provided URL schemes.
func WithAllowedSchemes(schemes ...string) ResourceOption {
	return func(c *resourceConfig) { c.allowedSchemes = append([]string(nil), schemes...) }
}

// WithAllowedHosts restricts HTTP(S) resource helpers to the provided host names.
func WithAllowedHosts(hosts ...string) ResourceOption {
	return func(c *resourceConfig) { c.allowedHosts = append([]string(nil), hosts...) }
}

// WithRejectPrivateHosts rejects localhost, loopback, private, and link-local HTTP(S) hosts unless explicitly allowed.
func WithRejectPrivateHosts(reject bool) ResourceOption {
	return func(c *resourceConfig) { c.rejectPrivate = reject }
}

// WithAllowLocalFiles controls whether file URLs and plain filesystem paths are allowed.
func WithAllowLocalFiles(allow bool) ResourceOption {
	return func(c *resourceConfig) { c.allowLocal = allow }
}

func applyResourceOptions(opts []ResourceOption) resourceConfig {
	cfg := resourceConfig{ctx: context.Background(), client: http.DefaultClient, openFile: defaultOpenFile, stat: os.Stat, requestFactory: defaultRequestFactory, lookupIP: defaultLookupIP, allowLocal: true}
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
	if cfg.openFile == nil {
		cfg.openFile = defaultOpenFile
	}
	if cfg.stat == nil {
		cfg.stat = os.Stat
	}
	if cfg.requestFactory == nil {
		cfg.requestFactory = defaultRequestFactory
	}
	if cfg.lookupIP == nil {
		cfg.lookupIP = defaultLookupIP
	}
	return cfg
}

func safeResourceOptions(opts []ResourceOption) []ResourceOption {
	safe := make([]ResourceOption, 0, 6+len(opts))
	safe = append(safe,
		WithAllowedSchemes("http", "https"),
		WithRejectPrivateHosts(true),
		WithAllowLocalFiles(false),
		WithCheckStatus(true),
		WithTimeout(10*time.Second),
		WithMaxBytes(DefaultMaxBytes),
		func(c *resourceConfig) { c.checkRedirect = true },
	)
	return append(safe, opts...)
}

func defaultOpenFile(path string) (io.ReadCloser, error) {
	// #nosec G304 -- URL resource helpers intentionally read caller-provided file URLs or paths.
	return os.Open(path)
}

func defaultRequestFactory(ctx context.Context, method, raw string) (*http.Request, error) {
	return http.NewRequestWithContext(ctx, method, raw, nil)
}

func defaultLookupIP(ctx context.Context, host string) ([]net.IP, error) {
	return net.DefaultResolver.LookupIP(ctx, "ip", host)
}

type normalizeConfig struct {
	defaultScheme string
	encodePath    bool
	replaceSlash  bool
	setEncodePath bool
	setReplace    bool
}

// NormalizeOption customizes URL normalization.
type NormalizeOption func(*normalizeConfig)

// WithDefaultScheme sets the scheme used when NormalizeWithOptions receives a URL without scheme.
func WithDefaultScheme(scheme string) NormalizeOption {
	return func(c *normalizeConfig) { c.defaultScheme = scheme }
}

// WithEncodePath controls whether NormalizeUsingOptions escapes the normalized path.
func WithEncodePath(encode bool) NormalizeOption {
	return func(c *normalizeConfig) {
		c.encodePath = encode
		c.setEncodePath = true
	}
}

// WithReplaceSlash controls whether NormalizeUsingOptions collapses repeated slashes in the path.
func WithReplaceSlash(replace bool) NormalizeOption {
	return func(c *normalizeConfig) {
		c.replaceSlash = replace
		c.setReplace = true
	}
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
func Encode(s string) string { return EncodeWithOptions(s) }

// EncodeWithOptions escapes a string for URL query components with custom providers.
func EncodeWithOptions(s string, opts ...EncodeOption) string {
	return applyEncodeOptions(opts).queryEscape(s)
}

// URLEncode escapes a string for URL query components.
func URLEncode(s string) string { return Encode(s) }

// URLEncodeWithOptions escapes a string for URL query components with custom providers.
func URLEncodeWithOptions(s string, opts ...EncodeOption) string {
	return EncodeWithOptions(s, opts...)
}

// Decode unescapes a URL query component and converts plus signs to spaces.
func Decode(s string) (string, error) { return DecodePlus(s, true) }

// URLDecode unescapes a URL query component and converts plus signs to spaces.
func URLDecode(s string) (string, error) { return Decode(s) }

// DecodePlus unescapes percent-encoded text and controls whether plus signs become spaces.
func DecodePlus(s string, plusToSpace bool) (string, error) {
	return DecodeWithOptions(s, WithPlusAsSpace(plusToSpace))
}

type decodeConfig struct {
	plusToSpace bool
}

// DecodeOption customizes DecodeWithOptions.
type DecodeOption func(*decodeConfig)

// WithPlusAsSpace controls whether plus signs are decoded as spaces.
func WithPlusAsSpace(plusToSpace bool) DecodeOption {
	return func(c *decodeConfig) { c.plusToSpace = plusToSpace }
}

func applyDecodeOptions(opts []DecodeOption) decodeConfig {
	cfg := decodeConfig{plusToSpace: true}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return cfg
}

// DecodeWithOptions unescapes percent-encoded text with custom decoding behavior.
func DecodeWithOptions(s string, opts ...DecodeOption) (string, error) {
	cfg := applyDecodeOptions(opts)
	if cfg.plusToSpace {
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
//
// Security: Open is for trusted resource locations only because it may access
// local files and private network addresses. Use OpenSafe for resource locations
// that may cross a user, configuration, or network trust boundary.
func Open(raw string) (io.ReadCloser, error) {
	return OpenWithOptions(raw)
}

// OpenSafe opens an HTTP(S) URL with secure defaults for untrusted input.
func OpenSafe(raw string) (io.ReadCloser, error) { return OpenSafeWithOptions(raw) }

// OpenSafeWithOptions opens an HTTP(S) URL with secure defaults for untrusted input.
func OpenSafeWithOptions(raw string, opts ...ResourceOption) (io.ReadCloser, error) {
	return OpenWithOptions(raw, safeResourceOptions(opts)...)
}

// OpenWithOptions opens a URL resource with per-call options.
//
// Security: OpenWithOptions is for trusted resource locations unless options
// explicitly restrict schemes, local files, redirects, and private hosts. Prefer
// OpenSafeWithOptions for untrusted input.
func OpenWithOptions(raw string, opts ...ResourceOption) (io.ReadCloser, error) {
	cfg := applyResourceOptions(opts)
	u, err := neturl.Parse(raw)
	if err != nil {
		return nil, err
	}
	if err := validateResourceURL(u, cfg); err != nil {
		return nil, err
	}
	if u.Scheme == "http" || u.Scheme == "https" {
		resp, err := doResourceRequest(raw, http.MethodGet, cfg)
		if err != nil {
			return nil, err
		}
		if cfg.maxBytes > 0 && resp.ContentLength > cfg.maxBytes {
			defer func() { _ = resp.Body.Close() }()
			return nil, fmt.Errorf("response body exceeds max bytes: %d", cfg.maxBytes)
		}
		if cfg.maxBytes > 0 {
			return limitReadCloser(resp.Body, cfg.maxBytes), nil
		}
		return resp.Body, nil
	}
	path := raw
	if IsFileURL(u) {
		path = u.Path
	}
	// #nosec G304 -- URL utility intentionally opens the caller-provided file URL or path.
	return cfg.openFile(path)
}

// ContentLength returns the resource content length. Unknown lengths return -1.
//
// Security: ContentLength is for trusted resource locations only because it may
// access local files and private network addresses. Use ContentLengthSafe for
// untrusted input.
func ContentLength(raw string) (int64, error) {
	return ContentLengthWithOptions(raw)
}

// ContentLengthSafe returns an HTTP(S) resource content length with secure defaults for untrusted input.
func ContentLengthSafe(raw string) (int64, error) { return ContentLengthSafeWithOptions(raw) }

// ContentLengthSafeWithOptions returns an HTTP(S) resource content length with secure defaults for untrusted input.
func ContentLengthSafeWithOptions(raw string, opts ...ResourceOption) (int64, error) {
	return ContentLengthWithOptions(raw, safeResourceOptions(opts)...)
}

// ContentLengthWithOptions returns the resource content length with per-call options.
//
// Security: ContentLengthWithOptions is for trusted resource locations unless
// options explicitly restrict schemes, local files, redirects, and private hosts.
// Prefer ContentLengthSafeWithOptions for untrusted input.
func ContentLengthWithOptions(raw string, opts ...ResourceOption) (int64, error) {
	cfg := applyResourceOptions(opts)
	if raw == "" {
		return -1, nil
	}
	u, err := neturl.Parse(raw)
	if err != nil {
		return -1, err
	}
	if err := validateResourceURL(u, cfg); err != nil {
		return -1, err
	}
	if u.Scheme == "http" || u.Scheme == "https" {
		resp, err := doResourceRequest(raw, http.MethodHead, cfg)
		if err != nil {
			return -1, err
		}
		defer func() { _ = resp.Body.Close() }()
		if cfg.maxBytes > 0 && resp.ContentLength > cfg.maxBytes {
			return -1, fmt.Errorf("response body exceeds max bytes: %d", cfg.maxBytes)
		}
		return resp.ContentLength, nil
	}
	path := raw
	if IsFileURL(u) {
		path = u.Path
	}
	info, err := cfg.stat(path)
	if err != nil {
		return -1, err
	}
	return info.Size(), nil
}

// Size returns the resource size.
//
// Security: Size follows ContentLength and is for trusted resource locations
// only. Use ContentLengthSafe for untrusted input.
func Size(raw string) (int64, error) { return ContentLength(raw) }

// SizeWithOptions returns the resource size with per-call options.
//
// Security: SizeWithOptions follows ContentLengthWithOptions and is for trusted
// resource locations unless safe options are enabled.
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
	req, err := cfg.requestFactory(ctx, method, raw)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, fmt.Errorf("request factory returned nil for %s %s", method, raw)
	}
	if req.Context() != ctx {
		req = req.WithContext(ctx)
	}
	if err := validateResourceURL(req.URL, cfg); err != nil {
		return nil, err
	}
	for key, values := range cfg.headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	client := cfg.client
	if cfg.checkRedirect {
		clone := *client
		clone.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return validateResourceURL(req.URL, cfg)
		}
		client = &clone
	}
	if cfg.rejectPrivate {
		client = clientWithSafeTransport(client, cfg)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if cfg.checkStatus && (resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices) {
		defer func() { _ = resp.Body.Close() }()
		return nil, fmt.Errorf("unexpected http status %d for %s", resp.StatusCode, raw)
	}
	return resp, nil
}

func validateResourceURL(u *neturl.URL, cfg resourceConfig) error {
	if u == nil {
		return fmt.Errorf("url is nil")
	}
	scheme := strings.ToLower(strings.TrimSpace(u.Scheme))
	if len(cfg.allowedSchemes) > 0 && !containsFold(cfg.allowedSchemes, scheme) {
		return fmt.Errorf("url scheme %q is not allowed", scheme)
	}
	if scheme == "http" || scheme == "https" {
		host := strings.ToLower(u.Hostname())
		if host == "" {
			return fmt.Errorf("http url host is blank")
		}
		if len(cfg.allowedHosts) > 0 && !containsFold(cfg.allowedHosts, host) {
			return fmt.Errorf("url host %q is not allowed", host)
		}
		if cfg.rejectPrivate {
			private, err := isPrivateHost(cfg.ctx, cfg.lookupIP, host)
			if err != nil {
				return err
			}
			if private {
				return fmt.Errorf("url host %q resolves to a private address", host)
			}
		}
		return nil
	}
	if !cfg.allowLocal && (scheme == "" || IsFileURL(u)) {
		return fmt.Errorf("local file resources are not allowed")
	}
	return nil
}

func containsFold(values []string, target string) bool {
	target = strings.TrimSpace(target)
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value), target) {
			return true
		}
	}
	return false
}

func isPrivateHost(ctx context.Context, lookupIP func(context.Context, string) ([]net.IP, error), host string) (bool, error) {
	if strings.EqualFold(host, "localhost") {
		return true, nil
	}
	if ip := net.ParseIP(host); ip != nil {
		return isPrivateIP(ip), nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if lookupIP == nil {
		lookupIP = defaultLookupIP
	}
	ips, err := lookupIP(ctx, host)
	if err != nil {
		return false, fmt.Errorf("resolve url host %q: %w", host, err)
	}
	for _, ip := range ips {
		if isPrivateIP(ip) {
			return true, nil
		}
	}
	return false, nil
}

func isPrivateIP(ip net.IP) bool {
	return ip == nil || !ip.IsGlobalUnicast() || ip.IsPrivate()
}

func clientWithSafeTransport(client *http.Client, cfg resourceConfig) *http.Client {
	clone := *client
	base := client.Transport
	if base == nil {
		base = http.DefaultTransport
	}
	if transport, ok := base.(*http.Transport); ok {
		transportClone := transport.Clone()
		transportClone.DialContext = safeDialContext(cfg)
		base = transportClone
	}
	clone.Transport = safeResourceTransport{base: base, cfg: cfg}
	return &clone
}

func safeDialContext(cfg resourceConfig) func(context.Context, string, string) (net.Conn, error) {
	dialer := &net.Dialer{}
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}
		host = strings.ToLower(strings.TrimSpace(host))
		if host == "" {
			return nil, fmt.Errorf("dial host is blank")
		}
		ips, err := publicHostIPs(ctx, cfg, host)
		if err != nil {
			return nil, err
		}
		return dialer.DialContext(ctx, network, net.JoinHostPort(ips[0].String(), port))
	}
}

func publicHostIPs(ctx context.Context, cfg resourceConfig, host string) ([]net.IP, error) {
	if strings.EqualFold(host, "localhost") {
		return nil, fmt.Errorf("url host %q resolves to a private address", host)
	}
	if ip := net.ParseIP(host); ip != nil {
		if isPrivateIP(ip) {
			return nil, fmt.Errorf("url host %q resolves to a private address", host)
		}
		return []net.IP{ip}, nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	lookupIP := cfg.lookupIP
	if lookupIP == nil {
		lookupIP = defaultLookupIP
	}
	ips, err := lookupIP(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("resolve url host %q: %w", host, err)
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("resolve url host %q: no addresses", host)
	}
	public := make([]net.IP, 0, len(ips))
	for _, ip := range ips {
		if isPrivateIP(ip) {
			return nil, fmt.Errorf("url host %q resolves to a private address", host)
		}
		public = append(public, ip)
	}
	return public, nil
}

type safeResourceTransport struct {
	base http.RoundTripper
	cfg  resourceConfig
}

func (t safeResourceTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req == nil || req.URL == nil {
		return nil, fmt.Errorf("request url is nil")
	}
	if err := validateResourceURL(req.URL, t.cfg); err != nil {
		return nil, err
	}
	if req.URL.Scheme == "http" || req.URL.Scheme == "https" {
		host := strings.ToLower(req.URL.Hostname())
		private, err := isPrivateHost(req.Context(), t.cfg.lookupIP, host)
		if err != nil {
			return nil, err
		}
		if private {
			return nil, fmt.Errorf("url host %q resolves to a private address", host)
		}
	}
	return t.base.RoundTrip(req)
}

type limitedReadCloser struct {
	r         io.Reader
	c         io.Closer
	remaining int64
}

func limitReadCloser(rc io.ReadCloser, maxBytes int64) io.ReadCloser {
	return &limitedReadCloser{r: rc, c: rc, remaining: maxBytes}
}

func (r *limitedReadCloser) Read(p []byte) (int, error) {
	if r.remaining <= 0 {
		var b [1]byte
		n, err := r.r.Read(b[:])
		if n > 0 {
			return 0, fmt.Errorf("response body exceeds max bytes")
		}
		return 0, err
	}
	if int64(len(p)) > r.remaining {
		p = p[:r.remaining]
	}
	n, err := r.r.Read(p)
	r.remaining -= int64(n)
	return n, err
}

func (r *limitedReadCloser) Close() error { return r.c.Close() }

// Normalize normalizes a URL string by adding a default scheme and cleaning slashes.
func Normalize(raw string, encodePath, replaceSlash bool) string {
	return NormalizeWithOptions(raw, encodePath, replaceSlash)
}

// NormalizeWithOptions normalizes a URL string with per-call options.
func NormalizeWithOptions(raw string, encodePath, replaceSlash bool, opts ...NormalizeOption) string {
	return normalize(raw, encodePath, replaceSlash, opts...)
}

// NormalizeUsingOptions normalizes a URL string using only functional options for optional behavior.
func NormalizeUsingOptions(raw string, opts ...NormalizeOption) string {
	cfg := applyNormalizeOptions(opts)
	return normalize(raw, cfg.encodePath, cfg.replaceSlash, opts...)
}

func normalize(raw string, encodePath, replaceSlash bool, opts ...NormalizeOption) string {
	if strings.TrimSpace(raw) == "" {
		return raw
	}
	cfg := applyNormalizeOptions(opts)
	if cfg.setEncodePath {
		encodePath = cfg.encodePath
	}
	if cfg.setReplace {
		replaceSlash = cfg.replaceSlash
	}
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
