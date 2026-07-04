package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"maps"
	"net"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"

	knifer "github.com/imajinyun/knifer-go"
	"github.com/imajinyun/knifer-go/internal/httpboundary"
)

var (
	defaultTransportMu       sync.Mutex
	defaultTransport         *http.Transport
	defaultTransportProvider = cloneDefaultTransport
)

func cloneDefaultTransport() *http.Transport {
	if transport, ok := http.DefaultTransport.(*http.Transport); ok && transport != nil {
		return transport.Clone()
	}
	return &http.Transport{}
}

func getDefaultTransport() *http.Transport {
	defaultTransportMu.Lock()
	defer defaultTransportMu.Unlock()
	if defaultTransport == nil {
		defaultTransport = defaultTransportProvider()
		if defaultTransport == nil {
			defaultTransport = cloneDefaultTransport()
		}
	}
	return defaultTransport
}

// ConfigureDefaultTransportProvider sets the provider used to initialize the shared default transport.
// Passing nil restores the provider that clones http.DefaultTransport. Existing idle connections are closed.
func ConfigureDefaultTransportProvider(provider func() *http.Transport) {
	defaultTransportMu.Lock()
	defer defaultTransportMu.Unlock()
	if defaultTransport != nil {
		defaultTransport.CloseIdleConnections()
	}
	defaultTransport = nil
	if provider == nil {
		defaultTransportProvider = cloneDefaultTransport
		return
	}
	defaultTransportProvider = provider
}

// ResetDefaultTransport clears the cached shared default transport and restores the standard provider.
func ResetDefaultTransport() { ConfigureDefaultTransportProvider(nil) }

// HTTPRequest is a chainable HTTP request builder, aligned with the utility toolkit-http HttpRequest.
type HTTPRequest struct {
	mu           sync.Mutex
	used         bool
	method       Method
	rawURL       string
	queryParams  url.Values
	headers      http.Header
	cookies      []*http.Cookie
	cookieJar    http.CookieJar
	body         []byte
	bodyReader   io.Reader
	form         map[string]any
	multipart    bool
	multipartFs  []*formFile
	contentType  string
	charset      string
	timeout      time.Duration
	followRedir  *bool
	maxRedirects int
	tlsConfig    *tls.Config
	userAgent    string
	transport    http.RoundTripper
	transportFn  func() http.RoundTripper
	basicUser    string
	basicPass    string
	hasBasic     bool
	httpClient   *http.Client
	newRequest   NewRequestFunc
	multipartNew MultipartWriterFactory
	decodeConfig responseDecodeConfig
	urlPolicy    *URLPolicy
}

// URLPolicy controls SSRF-oriented request validation for untrusted URLs.
type URLPolicy struct {
	AllowedSchemes []string
	AllowedHosts   []string
	RejectPrivate  bool
	LookupIP       func(context.Context, string) ([]net.IP, error)
}

type formFile struct {
	field    string
	fileName string
	data     []byte
	reader   io.Reader
}

func (f *formFile) clone() *formFile {
	if f == nil {
		return nil
	}
	return &formFile{
		field:    f.field,
		fileName: f.fileName,
		data:     slices.Clone(f.data),
		reader:   f.reader,
	}
}

// RequestOption customizes one HTTP request at construction time.
type RequestOption func(*HTTPRequest)

// Client is an explicit HTTP request factory with a captured configuration snapshot.
// Use it when code should not read package-level global defaults on every request.
type Client struct {
	cfg  GlobalConfig
	opts []RequestOption
}

// ClientOption customizes a Client.
type ClientOption func(*Client)

// WithClientGlobalConfig sets the configuration snapshot used by a Client.
func WithClientGlobalConfig(cfg GlobalConfig) ClientOption {
	return func(c *Client) { c.cfg = cfg }
}

// WithClientRequestOptions sets request options applied to every request created by a Client.
func WithClientRequestOptions(opts ...RequestOption) ClientOption {
	return func(c *Client) { c.opts = slices.Clone(opts) }
}

// NewClient creates a request factory using the current global configuration snapshot.
func NewClient(opts ...ClientOption) *Client {
	c := &Client{cfg: SnapshotGlobalConfig()}
	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}
	return c
}

// NewIsolatedClient creates a request factory without reading package-level global defaults.
func NewIsolatedClient(opts ...ClientOption) *Client {
	c := &Client{cfg: isolatedGlobalConfig()}
	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}
	return c
}

// NewClientWithConfig creates a request factory from an explicit configuration snapshot.
func NewClientWithConfig(cfg GlobalConfig, opts ...RequestOption) *Client {
	return &Client{cfg: cfg, opts: slices.Clone(opts)}
}

// NewRequest creates a request from the Client's captured configuration.
func (c *Client) NewRequest(method Method, rawURL string, opts ...RequestOption) *HTTPRequest {
	if c == nil {
		return NewIsolatedRequest(method, rawURL, opts...)
	}
	all := append(slices.Clone(c.opts), opts...)
	return NewRequestWithConfig(method, rawURL, c.cfg, all...)
}

// NewSafeRequest creates a safe request from the Client's captured configuration.
func (c *Client) NewSafeRequest(method Method, rawURL string, opts ...RequestOption) *HTTPRequest {
	if c == nil {
		return NewSafeRequest(method, rawURL, opts...)
	}
	safe := make([]RequestOption, 0, 4+len(c.opts)+len(opts))
	safe = append(safe,
		WithURLPolicy(URLPolicy{AllowedSchemes: []string{"http", "https"}, RejectPrivate: true}),
		WithTimeout(10*time.Second),
		WithMaxRedirects(10),
		WithMaxResponseBytes(defaultGlobalMaxResponseBytes),
	)
	safe = append(safe, c.opts...)
	safe = append(safe, opts...)
	return NewRequestWithConfig(method, rawURL, c.cfg, safe...)
}

// Get creates a GET request from the Client's captured configuration.
func (c *Client) Get(rawURL string, opts ...RequestOption) *HTTPRequest {
	return c.NewRequest(MethodGet, rawURL, opts...)
}

// GetSafe creates a safe GET request from the Client's captured configuration.
func (c *Client) GetSafe(rawURL string, opts ...RequestOption) *HTTPRequest {
	return c.NewSafeRequest(MethodGet, rawURL, opts...)
}

// Post creates a POST request from the Client's captured configuration.
func (c *Client) Post(rawURL string, opts ...RequestOption) *HTTPRequest {
	return c.NewRequest(MethodPost, rawURL, opts...)
}

// PostSafe creates a safe POST request from the Client's captured configuration.
func (c *Client) PostSafe(rawURL string, opts ...RequestOption) *HTTPRequest {
	return c.NewSafeRequest(MethodPost, rawURL, opts...)
}

// NewRequestFunc creates an outgoing HTTP request.
type NewRequestFunc func(method, url string, body io.Reader) (*http.Request, error)

// MultipartWriterFactory creates a multipart writer for request bodies.
type MultipartWriterFactory func(io.Writer) MultipartWriter

// MultipartWriter is the subset of multipart.Writer used by request construction.
type MultipartWriter interface {
	WriteField(string, string) error
	CreateFormFile(string, string) (io.Writer, error)
	Close() error
	FormDataContentType() string
}

// WithGlobalConfig initializes request defaults from a captured global configuration snapshot.
func WithGlobalConfig(cfg GlobalConfig) RequestOption {
	return func(r *HTTPRequest) {
		r.headers = cloneHeader(cfg.Headers)
		r.cookieJar = cfg.CookieJar
		r.timeout = cfg.Timeout
		follow := cfg.FollowRedirects
		r.followRedir = &follow
		r.maxRedirects = cfg.MaxRedirects
		r.userAgent = cfg.DefaultUserAgent
		r.decodeConfig = responseDecodeConfigFromGlobal(cfg)
	}
}

// NewRequest creates a request with the specified method and URL.
//
// Security: NewRequest is for trusted URLs. This package does not apply
// SSRF-oriented host validation by default; use NewSafeRequest, GetSafe, or
// PostSafe when the URL may come from users, config, or another untrusted trust boundary.
func NewRequest(method Method, rawURL string, opts ...RequestOption) *HTTPRequest {
	return NewRequestWithConfig(method, rawURL, SnapshotGlobalConfig(), opts...)
}

// NewIsolatedRequest creates a request without reading package-level global defaults.
func NewIsolatedRequest(method Method, rawURL string, opts ...RequestOption) *HTTPRequest {
	return NewRequestWithConfig(method, rawURL, isolatedGlobalConfig(), opts...)
}

func isolatedGlobalConfig() GlobalConfig {
	return GlobalConfig{
		Timeout:          defaultGlobalTimeout,
		FollowRedirects:  true,
		MaxRedirects:     defaultGlobalMaxRedirects,
		MaxResponseBytes: defaultGlobalMaxResponseBytes,
	}
}

// NewRequestWithConfig creates a request from an explicit global configuration snapshot.
//
// Security: NewRequestWithConfig is for trusted URLs. This package does not
// apply SSRF-oriented host validation by default; use NewSafeRequest when the
// URL may come from users, config, or another untrusted trust boundary.
func NewRequestWithConfig(method Method, rawURL string, cfg GlobalConfig, opts ...RequestOption) *HTTPRequest {
	follow := cfg.FollowRedirects
	r := &HTTPRequest{
		method:       method,
		rawURL:       rawURL,
		queryParams:  url.Values{},
		headers:      cloneHeader(cfg.Headers),
		cookieJar:    cfg.CookieJar,
		charset:      "UTF-8",
		timeout:      cfg.Timeout,
		followRedir:  &follow,
		maxRedirects: cfg.MaxRedirects,
		userAgent:    cfg.DefaultUserAgent,
		newRequest:   http.NewRequest,
		multipartNew: newMultipartWriter,
		decodeConfig: responseDecodeConfigFromGlobal(cfg),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(r)
		}
	}
	return r
}

// Get creates a GET request.
//
// Security: Get is for trusted URLs. Use GetSafe when the URL is untrusted.
func Get(rawURL string, opts ...RequestOption) *HTTPRequest {
	return NewRequest(MethodGet, rawURL, opts...)
}

// GetSafe creates a GET request with SSRF-oriented safety checks enabled.
func GetSafe(rawURL string, opts ...RequestOption) *HTTPRequest {
	return NewSafeRequest(MethodGet, rawURL, opts...)
}

// Post creates a POST request.
//
// Security: Post is for trusted URLs. Use PostSafe when the URL is untrusted.
func Post(rawURL string, opts ...RequestOption) *HTTPRequest {
	return NewRequest(MethodPost, rawURL, opts...)
}

// PostSafe creates a POST request with SSRF-oriented safety checks enabled.
func PostSafe(rawURL string, opts ...RequestOption) *HTTPRequest {
	return NewSafeRequest(MethodPost, rawURL, opts...)
}

// NewSafeRequest creates a request with SSRF-oriented safety checks enabled.
func NewSafeRequest(method Method, rawURL string, opts ...RequestOption) *HTTPRequest {
	safe := make([]RequestOption, 0, 4+len(opts))
	safe = append(safe,
		WithURLPolicy(URLPolicy{AllowedSchemes: []string{"http", "https"}, RejectPrivate: true}),
		WithTimeout(10*time.Second),
		WithMaxRedirects(10),
		WithMaxResponseBytes(defaultGlobalMaxResponseBytes),
	)
	safe = append(safe, opts...)
	return NewRequest(method, rawURL, safe...)
}

// Put creates a PUT request.
//
// Security: Put is for trusted URLs. Use PutSafe when the URL is untrusted.
func Put(rawURL string, opts ...RequestOption) *HTTPRequest {
	return NewRequest(MethodPut, rawURL, opts...)
}

// PutSafe creates a PUT request with SSRF-oriented safety checks enabled.
func PutSafe(rawURL string, opts ...RequestOption) *HTTPRequest {
	return NewSafeRequest(MethodPut, rawURL, opts...)
}

// Delete creates a DELETE request.
//
// Security: Delete is for trusted URLs. Use DeleteSafe when the URL is untrusted.
func Delete(rawURL string, opts ...RequestOption) *HTTPRequest {
	return NewRequest(MethodDelete, rawURL, opts...)
}

// DeleteSafe creates a DELETE request with SSRF-oriented safety checks enabled.
func DeleteSafe(rawURL string, opts ...RequestOption) *HTTPRequest {
	return NewSafeRequest(MethodDelete, rawURL, opts...)
}

// Patch creates a PATCH request.
//
// Security: Patch is for trusted URLs. Use PatchSafe when the URL is untrusted.
func Patch(rawURL string, opts ...RequestOption) *HTTPRequest {
	return NewRequest(MethodPatch, rawURL, opts...)
}

// PatchSafe creates a PATCH request with SSRF-oriented safety checks enabled.
func PatchSafe(rawURL string, opts ...RequestOption) *HTTPRequest {
	return NewSafeRequest(MethodPatch, rawURL, opts...)
}

// Head creates a HEAD request.
//
// Security: Head is for trusted URLs. Use HeadSafe when the URL is untrusted.
func Head(rawURL string, opts ...RequestOption) *HTTPRequest {
	return NewRequest(MethodHead, rawURL, opts...)
}

// HeadSafe creates a HEAD request with SSRF-oriented safety checks enabled.
func HeadSafe(rawURL string, opts ...RequestOption) *HTTPRequest {
	return NewSafeRequest(MethodHead, rawURL, opts...)
}

// Options creates an OPTIONS request.
//
// Security: Options is for trusted URLs. Use OptionsSafe when the URL is untrusted.
func Options(rawURL string, opts ...RequestOption) *HTTPRequest {
	return NewRequest(MethodOptions, rawURL, opts...)
}

// OptionsSafe creates an OPTIONS request with SSRF-oriented safety checks enabled.
func OptionsSafe(rawURL string, opts ...RequestOption) *HTTPRequest {
	return NewSafeRequest(MethodOptions, rawURL, opts...)
}

// WithTimeout sets a per-request timeout.
func WithTimeout(d time.Duration) RequestOption { return func(r *HTTPRequest) { r.Timeout(d) } }

// WithHeader sets one per-request header.
func WithHeader(name, value string) RequestOption {
	return func(r *HTTPRequest) { r.Header(name, value) }
}

// WithHeaders sets per-request headers in batch.
func WithHeaders(headers map[string]string) RequestOption {
	return func(r *HTTPRequest) { r.Headers(headers) }
}

// WithFollowRedirects sets per-request redirect behavior.
func WithFollowRedirects(b bool) RequestOption { return func(r *HTTPRequest) { r.FollowRedirects(b) } }

// WithMaxRedirects sets the per-request redirect limit.
func WithMaxRedirects(n int) RequestOption { return func(r *HTTPRequest) { r.MaxRedirects(n) } }

// WithTLSConfig sets a per-request TLS config. It is ignored when a custom client or transport is set.
func WithTLSConfig(cfg *tls.Config) RequestOption { return func(r *HTTPRequest) { r.TLSConfig(cfg) } }

// WithTransport sets a per-request RoundTripper.
func WithTransport(t http.RoundTripper) RequestOption {
	return func(r *HTTPRequest) {
		if t != nil {
			r.Transport(t)
		}
	}
}

// WithTransportProvider sets a per-request RoundTripper provider evaluated when the request is built.
func WithTransportProvider(provider func() http.RoundTripper) RequestOption {
	return func(r *HTTPRequest) {
		if provider != nil {
			r.transportFn = provider
		}
	}
}

// WithClient sets a per-request HTTP client.
func WithClient(c *http.Client) RequestOption {
	return func(r *HTTPRequest) {
		if c != nil {
			r.Client(c)
		}
	}
}

// WithCookieJar sets a per-request CookieJar. nil disables cookie management for this request.
func WithCookieJar(jar http.CookieJar) RequestOption {
	return func(r *HTTPRequest) { r.cookieJar = jar }
}

// WithUserAgent sets a per-request User-Agent.
func WithUserAgent(ua string) RequestOption {
	return func(r *HTTPRequest) {
		r.userAgent = ua
		r.headers.Set(string(HeaderUserAgent), ua)
	}
}

// WithContentType sets a per-request Content-Type at construction time.
func WithContentType(ct string) RequestOption { return func(r *HTTPRequest) { r.ContentType(ct) } }

// WithCharset sets a per-request charset at construction time.
func WithCharset(charset string) RequestOption { return func(r *HTTPRequest) { r.Charset(charset) } }

// WithAutoDecodeResponse controls whether response bodies are decoded by Content-Encoding.
func WithAutoDecodeResponse(autoDecode bool) RequestOption {
	return func(r *HTTPRequest) { r.decodeConfig.autoDecode = autoDecode }
}

// WithMaxResponseBytes limits bytes read by response Bytes/Body helpers. Non-positive means unlimited.
func WithMaxResponseBytes(maxBytes int64) RequestOption {
	return func(r *HTTPRequest) { r.decodeConfig.maxBytes = maxBytes }
}

// WithResponseReadAllFunc sets the reader used by response Bytes/Body helpers.
func WithResponseReadAllFunc(readAll func(io.Reader) ([]byte, error)) RequestOption {
	return func(r *HTTPRequest) {
		if readAll != nil {
			r.decodeConfig.readAll = readAll
		}
	}
}

// WithRequestFactory sets the HTTP request factory used at execution time.
func WithRequestFactory(newRequest NewRequestFunc) RequestOption {
	return func(r *HTTPRequest) {
		if newRequest != nil {
			r.newRequest = newRequest
		}
	}
}

// WithMultipartWriterFactory sets the multipart writer factory used when building multipart request bodies.
func WithMultipartWriterFactory(factory MultipartWriterFactory) RequestOption {
	return func(r *HTTPRequest) {
		if factory != nil {
			r.multipartNew = factory
		}
	}
}

// WithURLPolicy sets SSRF-oriented validation for the request URL and redirect targets.
func WithURLPolicy(policy URLPolicy) RequestOption {
	return func(r *HTTPRequest) {
		p := policy
		p.AllowedSchemes = slices.Clone(policy.AllowedSchemes)
		p.AllowedHosts = slices.Clone(policy.AllowedHosts)
		r.urlPolicy = &p
	}
}

// WithAllowedHosts restricts Safe requests to the provided host names. It does
// not bypass RejectPrivate; allowlisted hosts that resolve to private addresses
// are still rejected unless the URLPolicy explicitly disables RejectPrivate.
func WithAllowedHosts(hosts ...string) RequestOption {
	return func(r *HTTPRequest) {
		if r.urlPolicy == nil {
			r.urlPolicy = &URLPolicy{AllowedSchemes: []string{"http", "https"}, RejectPrivate: true}
		}
		r.urlPolicy.AllowedHosts = slices.Clone(hosts)
	}
}

// WithLookupIP sets the host resolver used by SSRF-oriented URL validation.
func WithLookupIP(lookupIP func(context.Context, string) ([]net.IP, error)) RequestOption {
	return func(r *HTTPRequest) {
		if r.urlPolicy == nil {
			r.urlPolicy = &URLPolicy{AllowedSchemes: []string{"http", "https"}, RejectPrivate: true}
		}
		r.urlPolicy.LookupIP = lookupIP
	}
}

// WithContentDecoder registers a per-request response body decoder for encoding.
func WithContentDecoder(encoding string, decoder ContentDecoder) RequestOption {
	return func(r *HTTPRequest) { r.decodeConfig.setDecoder(encoding, decoder) }
}

// Clone returns an independent request builder snapshot.
//
// Replayable data such as headers, query values, cookies, byte bodies, forms,
// and byte-backed multipart files is copied. Reader-backed bodies and
// multipart files share the same reader and therefore remain single-use unless
// callers replace them on the clone with a fresh reader.
func (r *HTTPRequest) Clone() *HTTPRequest {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	clone := &HTTPRequest{
		method:       r.method,
		rawURL:       r.rawURL,
		queryParams:  cloneValues(r.queryParams),
		headers:      cloneHeader(r.headers),
		cookieJar:    r.cookieJar,
		body:         slices.Clone(r.body),
		bodyReader:   r.bodyReader,
		form:         cloneAnyMap(r.form),
		multipart:    r.multipart,
		contentType:  r.contentType,
		charset:      r.charset,
		timeout:      r.timeout,
		maxRedirects: r.maxRedirects,
		userAgent:    r.userAgent,
		transport:    r.transport,
		transportFn:  r.transportFn,
		basicUser:    r.basicUser,
		basicPass:    r.basicPass,
		hasBasic:     r.hasBasic,
		httpClient:   r.httpClient,
		newRequest:   r.newRequest,
		multipartNew: r.multipartNew,
		decodeConfig: r.decodeConfig.normalized(),
		urlPolicy:    cloneURLPolicy(r.urlPolicy),
	}
	if r.followRedir != nil {
		follow := *r.followRedir
		clone.followRedir = &follow
	}
	if r.tlsConfig != nil {
		clone.tlsConfig = r.tlsConfig.Clone()
	}
	if len(r.cookies) > 0 {
		clone.cookies = make([]*http.Cookie, 0, len(r.cookies))
		for _, cookie := range r.cookies {
			if cookie == nil {
				clone.cookies = append(clone.cookies, nil)
				continue
			}
			copied := *cookie
			clone.cookies = append(clone.cookies, &copied)
		}
	}
	if len(r.multipartFs) > 0 {
		clone.multipartFs = make([]*formFile, 0, len(r.multipartFs))
		for _, file := range r.multipartFs {
			clone.multipartFs = append(clone.multipartFs, file.clone())
		}
	}
	return clone
}

// Method sets the HTTP method.
func (r *HTTPRequest) Method(m Method) *HTTPRequest { r.method = m; return r }

// URL sets the request URL.
func (r *HTTPRequest) URL(u string) *HTTPRequest { r.rawURL = u; return r }

// Header sets a single request header, replacing existing values.
func (r *HTTPRequest) Header(name, value string) *HTTPRequest {
	r.headers.Set(name, value)
	return r
}

// AddHeader appends a single request header value.
func (r *HTTPRequest) AddHeader(name, value string) *HTTPRequest {
	r.headers.Add(name, value)
	return r
}

// Headers sets request headers in batch.
func (r *HTTPRequest) Headers(h map[string]string) *HTTPRequest {
	for k, v := range h {
		r.headers.Set(k, v)
	}
	return r
}

// Cookie adds a Cookie.
func (r *HTTPRequest) Cookie(c *http.Cookie) *HTTPRequest {
	r.cookies = append(r.cookies, c)
	return r
}

// CookieString adds a Cookie header from a raw string.
func (r *HTTPRequest) CookieString(s string) *HTTPRequest {
	r.headers.Set(string(HeaderCookie), s)
	return r
}

// ContentType sets Content-Type.
func (r *HTTPRequest) ContentType(ct string) *HTTPRequest {
	r.contentType = ct
	return r
}

// Charset sets the request charset.
func (r *HTTPRequest) Charset(c string) *HTTPRequest { r.charset = c; return r }

// Timeout sets the request timeout.
func (r *HTTPRequest) Timeout(d time.Duration) *HTTPRequest { r.timeout = d; return r }

// FollowRedirects sets whether redirects are followed.
func (r *HTTPRequest) FollowRedirects(b bool) *HTTPRequest {
	r.followRedir = &b
	return r
}

// MaxRedirects sets the maximum redirect count.
func (r *HTTPRequest) MaxRedirects(n int) *HTTPRequest { r.maxRedirects = n; return r }

// TLSConfig sets a custom TLS config for the generated HTTP transport.
func (r *HTTPRequest) TLSConfig(cfg *tls.Config) *HTTPRequest { r.tlsConfig = cfg; return r }

// Transport sets a custom RoundTripper.
func (r *HTTPRequest) Transport(t http.RoundTripper) *HTTPRequest {
	r.transport = t
	r.transportFn = nil
	return r
}

// Client sets a custom *http.Client, overriding Transport, Timeout, and related options.
func (r *HTTPRequest) Client(c *http.Client) *HTTPRequest { r.httpClient = c; return r }

// URLPolicy sets SSRF-oriented validation for this request.
func (r *HTTPRequest) URLPolicy(policy URLPolicy) *HTTPRequest {
	r.urlPolicy = cloneURLPolicy(&policy)
	return r
}

// BasicAuth sets Basic Auth.
func (r *HTTPRequest) BasicAuth(user, pass string) *HTTPRequest {
	r.basicUser = user
	r.basicPass = pass
	r.hasBasic = true
	return r
}

// BearerAuth sets the Bearer Token.
func (r *HTTPRequest) BearerAuth(token string) *HTTPRequest {
	r.headers.Set(string(HeaderAuthorization), "Bearer "+token)
	return r
}

// Query adds a single URL query parameter.
func (r *HTTPRequest) Query(key string, value any) *HTTPRequest {
	r.queryParams.Add(key, toString(value))
	return r
}

// QueryMap sets URL query parameters in batch.
func (r *HTTPRequest) QueryMap(m map[string]any) *HTTPRequest {
	for k, v := range m {
		r.queryParams.Set(k, toString(v))
	}
	return r
}

// Body sets the raw request body.
func (r *HTTPRequest) Body(body []byte) *HTTPRequest {
	r.body = body
	r.bodyReader = nil
	if r.contentType == "" {
		if ct := GuessContentType(string(body)); ct != "" {
			r.contentType = ct.WithCharset(r.charset)
		}
	}
	return r
}

// BodyString sets a string request body.
func (r *HTTPRequest) BodyString(s string) *HTTPRequest { return r.Body([]byte(s)) }

// BodyJSON sets a JSON request body; callers should serialize values or pass a string.
func (r *HTTPRequest) BodyJSON(s string) *HTTPRequest {
	r.contentType = ContentTypeJSON.WithCharset(r.charset)
	return r.Body([]byte(s))
}

// BodyReader sets the request body from an io.Reader.
func (r *HTTPRequest) BodyReader(reader io.Reader) *HTTPRequest {
	r.bodyReader = reader
	r.body = nil
	return r
}

// Form sets form parameters; it defaults to form-urlencoded and switches to multipart when files exist.
func (r *HTTPRequest) Form(m map[string]any) *HTTPRequest {
	if r.form == nil {
		r.form = make(map[string]any)
	}
	for k, v := range m {
		r.form[k] = v
	}
	return r
}

// FormFile adds a file upload field and enables multipart automatically.
func (r *HTTPRequest) FormFile(field, fileName string, data []byte) *HTTPRequest {
	r.multipart = true
	r.multipartFs = append(r.multipartFs, &formFile{
		field: field, fileName: fileName, data: data,
	})
	return r
}

// FormFileReader adds a file upload field from a reader.
func (r *HTTPRequest) FormFileReader(field, fileName string, reader io.Reader) *HTTPRequest {
	r.multipart = true
	r.multipartFs = append(r.multipartFs, &formFile{
		field: field, fileName: fileName, reader: reader,
	})
	return r
}

// Execute sends the request and returns the response.
func (r *HTTPRequest) Execute() *HTTPResponse {
	resp, err := r.doExecute()
	if err != nil {
		return &HTTPResponse{err: err}
	}
	return resp
}

// MustExecute sends the request and panics on failure.
func (r *HTTPRequest) MustExecute() *HTTPResponse {
	resp := r.Execute()
	if resp.err != nil {
		panic(resp.err)
	}
	return resp
}

func (r *HTTPRequest) buildURL() (string, error) {
	u, err := url.Parse(r.rawURL)
	if err != nil {
		return "", NewHTTPErrorWithCode(knifer.ErrCodeInvalidInput, "invalid url", err)
	}
	if len(r.queryParams) > 0 {
		q := u.Query()
		// Keep a stable output order.
		keys := make([]string, 0, len(r.queryParams))
		for k := range r.queryParams {
			keys = append(keys, k)
		}
		slices.Sort(keys)
		for _, k := range keys {
			for _, v := range r.queryParams[k] {
				q.Add(k, v)
			}
		}
		u.RawQuery = q.Encode()
	}
	return u.String(), nil
}

func (r *HTTPRequest) prepareBody() (io.Reader, string, error) {
	switch {
	case r.bodyReader != nil:
		return r.bodyReader, r.contentType, nil
	case len(r.body) > 0:
		return bytes.NewReader(r.body), r.contentType, nil
	case r.multipart || len(r.multipartFs) > 0:
		reader, ct, err := buildMultipartBody(r.form, r.multipartFs, r.multipartNew)
		if err != nil {
			return nil, "", err
		}
		return reader, ct, nil
	case len(r.form) > 0 && (r.method == MethodPost || r.method == MethodPut || r.method == MethodPatch):
		values := url.Values{}
		for k, v := range r.form {
			values.Set(k, toString(v))
		}
		ct := r.contentType
		if ct == "" {
			ct = ContentTypeFormURLEncoded.WithCharset(r.charset)
		}
		return strings.NewReader(values.Encode()), ct, nil
	case len(r.form) > 0:
		// GET and similar methods: merge form values into query.
		for k, v := range r.form {
			r.queryParams.Add(k, toString(v))
		}
		r.form = nil
		return nil, r.contentType, nil
	}
	return nil, r.contentType, nil
}

func cloneValues(values url.Values) url.Values {
	out := url.Values{}
	for k, v := range values {
		out[k] = slices.Clone(v)
	}
	return out
}

func cloneAnyMap(m map[string]any) map[string]any {
	return maps.Clone(m)
}

func cloneURLPolicy(policy *URLPolicy) *URLPolicy {
	if policy == nil {
		return nil
	}
	p := *policy
	p.AllowedSchemes = slices.Clone(policy.AllowedSchemes)
	p.AllowedHosts = slices.Clone(policy.AllowedHosts)
	return &p
}

func (r *HTTPRequest) hasStreamingBody() bool {
	if r.bodyReader != nil {
		return true
	}
	for _, file := range r.multipartFs {
		if file != nil && file.reader != nil {
			return true
		}
	}
	return false
}

func (r *HTTPRequest) beginExecution() error {
	if !r.hasStreamingBody() {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.used {
		return HTTPErrorfWithCode(knifer.ErrCodeUnsupported, "request contains reader-backed body and cannot be executed more than once; clone it and provide a fresh reader")
	}
	r.used = true
	return nil
}

func (r *HTTPRequest) buildClient() *http.Client {
	if r.httpClient != nil {
		if r.urlPolicy == nil {
			return r.httpClient
		}
		clone := *r.httpClient
		clone.Transport = safeHTTPTransport(clone.Transport, r.urlPolicy)
		clone.CheckRedirect = redirectPolicy(r.followRedir, r.maxRedirects, r.urlPolicy, r.httpClient.CheckRedirect)
		return &clone
	}
	transport := r.transport
	if transport == nil && r.transportFn != nil {
		transport = r.transportFn()
	}
	if transport == nil {
		baseTransport := getDefaultTransport()
		if r.tlsConfig != nil {
			t := baseTransport.Clone()
			t.TLSClientConfig = r.tlsConfig.Clone()
			transport = t
		} else {
			transport = baseTransport
		}
	}
	timeout := r.timeout
	follow := true
	if r.followRedir != nil {
		follow = *r.followRedir
	}
	max := r.maxRedirects
	if r.urlPolicy != nil {
		transport = safeHTTPTransport(transport, r.urlPolicy)
	}
	c := &http.Client{
		Timeout:       timeout,
		Transport:     transport,
		Jar:           r.cookieJar,
		CheckRedirect: redirectPolicy(&follow, max, r.urlPolicy, nil),
	}
	return c
}

func (r *HTTPRequest) doExecute() (*HTTPResponse, error) {
	finalURL, err := r.buildURL()
	if err != nil {
		return nil, err
	}
	if r.urlPolicy != nil {
		parsed, err := url.Parse(finalURL)
		if err != nil {
			return nil, NewHTTPErrorWithCode(knifer.ErrCodeInvalidInput, "invalid url", err)
		}
		if err := validateRequestURL(parsed, r.urlPolicy); err != nil {
			return nil, err
		}
	}
	if err := r.beginExecution(); err != nil {
		return nil, err
	}
	hadForm := len(r.form) > 0
	bodyReader, ct, err := r.prepareBody()
	if err != nil {
		return nil, err
	}
	// prepareBody may modify query values, so build the URL again.
	if hadForm {
		finalURL, err = r.buildURL()
		if err != nil {
			return nil, err
		}
		if r.urlPolicy != nil {
			parsed, err := url.Parse(finalURL)
			if err != nil {
				return nil, NewHTTPErrorWithCode(knifer.ErrCodeInvalidInput, "invalid url", err)
			}
			if err := validateRequestURL(parsed, r.urlPolicy); err != nil {
				return nil, err
			}
		}
	}

	newRequest := r.newRequest
	if newRequest == nil {
		newRequest = http.NewRequest
	}
	req, err := newRequest(string(r.method), finalURL, bodyReader)
	if err != nil {
		return nil, NewHTTPErrorWithCode(knifer.ErrCodeInvalidInput, "build request failed", err)
	}
	for k, vs := range r.headers {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	if ct != "" {
		req.Header.Set(string(HeaderContentType), ct)
	}
	if ua := r.userAgent; ua != "" && req.Header.Get(string(HeaderUserAgent)) == "" {
		req.Header.Set(string(HeaderUserAgent), ua)
	}
	for _, c := range r.cookies {
		req.AddCookie(c)
	}
	if r.hasBasic {
		token := base64.StdEncoding.EncodeToString([]byte(r.basicUser + ":" + r.basicPass))
		req.Header.Set(string(HeaderAuthorization), "Basic "+token)
	}

	client := r.buildClient()
	resp, err := client.Do(req)
	if err != nil {
		return nil, NewHTTPError("send request failed", err)
	}
	return wrapResponse(resp, r.decodeConfig), nil
}

func redirectPolicy(followRedir *bool, max int, policy *URLPolicy, next func(*http.Request, []*http.Request) error) func(*http.Request, []*http.Request) error {
	follow := true
	if followRedir != nil {
		follow = *followRedir
	}
	return func(req *http.Request, via []*http.Request) error {
		if !follow {
			return http.ErrUseLastResponse
		}
		if max > 0 && len(via) >= max {
			return HTTPErrorfWithCode(knifer.ErrCodeUnsupported, "stopped after %d redirects", max)
		}
		if req == nil || req.URL == nil {
			return HTTPErrorfWithCode(knifer.ErrCodeUnsafeResource, "redirect url is nil")
		}
		if err := validateRequestURL(req.URL, policy); err != nil {
			return err
		}
		if next != nil {
			return next(req, via)
		}
		return nil
	}
}

func safeHTTPTransport(base http.RoundTripper, policy *URLPolicy) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	if transport, ok := base.(*http.Transport); ok {
		transportClone := transport.Clone()
		transportClone.DialContext = safeDialContext(policy)
		base = transportClone
	}
	return safeRoundTripper{base: base, policy: policy}
}

func safeDialContext(policy *URLPolicy) func(context.Context, string, string) (net.Conn, error) {
	dialer := &net.Dialer{}
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}
		host = strings.ToLower(strings.TrimSpace(host))
		if host == "" {
			return nil, HTTPErrorfWithCode(knifer.ErrCodeUnsafeResource, "dial host is blank")
		}
		if policy == nil || !policy.RejectPrivate {
			return dialer.DialContext(ctx, network, address)
		}
		ips, err := publicHostIPs(ctx, policy, host)
		if err != nil {
			return nil, err
		}
		return dialer.DialContext(ctx, network, net.JoinHostPort(ips[0].String(), port))
	}
}

type safeRoundTripper struct {
	base   http.RoundTripper
	policy *URLPolicy
}

func (t safeRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req == nil || req.URL == nil {
		return nil, HTTPErrorfWithCode(knifer.ErrCodeUnsafeResource, "request url is nil")
	}
	if err := validateRequestURL(req.URL, t.policy); err != nil {
		return nil, err
	}
	if req.URL.Scheme == "http" || req.URL.Scheme == "https" {
		host := strings.ToLower(req.URL.Hostname())
		if t.policy != nil && t.policy.RejectPrivate {
			private, err := isPrivateHost(req.Context(), t.policy.LookupIP, host)
			if err != nil {
				return nil, NewHTTPErrorWithCode(knifer.ErrCodeUnsafeResource, "resolve url host failed", err)
			}
			if private {
				return nil, HTTPErrorfWithCode(knifer.ErrCodeUnsafeResource, "url host %q resolves to a private address", host)
			}
		}
	}
	return t.base.RoundTrip(req)
}

func validateRequestURL(u *url.URL, policy *URLPolicy) error {
	if policy == nil {
		return nil
	}
	if u == nil {
		return HTTPErrorfWithCode(knifer.ErrCodeUnsafeResource, "url is nil")
	}
	scheme := strings.ToLower(strings.TrimSpace(u.Scheme))
	if len(policy.AllowedSchemes) > 0 && !containsFold(policy.AllowedSchemes, scheme) {
		return HTTPErrorfWithCode(knifer.ErrCodeUnsafeResource, "url scheme %q is not allowed", scheme)
	}
	if scheme != "http" && scheme != "https" {
		return nil
	}
	host := strings.ToLower(u.Hostname())
	if host == "" {
		return HTTPErrorfWithCode(knifer.ErrCodeUnsafeResource, "http url host is blank")
	}
	if len(policy.AllowedHosts) > 0 && !containsFold(policy.AllowedHosts, host) {
		return HTTPErrorfWithCode(knifer.ErrCodeUnsafeResource, "url host %q is not allowed", host)
	}
	if policy.RejectPrivate {
		private, err := isPrivateHost(context.Background(), policy.LookupIP, host)
		if err != nil {
			return NewHTTPErrorWithCode(knifer.ErrCodeUnsafeResource, "resolve url host failed", err)
		}
		if private {
			return HTTPErrorfWithCode(knifer.ErrCodeUnsafeResource, "url host %q resolves to a private address", host)
		}
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
	return httpboundary.IsPrivateHost(ctx, lookupIP, host)
}

func defaultLookupIP(ctx context.Context, host string) ([]net.IP, error) {
	return httpboundary.DefaultLookupIP(ctx, host)
}

func publicHostIPs(ctx context.Context, policy *URLPolicy, host string) ([]net.IP, error) {
	lookupIP := defaultLookupIP
	if policy != nil && policy.LookupIP != nil {
		lookupIP = policy.LookupIP
	}
	ips, err := httpboundary.PublicHostIPs(ctx, lookupIP, host)
	if err != nil {
		if errors.Is(err, httpboundary.ErrPrivateHost) {
			return nil, HTTPErrorfWithCode(knifer.ErrCodeUnsafeResource, "url host %q resolves to a private address", host)
		}
		if errors.Is(err, httpboundary.ErrNoAddresses) {
			return nil, HTTPErrorfWithCode(knifer.ErrCodeUnsafeResource, "resolve url host %q: no addresses", host)
		}
		return nil, NewHTTPErrorWithCode(knifer.ErrCodeUnsafeResource, "resolve url host failed", err)
	}
	return ips, nil
}

func toString(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case fmt.Stringer:
		return x.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}
