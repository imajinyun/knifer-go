package vhttp

import httpx "github.com/imajinyun/go-knifer/internal/httpx/http"

// BuildBasicAuth builds a Basic authorization value.
func BuildBasicAuth(user, pass string) string { return httpx.BuildBasicAuth(user, pass) }

// ParseUserAgent parses a User-Agent string.
func ParseUserAgent(ua string) *UserAgent { return httpx.ParseUserAgent(ua) }

// IsRedirected delegates to the internal httpx implementation.
func IsRedirected(status int) bool {
	return httpx.IsRedirected(status)
}
