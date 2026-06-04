package vhttp

import httpx "github.com/imajinyun/go-knifer/internal/httpx/http"

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

// GetCharsetFromHTML delegates to the internal httpx implementation.
func GetCharsetFromHTML(html string) string {
	return httpx.GetCharsetFromHTML(html)
}

// GetMimeType delegates to the internal httpx implementation.
func GetMimeType(filename string) string {
	return httpx.GetMimeType(filename)
}
