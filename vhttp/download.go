package vhttp

import httpx "github.com/imajinyun/go-knifer/internal/httpx/http"

// DownloadString delegates to the internal httpx implementation.
func DownloadString(rawURL, customCharset string) string {
	return httpx.DownloadString(rawURL, customCharset)
}

// DownloadBytes delegates to the internal httpx implementation.
func DownloadBytes(rawURL string) []byte {
	return httpx.DownloadBytes(rawURL)
}
