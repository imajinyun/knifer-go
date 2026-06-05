package vhttp

import httpx "github.com/imajinyun/go-knifer/internal/httpx/http"

// DownloadString delegates to the internal httpx implementation.
func DownloadString(rawURL, customCharset string) string {
	return httpx.DownloadString(rawURL, customCharset)
}

// DownloadStringWithOptions downloads remote text with per-request options.
func DownloadStringWithOptions(rawURL, customCharset string, opts ...RequestOption) string {
	return httpx.DownloadStringWithOptions(rawURL, customCharset, opts...)
}

// DownloadBytes delegates to the internal httpx implementation.
func DownloadBytes(rawURL string) []byte {
	return httpx.DownloadBytes(rawURL)
}

// DownloadBytesWithOptions downloads and returns bytes with per-request options.
func DownloadBytesWithOptions(rawURL string, opts ...RequestOption) []byte {
	return httpx.DownloadBytesWithOptions(rawURL, opts...)
}
