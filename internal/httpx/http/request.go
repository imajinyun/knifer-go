package http

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

var defaultTransport = http.DefaultTransport.(*http.Transport).Clone()

// HTTPRequest is a chainable HTTP request builder, aligned with the utility toolkit-http HttpRequest.
type HTTPRequest struct {
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
	tlsSkip      bool
	tlsConfig    *tls.Config
	userAgent    string
	transport    http.RoundTripper
	basicUser    string
	basicPass    string
	hasBasic     bool
	httpClient   *http.Client
}

type formFile struct {
	field    string
	fileName string
	data     []byte
	reader   io.Reader
}

// RequestOption customizes one HTTP request at construction time.
type RequestOption func(*HTTPRequest)

// NewRequest creates a request with the specified method and URL.
func NewRequest(method Method, rawURL string, opts ...RequestOption) *HTTPRequest {
	follow := GetGlobalFollowRedirects()
	r := &HTTPRequest{
		method:       method,
		rawURL:       rawURL,
		queryParams:  url.Values{},
		headers:      CloneGlobalHeaders(),
		cookieJar:    GetCookieJar(),
		charset:      "UTF-8",
		timeout:      GetGlobalTimeout(),
		followRedir:  &follow,
		maxRedirects: GetGlobalMaxRedirects(),
		tlsSkip:      IsTrustAnyHost(),
		userAgent:    GetGlobalUserAgent(),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(r)
		}
	}
	return r
}

// Get creates a GET request.
func Get(rawURL string, opts ...RequestOption) *HTTPRequest {
	return NewRequest(MethodGet, rawURL, opts...)
}

// Post creates a POST request.
func Post(rawURL string, opts ...RequestOption) *HTTPRequest {
	return NewRequest(MethodPost, rawURL, opts...)
}

// Put creates a PUT request.
func Put(rawURL string, opts ...RequestOption) *HTTPRequest {
	return NewRequest(MethodPut, rawURL, opts...)
}

// Delete creates a DELETE request.
func Delete(rawURL string, opts ...RequestOption) *HTTPRequest {
	return NewRequest(MethodDelete, rawURL, opts...)
}

// Patch creates a PATCH request.
func Patch(rawURL string, opts ...RequestOption) *HTTPRequest {
	return NewRequest(MethodPatch, rawURL, opts...)
}

// Head creates a HEAD request.
func Head(rawURL string, opts ...RequestOption) *HTTPRequest {
	return NewRequest(MethodHead, rawURL, opts...)
}

// Options creates an OPTIONS request.
func Options(rawURL string, opts ...RequestOption) *HTTPRequest {
	return NewRequest(MethodOptions, rawURL, opts...)
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

// WithSkipTLSVerify sets per-request TLS verification behavior.
func WithSkipTLSVerify(b bool) RequestOption { return func(r *HTTPRequest) { r.SkipTLSVerify(b) } }

// WithTLSConfig sets a per-request TLS config. It is ignored when a custom client or transport is set.
func WithTLSConfig(cfg *tls.Config) RequestOption { return func(r *HTTPRequest) { r.TLSConfig(cfg) } }

// WithTransport sets a per-request RoundTripper.
func WithTransport(t http.RoundTripper) RequestOption { return func(r *HTTPRequest) { r.Transport(t) } }

// WithClient sets a per-request HTTP client.
func WithClient(c *http.Client) RequestOption { return func(r *HTTPRequest) { r.Client(c) } }

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

// SkipTLSVerify skips TLS certificate verification.
func (r *HTTPRequest) SkipTLSVerify(b bool) *HTTPRequest { r.tlsSkip = b; return r }

// TLSConfig sets a custom TLS config for the generated HTTP transport.
func (r *HTTPRequest) TLSConfig(cfg *tls.Config) *HTTPRequest { r.tlsConfig = cfg; return r }

// Transport sets a custom RoundTripper.
func (r *HTTPRequest) Transport(t http.RoundTripper) *HTTPRequest { r.transport = t; return r }

// Client sets a custom *http.Client, overriding Transport, Timeout, and related options.
func (r *HTTPRequest) Client(c *http.Client) *HTTPRequest { r.httpClient = c; return r }

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
		return "", NewHTTPError("invalid url", err)
	}
	if len(r.queryParams) > 0 {
		q := u.Query()
		// Keep a stable output order.
		keys := make([]string, 0, len(r.queryParams))
		for k := range r.queryParams {
			keys = append(keys, k)
		}
		sort.Strings(keys)
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
		reader, ct, err := buildMultipartBody(r.form, r.multipartFs)
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

func (r *HTTPRequest) buildClient() *http.Client {
	if r.httpClient != nil {
		return r.httpClient
	}
	transport := r.transport
	if transport == nil {
		if r.tlsSkip || r.tlsConfig != nil {
			t := defaultTransport.Clone()
			if r.tlsConfig != nil {
				t.TLSClientConfig = r.tlsConfig.Clone()
			} else {
				t.TLSClientConfig = &tls.Config{}
			}
			if r.tlsSkip {
				t.TLSClientConfig.InsecureSkipVerify = true
			}
			transport = t
		} else {
			transport = defaultTransport
		}
	}
	timeout := r.timeout
	follow := true
	if r.followRedir != nil {
		follow = *r.followRedir
	}
	max := r.maxRedirects
	c := &http.Client{
		Timeout:   timeout,
		Transport: transport,
		Jar:       r.cookieJar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if !follow {
				return http.ErrUseLastResponse
			}
			if max > 0 && len(via) >= max {
				return HTTPErrorf("stopped after %d redirects", max)
			}
			return nil
		},
	}
	return c
}

func (r *HTTPRequest) doExecute() (*HTTPResponse, error) {
	finalURL, err := r.buildURL()
	if err != nil {
		return nil, err
	}
	bodyReader, ct, err := r.prepareBody()
	if err != nil {
		return nil, err
	}
	// prepareBody may modify query values, so build the URL again.
	if r.form != nil {
		finalURL, err = r.buildURL()
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(string(r.method), finalURL, bodyReader)
	if err != nil {
		return nil, NewHTTPError("build request failed", err)
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
