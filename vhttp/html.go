package vhttp

import httpx "github.com/imajinyun/go-knifer/internal/httpx/http"

// HTMLEscape delegates to the internal httpx implementation.
func HTMLEscape(s string) string {
	return httpx.HTMLEscape(s)
}

// HTMLUnescape delegates to the internal httpx implementation.
func HTMLUnescape(s string) string {
	return httpx.HTMLUnescape(s)
}

// CleanHTML delegates to the internal httpx implementation.
func CleanHTML(s string) string {
	return httpx.CleanHTML(s)
}

// FilterHTMLTag delegates to the internal httpx implementation.
func FilterHTMLTag(s string, tagNames ...string) string {
	return httpx.FilterHTMLTag(s, tagNames...)
}
