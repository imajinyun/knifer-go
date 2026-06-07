package http

import (
	"io"
	"regexp"
	"time"

	"github.com/imajinyun/go-knifer/internal/httpx/internal/shared"
	urlimpl "github.com/imajinyun/go-knifer/internal/url"
)

type CharsetOption = shared.CharsetOption

// WithCharsetRegexp sets the regexp used by GetCharsetFromContentTypeWithOptions.
func WithCharsetRegexp(re *regexp.Regexp) CharsetOption { return shared.WithCharsetRegexp(re) }

// WithMetaCharsetRegexp sets the regexp used by GetCharsetFromHTMLWithOptions.
func WithMetaCharsetRegexp(re *regexp.Regexp) CharsetOption { return shared.WithMetaCharsetRegexp(re) }

// IsHTTPS reports whether the given URL is https.
func IsHTTPS(u string) bool { return urlimpl.IsHTTPS(u) }

// IsHTTP reports whether the given URL is http.
func IsHTTP(u string) bool { return urlimpl.IsHTTP(u) }

// CreateRequest creates a request with the specified method, aligned with HttpUtil.createRequest.
func CreateRequest(method Method, rawURL string, opts ...RequestOption) *HTTPRequest {
	return NewRequest(method, rawURL, opts...)
}

// CreateGet creates a GET request and sets whether redirects are followed.
func CreateGet(rawURL string, followRedirects bool) *HTTPRequest {
	return CreateGetWithOptions(rawURL, followRedirects)
}

// CreateGetWithOptions creates a GET request with options and sets whether redirects are followed.
func CreateGetWithOptions(rawURL string, followRedirects bool, opts ...RequestOption) *HTTPRequest {
	return Get(rawURL, opts...).FollowRedirects(followRedirects)
}

// CreatePost creates a POST request.
func CreatePost(rawURL string) *HTTPRequest { return CreatePostWithOptions(rawURL) }

// CreatePostWithOptions creates a POST request with options.
func CreatePostWithOptions(rawURL string, opts ...RequestOption) *HTTPRequest {
	return Post(rawURL, opts...)
}

// GetString sends a GET request and returns the response body as a string.
func GetString(rawURL string) string { return GetStringWithOptions(rawURL) }

// GetStringWithOptions sends a GET request with options and returns the response body as a string.
func GetStringWithOptions(rawURL string, opts ...RequestOption) string {
	return Get(rawURL, opts...).Execute().Body()
}

// GetWithTimeout sends a GET request with a timeout.
func GetWithTimeout(rawURL string, timeout time.Duration) string {
	return GetWithTimeoutWithOptions(rawURL, timeout)
}

// GetWithTimeoutWithOptions sends a GET request with a timeout and custom options.
func GetWithTimeoutWithOptions(rawURL string, timeout time.Duration, opts ...RequestOption) string {
	return Get(rawURL, opts...).Timeout(timeout).Execute().Body()
}

// GetWithParams sends a GET request with form parameters.
func GetWithParams(rawURL string, params map[string]any) string {
	return GetWithParamsWithOptions(rawURL, params)
}

// GetWithParamsWithOptions sends a GET request with form parameters and custom options.
func GetWithParamsWithOptions(rawURL string, params map[string]any, opts ...RequestOption) string {
	return Get(rawURL, opts...).Form(params).Execute().Body()
}

// PostString sends a POST request with a string body.
func PostString(rawURL, body string) string {
	return PostStringWithOptions(rawURL, body)
}

// PostStringWithOptions sends a POST request with a string body and custom options.
func PostStringWithOptions(rawURL, body string, opts ...RequestOption) string {
	return Post(rawURL, opts...).BodyString(body).Execute().Body()
}

// PostForm sends a POST request with form parameters.
func PostForm(rawURL string, params map[string]any) string {
	return PostFormWithOptions(rawURL, params)
}

// PostFormWithOptions sends a POST request with form parameters and custom options.
func PostFormWithOptions(rawURL string, params map[string]any, opts ...RequestOption) string {
	return Post(rawURL, opts...).Form(params).Execute().Body()
}

// PostJSON sends a POST request with a JSON string body.
func PostJSON(rawURL, jsonStr string) string {
	return PostJSONWithOptions(rawURL, jsonStr)
}

// PostJSONWithOptions sends a POST request with a JSON string body and custom options.
func PostJSONWithOptions(rawURL, jsonStr string, opts ...RequestOption) string {
	return Post(rawURL, opts...).BodyJSON(jsonStr).Execute().Body()
}

// DownloadString downloads remote text and detects charset from response headers when customCharset is empty.
func DownloadString(rawURL, customCharset string) string {
	return DownloadStringWithOptions(rawURL, customCharset)
}

// DownloadStringWithOptions downloads remote text with per-request options.
func DownloadStringWithOptions(rawURL, customCharset string, opts ...RequestOption) string {
	resp := Get(rawURL, opts...).Execute()
	if resp.err != nil {
		return ""
	}
	if customCharset != "" {
		// Go does not provide built-in charset conversion; return bytes directly and let callers convert if needed.
		_ = customCharset
	}
	return resp.Body()
}

// DownloadFile downloads to a file, using URL or response headers for the file name when dest is a directory.
func DownloadFile(rawURL, dest string, opts ...SaveOption) (int64, error) {
	return DownloadFileWithOptions(rawURL, dest, nil, opts...)
}

// DownloadFileWithOptions downloads to a file with per-request and per-save options.
func DownloadFileWithOptions(rawURL, dest string, requestOpts []RequestOption, saveOpts ...SaveOption) (int64, error) {
	resp := Get(rawURL, requestOpts...).Execute()
	if resp.err != nil {
		return 0, resp.err
	}
	return resp.SaveAs(dest, saveOpts...)
}

// Download downloads to a Writer.
func Download(rawURL string, w io.Writer) (int64, error) {
	return DownloadWithOptions(rawURL, w)
}

// DownloadWithOptions downloads to a Writer with per-request options.
func DownloadWithOptions(rawURL string, w io.Writer, opts ...RequestOption) (int64, error) {
	resp := Get(rawURL, opts...).Execute()
	if resp.err != nil {
		return 0, resp.err
	}
	return resp.writeBodyTo(w)
}

// DownloadBytes downloads and returns bytes.
func DownloadBytes(rawURL string) []byte { return DownloadBytesWithOptions(rawURL) }

// DownloadBytesWithOptions downloads and returns bytes with per-request options.
func DownloadBytesWithOptions(rawURL string, opts ...RequestOption) []byte {
	return Get(rawURL, opts...).Execute().Bytes()
}

// ToParams converts a map to a URL query string.
func ToParams(m map[string]any) string { return urlimpl.EncodeQueryMap(m) }

// EncodeParams encodes a URL containing parameters; only the part after ? is encoded.
func EncodeParams(rawURL string) string { return urlimpl.EncodeParams(rawURL) }

// DecodeParamMap parses a query string into a map.
func DecodeParamMap(paramsStr string) map[string]string { return urlimpl.DecodeQueryFirst(paramsStr) }

// DecodeParams parses a query string into a multi-value map.
func DecodeParams(paramsStr string) map[string][]string { return urlimpl.DecodeQuery(paramsStr) }

// URLWithForm appends form values to a URL.
func URLWithForm(rawURL string, form map[string]any) string { return urlimpl.AppendQuery(rawURL, form) }

// BuildBasicAuth builds a Basic Auth string.
func BuildBasicAuth(user, pass string) string {
	return shared.BuildBasicAuth(user, pass)
}

var (
	// CharsetPattern matches charset in Content-Type.
	CharsetPattern = shared.CharsetPattern
	// MetaCharsetPattern matches charset in HTML meta tags.
	MetaCharsetPattern = shared.MetaCharsetPattern
)

// GetCharsetFromContentType extracts charset from Content-Type.
func GetCharsetFromContentType(ct string) string {
	return shared.GetCharsetFromContentType(ct)
}

// GetCharsetFromContentTypeWithOptions extracts charset from Content-Type with options.
func GetCharsetFromContentTypeWithOptions(ct string, opts ...CharsetOption) string {
	return shared.GetCharsetFromContentTypeWithOptions(ct, opts...)
}

// GetCharsetFromHTML extracts charset from HTML meta tags.
func GetCharsetFromHTML(html string) string {
	return shared.GetCharsetFromHTML(html)
}

// GetCharsetFromHTMLWithOptions extracts charset from HTML meta tags with options.
func GetCharsetFromHTMLWithOptions(html string, opts ...CharsetOption) string {
	return shared.GetCharsetFromHTMLWithOptions(html, opts...)
}

// GetMimeType returns the MIME type by file extension, or an empty string when unknown.
func GetMimeType(filename string) string {
	return shared.GetMimeType(filename)
}
