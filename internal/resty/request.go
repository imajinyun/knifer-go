package resty

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/url"
	"sort"
	"time"

	grestry "resty.dev/v3"
)

// HTTPRequest is a chainable HTTP request builder backed by go-resty/resty.
type HTTPRequest struct {
	method       Method
	rawURL       string
	queryParams  url.Values
	headers      HeaderValues
	body         any
	form         map[string]any
	files        []*formFile
	contentType  string
	charset      string
	timeout      time.Duration
	followRedir  *bool
	maxRedirects int
	tlsSkip      bool
	basicUser    string
	basicPass    string
	hasBasic     bool
	restyClient  *grestry.Client
}

type formFile struct {
	field    string
	fileName string
	data     []byte
	reader   io.Reader
}

// NewRequest creates a request with the specified method and URL.
func NewRequest(method Method, rawURL string) *HTTPRequest {
	return &HTTPRequest{
		method:       method,
		rawURL:       rawURL,
		queryParams:  url.Values{},
		headers:      CloneGlobalHeaders(),
		charset:      "UTF-8",
		maxRedirects: GetGlobalMaxRedirects(),
	}
}

// Get creates a GET request.
func Get(rawURL string) *HTTPRequest { return NewRequest(MethodGet, rawURL) }

// Post creates a POST request.
func Post(rawURL string) *HTTPRequest { return NewRequest(MethodPost, rawURL) }

// Put creates a PUT request.
func Put(rawURL string) *HTTPRequest { return NewRequest(MethodPut, rawURL) }

// Delete creates a DELETE request.
func Delete(rawURL string) *HTTPRequest { return NewRequest(MethodDelete, rawURL) }

// Patch creates a PATCH request.
func Patch(rawURL string) *HTTPRequest { return NewRequest(MethodPatch, rawURL) }

// Head creates a HEAD request.
func Head(rawURL string) *HTTPRequest { return NewRequest(MethodHead, rawURL) }

// Options creates an OPTIONS request.
func Options(rawURL string) *HTTPRequest { return NewRequest(MethodOptions, rawURL) }

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

// SkipTLSVerify skips TLS certificate verification.
func (r *HTTPRequest) SkipTLSVerify(b bool) *HTTPRequest { r.tlsSkip = b; return r }

// RestyClient sets a custom resty client.
func (r *HTTPRequest) RestyClient(c *grestry.Client) *HTTPRequest { r.restyClient = c; return r }

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
		return r.restyClient
	}
	c := grestry.New()
	if isCookieDisabled() {
		c.SetCookieJar(nil)
	}
	timeout := r.timeout
	if timeout == 0 {
		timeout = GetGlobalTimeout()
	}
	if timeout > 0 {
		c.SetTimeout(timeout)
	}
	if r.tlsSkip || IsTrustAnyHost() {
		c.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}) // #nosec G402 -- caller explicitly requested skipping TLS verification.
	}
	follow := GetGlobalFollowRedirects()
	if r.followRedir != nil {
		follow = *r.followRedir
	}
	if !follow {
		c.SetRedirectPolicy(grestry.NoRedirectPolicy())
	} else if r.maxRedirects > 0 {
		c.SetRedirectPolicy(grestry.FlexibleRedirectPolicy(r.maxRedirects))
	}
	return c
}

func (r *HTTPRequest) doExecute() (*HTTPResponse, error) {
	if _, err := url.Parse(r.rawURL); err != nil {
		return nil, NewHTTPError("invalid url", err)
	}
	req := r.buildClient().R()
	for k, vs := range r.headers {
		for _, v := range vs {
			req.SetHeader(k, v)
		}
	}
	if r.contentType != "" {
		req.SetHeader(string(HeaderContentType), r.contentType)
	}
	if ua := GetGlobalUserAgent(); ua != "" && getHeader(r.headers, string(HeaderUserAgent)) == "" {
		req.SetHeader(string(HeaderUserAgent), ua)
	}
	if r.hasBasic {
		req.SetBasicAuth(r.basicUser, r.basicPass)
	}
	keys := make([]string, 0, len(r.queryParams))
	for k := range r.queryParams {
		keys = append(keys, k)
	}
	sort.Strings(keys)
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
