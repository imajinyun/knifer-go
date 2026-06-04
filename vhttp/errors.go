package vhttp

import httpx "github.com/imajinyun/go-knifer/internal/httpx/http"

// NewError delegates to the internal httpx implementation.
func NewError(msg string, cause error) *Error {
	return httpx.NewHTTPError(msg, cause)
}

// Errorf delegates to the internal httpx implementation.
func Errorf(format string, args ...any) *Error {
	return httpx.HTTPErrorf(format, args...)
}
