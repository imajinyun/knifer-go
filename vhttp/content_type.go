package vhttp

import (
	"regexp"

	httpx "github.com/imajinyun/knifer-go/internal/httpx/http"
)

type CharsetOption = httpx.CharsetOption

// WithCharsetRegexp sets the regexp used by GetCharsetFromContentTypeWithOptions.
func WithCharsetRegexp(re *regexp.Regexp) CharsetOption { return httpx.WithCharsetRegexp(re) }

// WithMetaCharsetRegexp sets the regexp used by GetCharsetFromHTMLWithOptions.
func WithMetaCharsetRegexp(re *regexp.Regexp) CharsetOption { return httpx.WithMetaCharsetRegexp(re) }

// BuildContentType delegates to the internal httpx implementation.
func BuildContentType(contentType, charset string) string {
	return httpx.BuildContentType(contentType, charset)
}

// IsDefaultContentType delegates to the internal httpx implementation.
func IsDefaultContentType(contentType string) bool {
	return httpx.IsDefaultContentType(contentType)
}

// IsFormURLEncoded delegates to the internal httpx implementation.
func IsFormURLEncoded(contentType string) bool {
	return httpx.IsFormURLEncoded(contentType)
}

// GuessContentType delegates to the internal httpx implementation.
func GuessContentType(body string) ContentType {
	return httpx.GuessContentType(body)
}

// GetCharsetFromContentType delegates to the internal httpx implementation.
func GetCharsetFromContentType(ct string) string {
	return httpx.GetCharsetFromContentType(ct)
}

// GetCharsetFromContentTypeWithOptions delegates to the internal httpx implementation.
func GetCharsetFromContentTypeWithOptions(ct string, opts ...CharsetOption) string {
	return httpx.GetCharsetFromContentTypeWithOptions(ct, opts...)
}

// GetCharsetFromHTML delegates to the internal httpx implementation.
func GetCharsetFromHTML(html string) string {
	return httpx.GetCharsetFromHTML(html)
}

// GetCharsetFromHTMLWithOptions delegates to the internal httpx implementation.
func GetCharsetFromHTMLWithOptions(html string, opts ...CharsetOption) string {
	return httpx.GetCharsetFromHTMLWithOptions(html, opts...)
}

// GetMimeType delegates to the internal httpx implementation.
func GetMimeType(filename string) string {
	return httpx.GetMimeType(filename)
}
