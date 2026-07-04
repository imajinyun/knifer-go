package resty

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"

	knifer "github.com/imajinyun/knifer-go"
	"github.com/imajinyun/knifer-go/internal/httpboundary"
	grestry "resty.dev/v3"
)

// HTTPRequest is a chainable HTTP request builder backed by go-resty/resty.
type HTTPRequest struct {
	method        Method
	rawURL        string
	queryParams   url.Values
	headers       HeaderValues
	body          any
	form          map[string]any
	files         []*formFile
	contentType   string
	charset       string
	timeout       time.Duration
	followRedir   *bool
	maxRedirects  int
	tlsConfig     *tls.Config
	userAgent     string
	cookieOff     bool
	basicUser     string
	basicPass     string
	hasBasic      bool
	restyClient   *grestry.Client
	clientFactory func() *grestry.Client
	jsonMarshal   func(any) ([]byte, error)
	jsonUnmarshal func([]byte, any) error
	jsonReadAll   func(io.Reader) ([]byte, error)
	maxDecode     int64
	maxResponse   int64
	urlPolicy     *URLPolicy
	result        any
	errorResult   any
}

type formFile struct {
	field    string
	fileName string
	data     []byte
	reader   io.Reader
}

// URLPolicy controls SSRF-oriented request validation for untrusted URLs.
type URLPolicy struct {
	AllowedSchemes []string
	AllowedHosts   []string
	RejectPrivate  bool
	LookupIP       func(context.Context, string) ([]net.IP, error)
}

var defaultRestyClientProvider = struct {
	sync.RWMutex
	provider func() *grestry.Client
}{provider: grestry.New}

// RequestOption customizes one HTTP request at construction time.
type RequestOption func(*HTTPRequest)

// Client is an explicit resty request factory with a captured configuration snapshot.
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

// NewRequest creates a request with the specified method and URL.
//
// Security: NewRequest is for trusted URLs unless callers provide WithURLPolicy
// with RejectPrivate enabled. Use NewSafeRequest, GetSafe, or PostSafe when the
// URL may come from users, config, or another untrusted trust boundary.
func NewRequest(method Method, rawURL string, opts ...RequestOption) *HTTPRequest {
	return NewRequestWithConfig(method, rawURL, SnapshotGlobalConfig(), opts...)
}

// NewIsolatedRequest creates a request without reading package-level global defaults.
func NewIsolatedRequest(method Method, rawURL string, opts ...RequestOption) *HTTPRequest {
	return NewRequestWithConfig(method, rawURL, isolatedGlobalConfig(), opts...)
}

// NewRequestWithConfig creates a request from an explicit global configuration snapshot.
//
// Security: NewRequestWithConfig is for trusted URLs unless callers provide
// WithURLPolicy with RejectPrivate enabled. Use NewSafeRequest for untrusted
// URLs.
func NewRequestWithConfig(method Method, rawURL string, cfg GlobalConfig, opts ...RequestOption) *HTTPRequest {
	follow := cfg.FollowRedirects
	r := &HTTPRequest{
		method:       method,
		rawURL:       rawURL,
		queryParams:  url.Values{},
		headers:      cloneHeaders(cfg.Headers),
		charset:      "UTF-8",
		timeout:      cfg.Timeout,
		followRedir:  &follow,
		maxRedirects: cfg.MaxRedirects,
		maxResponse:  cfg.MaxResponseBytes,
		userAgent:    cfg.DefaultUserAgent,
		cookieOff:    cfg.CookieDisabled,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(r)
		}
	}
	return r
}

// WithGlobalConfig initializes request defaults from a captured global configuration snapshot.
func WithGlobalConfig(cfg GlobalConfig) RequestOption {
	return func(r *HTTPRequest) {
		follow := cfg.FollowRedirects
		r.headers = cloneHeaders(cfg.Headers)
		r.timeout = cfg.Timeout
		r.followRedir = &follow
		r.maxRedirects = cfg.MaxRedirects
		r.maxResponse = cfg.MaxResponseBytes
		r.userAgent = cfg.DefaultUserAgent
		r.cookieOff = cfg.CookieDisabled
	}
}

// Get creates a GET request.
//
// Security: Get is for trusted URLs. Use GetSafe for untrusted URLs.
func Get(rawURL string, opts ...RequestOption) *HTTPRequest {
	return NewRequest(MethodGet, rawURL, opts...)
}

// GetSafe creates a GET request with SSRF-oriented safety checks enabled.
func GetSafe(rawURL string, opts ...RequestOption) *HTTPRequest {
	return NewSafeRequest(MethodGet, rawURL, opts...)
}

// Post creates a POST request.
//
// Security: Post is for trusted URLs. Use PostSafe for untrusted URLs.
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

// WithTLSConfig sets a per-request TLS config. It is ignored when a custom resty client is set.
func WithTLSConfig(cfg *tls.Config) RequestOption { return func(r *HTTPRequest) { r.TLSConfig(cfg) } }

// WithRestyClient sets a per-request resty client.
func WithRestyClient(c *grestry.Client) RequestOption {
	return func(r *HTTPRequest) {
		if c != nil {
			r.RestyClient(c)
		}
	}
}

// WithRestyClientFactory sets a per-request resty client factory.
func WithRestyClientFactory(factory func() *grestry.Client) RequestOption {
	return func(r *HTTPRequest) {
		if factory != nil {
			r.clientFactory = factory
		}
	}
}

// ConfigureDefaultRestyClientProvider sets the provider used to create resty clients when no per-request client is set.
// Passing nil restores resty.New.
func ConfigureDefaultRestyClientProvider(provider func() *grestry.Client) {
	defaultRestyClientProvider.Lock()
	defer defaultRestyClientProvider.Unlock()
	if provider == nil {
		defaultRestyClientProvider.provider = grestry.New
		return
	}
	defaultRestyClientProvider.provider = provider
}

// ResetDefaultRestyClientProvider restores resty.New as the default client provider.
func ResetDefaultRestyClientProvider() { ConfigureDefaultRestyClientProvider(nil) }

// WithUserAgent sets a per-request User-Agent.
func WithUserAgent(ua string) RequestOption {
	return func(r *HTTPRequest) {
		r.userAgent = ua
		setHeader(r.headers, string(HeaderUserAgent), ua)
	}
}

// WithCookieDisabled sets per-request cookie management behavior.
func WithCookieDisabled(disabled bool) RequestOption {
	return func(r *HTTPRequest) { r.cookieOff = disabled }
}

// WithContentType sets a per-request Content-Type at construction time.
func WithContentType(ct string) RequestOption { return func(r *HTTPRequest) { r.ContentType(ct) } }

// WithCharset sets a per-request charset at construction time.
func WithCharset(charset string) RequestOption { return func(r *HTTPRequest) { r.Charset(charset) } }

// WithJSONMarshalFunc sets the JSON marshal provider used by request body encoding.
func WithJSONMarshalFunc(marshal func(any) ([]byte, error)) RequestOption {
	return func(r *HTTPRequest) {
		if marshal != nil {
			r.jsonMarshal = marshal
		}
	}
}

// WithJSONUnmarshalFunc sets the JSON unmarshal provider used by response decoding.
func WithJSONUnmarshalFunc(unmarshal func([]byte, any) error) RequestOption {
	return func(r *HTTPRequest) {
		if unmarshal != nil {
			r.jsonUnmarshal = unmarshal
		}
	}
}

// WithJSONDecodeReadAllFunc sets the reader used before custom JSON unmarshalling.
func WithJSONDecodeReadAllFunc(readAll func(io.Reader) ([]byte, error)) RequestOption {
	return func(r *HTTPRequest) {
		if readAll != nil {
			r.jsonReadAll = readAll
		}
	}
}

// WithMaxDecodeBytes limits bytes read before custom JSON unmarshalling. Non-positive means unlimited.
func WithMaxDecodeBytes(maxBytes int64) RequestOption {
	return func(r *HTTPRequest) { r.maxDecode = maxBytes }
}

// WithMaxResponseBytes limits response bytes read into memory. Non-positive means unlimited.
func WithMaxResponseBytes(maxBytes int64) RequestOption {
	return func(r *HTTPRequest) { r.maxResponse = maxBytes }
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

// Method sets the HTTP method.
func (r *HTTPRequest) Method(m Method) *HTTPRequest { r.method = m; return r }

// URL sets the request URL.
func (r *HTTPRequest) URL(u string) *HTTPRequest { r.rawURL = u; return r }

// Header sets a single request header, replacing existing values.
func (r *HTTPRequest) Header(name, value string) *HTTPRequest {
	setHeader(r.headers, name, value)
	return r
}

// AddHeader appends a single request header value.
func (r *HTTPRequest) AddHeader(name, value string) *HTTPRequest {
	r.headers[name] = append(r.headers[name], value)
	return r
}

// Headers sets request headers in batch.
func (r *HTTPRequest) Headers(h map[string]string) *HTTPRequest {
	for k, v := range h {
		setHeader(r.headers, k, v)
	}
	return r
}

// Cookie adds a cookie by name and value.
func (r *HTTPRequest) Cookie(name, value string) *HTTPRequest {
	if name == "" {
		return r
	}
	r.AddHeader(string(HeaderCookie), name+"="+value)
	return r
}

// CookieString adds a Cookie header from a raw string.
func (r *HTTPRequest) CookieString(s string) *HTTPRequest {
	setHeader(r.headers, string(HeaderCookie), s)
	return r
}

// ContentType sets Content-Type.
func (r *HTTPRequest) ContentType(ct string) *HTTPRequest { r.contentType = ct; return r }

// Charset sets the request charset.
func (r *HTTPRequest) Charset(c string) *HTTPRequest { r.charset = c; return r }

// Timeout sets the request timeout.
func (r *HTTPRequest) Timeout(d time.Duration) *HTTPRequest { r.timeout = d; return r }

// FollowRedirects sets whether redirects are followed.
func (r *HTTPRequest) FollowRedirects(b bool) *HTTPRequest { r.followRedir = &b; return r }

// MaxRedirects sets the maximum redirect count.
func (r *HTTPRequest) MaxRedirects(n int) *HTTPRequest { r.maxRedirects = n; return r }

// TLSConfig sets a custom TLS config for the generated resty client.
func (r *HTTPRequest) TLSConfig(cfg *tls.Config) *HTTPRequest { r.tlsConfig = cfg; return r }

// RestyClient sets a custom resty client.
func (r *HTTPRequest) RestyClient(c *grestry.Client) *HTTPRequest { r.restyClient = c; return r }

// URLPolicy sets SSRF-oriented validation for this request.
func (r *HTTPRequest) URLPolicy(policy URLPolicy) *HTTPRequest {
	p := policy
	p.AllowedSchemes = slices.Clone(policy.AllowedSchemes)
	p.AllowedHosts = slices.Clone(policy.AllowedHosts)
	r.urlPolicy = &p
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
	setHeader(r.headers, string(HeaderAuthorization), "Bearer "+token)
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
	if r.contentType == "" {
		if ct := GuessContentType(string(body)); ct != "" {
			r.contentType = ct.WithCharset(r.charset)
		}
	}
	return r
}

// BodyString sets a string request body.
func (r *HTTPRequest) BodyString(s string) *HTTPRequest { return r.Body([]byte(s)) }

// BodyJSON sets a JSON request body.
func (r *HTTPRequest) BodyJSON(s string) *HTTPRequest {
	r.contentType = ContentTypeJSON.WithCharset(r.charset)
	return r.Body([]byte(s))
}

// BodyJSONValue sets a JSON request body value to be encoded by resty or the configured JSON marshal provider.
func (r *HTTPRequest) BodyJSONValue(v any) *HTTPRequest {
	r.contentType = ContentTypeJSON.WithCharset(r.charset)
	r.body = v
	return r
}

// Result registers a value for automatic response decoding.
func (r *HTTPRequest) Result(v any) *HTTPRequest { r.result = v; return r }

// ErrorResult registers a value for automatic error response decoding.
func (r *HTTPRequest) ErrorResult(v any) *HTTPRequest { r.errorResult = v; return r }

// BodyReader sets the request body from an io.Reader.
func (r *HTTPRequest) BodyReader(reader io.Reader) *HTTPRequest { r.body = reader; return r }

// Form sets form parameters.
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
	r.files = append(r.files, &formFile{field: field, fileName: fileName, data: data})
	return r
}

// FormFileReader adds a file upload field from a reader.
func (r *HTTPRequest) FormFileReader(field, fileName string, reader io.Reader) *HTTPRequest {
	r.files = append(r.files, &formFile{field: field, fileName: fileName, reader: reader})
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

func (r *HTTPRequest) buildClient() *grestry.Client {
	if r.restyClient != nil {
		if r.urlPolicy != nil {
			c := r.restyClient.Clone(context.Background())
			c.SetTransport(safeRestyTransport(c.Transport(), r.urlPolicy))
			return c
		}
		return r.restyClient
	}
	c := r.newRestyClient()
	if r.cookieOff {
		c.SetCookieJar(nil)
	}
	if r.timeout > 0 {
		c.SetTimeout(r.timeout)
	}
	if r.maxResponse != 0 {
		c.SetResponseBodyLimit(r.maxResponse)
	}
	if r.tlsConfig != nil {
		c.SetTLSClientConfig(r.tlsConfig.Clone())
	}
	if r.urlPolicy != nil {
		c.SetTransport(safeRestyTransport(c.Transport(), r.urlPolicy))
	}
	follow := true
	if r.followRedir != nil {
		follow = *r.followRedir
	}
	switch {
	case !follow:
		c.SetRedirectPolicy(grestry.RedirectNoPolicy())
	case r.urlPolicy != nil:
		if r.maxRedirects > 0 {
			c.SetRedirectPolicy(grestry.RedirectFlexiblePolicy(r.maxRedirects), safeRedirectPolicy(r.urlPolicy))
		} else {
			c.SetRedirectPolicy(safeRedirectPolicy(r.urlPolicy))
		}
	case r.maxRedirects > 0:
		c.SetRedirectPolicy(grestry.RedirectFlexiblePolicy(r.maxRedirects))
	}
	if r.jsonMarshal != nil {
		c.AddContentTypeEncoder("json", func(w io.Writer, v any) error {
			data, err := r.jsonMarshal(v)
			if err != nil {
				return err
			}
			_, err = w.Write(data)
			return err
		})
	}
	if r.jsonUnmarshal != nil {
		c.AddContentTypeDecoder("json", func(reader io.Reader, v any) error {
			data, err := readAllWithLimit(reader, r.maxDecode, r.jsonReadAll)
			if err != nil {
				return err
			}
			return r.jsonUnmarshal(data, v)
		})
	}
	return c
}

func (r *HTTPRequest) newRestyClient() *grestry.Client {
	if r.clientFactory != nil {
		if c := r.clientFactory(); c != nil {
			return c
		}
	}
	defaultRestyClientProvider.RLock()
	provider := defaultRestyClientProvider.provider
	defaultRestyClientProvider.RUnlock()
	if provider != nil {
		if c := provider(); c != nil {
			return c
		}
	}
	return grestry.New()
}

func (r *HTTPRequest) doExecute() (*HTTPResponse, error) {
	parsed, err := url.Parse(r.rawURL)
	if err != nil {
		return nil, NewHTTPError("invalid url", err)
	}
	if err := validateRequestURL(parsed, r.urlPolicy); err != nil {
		return nil, err
	}
	req := r.buildClient().R()
	if r.maxResponse != 0 {
		req.SetResponseBodyLimit(r.maxResponse)
	}
	for k, vs := range r.headers {
		for _, v := range vs {
			req.SetHeader(k, v)
		}
	}
	if r.contentType != "" {
		req.SetHeader(string(HeaderContentType), r.contentType)
	}
	if ua := r.userAgent; ua != "" && getHeader(r.headers, string(HeaderUserAgent)) == "" {
		req.SetHeader(string(HeaderUserAgent), ua)
	}
	if r.hasBasic {
		req.SetBasicAuth(r.basicUser, r.basicPass)
	}
	if r.result != nil {
		req.SetResult(r.result)
	}
	if r.errorResult != nil {
		req.SetResultError(r.errorResult)
	}
	keys := make([]string, 0, len(r.queryParams))
	for k := range r.queryParams {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	for _, k := range keys {
		for _, v := range r.queryParams[k] {
			req.SetQueryParam(k, v)
		}
	}
	switch {
	case len(r.files) > 0:
		for k, v := range r.form {
			req.SetFormData(map[string]string{k: toString(v)})
		}
		for _, f := range r.files {
			if f.reader != nil {
				req.SetFileReader(f.field, f.fileName, f.reader)
			} else {
				req.SetFileReader(f.field, f.fileName, bytesReader(f.data))
			}
		}
	case len(r.form) > 0:
		if r.method == MethodPost || r.method == MethodPut || r.method == MethodPatch {
			data := make(map[string]string, len(r.form))
			for k, v := range r.form {
				data[k] = toString(v)
			}
			req.SetFormData(data)
		} else {
			for k, v := range r.form {
				req.SetQueryParam(k, toString(v))
			}
		}
	case r.body != nil:
		req.SetBody(r.body)
	}
	resp, err := req.Execute(string(r.method), r.rawURL)
	if err != nil {
		return nil, NewHTTPError("send request failed", err)
	}
	return wrapResponse(resp), nil
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

func safeRedirectPolicy(policy *URLPolicy) grestry.RedirectPolicy {
	return grestry.RedirectPolicyFunc(func(req *http.Request, via []*http.Request) error {
		if req == nil || req.URL == nil {
			return HTTPErrorfWithCode(knifer.ErrCodeUnsafeResource, "redirect url is nil")
		}
		return validateRequestURL(req.URL, policy)
	})
}

func safeRestyTransport(base http.RoundTripper, policy *URLPolicy) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	if transport, ok := base.(*http.Transport); ok {
		transportClone := transport.Clone()
		transportClone.DialContext = safeDialContext(policy)
		base = transportClone
	}
	return safeRestyRoundTripper{base: base, policy: policy}
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

type safeRestyRoundTripper struct {
	base   http.RoundTripper
	policy *URLPolicy
}

func (t safeRestyRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
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
