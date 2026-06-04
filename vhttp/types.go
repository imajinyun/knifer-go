package vhttp

import httpx "github.com/imajinyun/go-knifer/internal/httpx/http"

// Request is a chainable HTTP request builder.
type Request = httpx.HTTPRequest

// RequestOption customizes one HTTP request at construction time.
type RequestOption = httpx.RequestOption

// Response wraps an HTTP response.
type Response = httpx.HTTPResponse

// SaveOption customizes response file saving.
type SaveOption = httpx.SaveOption

// Method represents an HTTP method.
type Method = httpx.Method

// Header represents an HTTP header name.
type Header = httpx.Header

// ContentType represents an HTTP content type.
type ContentType = httpx.ContentType

// Error is the HTTP module error type.
type Error = httpx.HTTPError

// SimpleServer is a small HTTP server helper.
type SimpleServer = httpx.SimpleServer

// UserAgent describes parsed User-Agent information.
type UserAgent = httpx.UserAgent

const (
	// MethodGet is GET.
	MethodGet Method = httpx.MethodGet
	// MethodPost is POST.
	MethodPost Method = httpx.MethodPost
	// MethodPut is PUT.
	MethodPut Method = httpx.MethodPut
	// MethodDelete is DELETE.
	MethodDelete Method = httpx.MethodDelete
	// MethodPatch is PATCH.
	MethodPatch Method = httpx.MethodPatch
	// MethodHead is HEAD.
	MethodHead Method = httpx.MethodHead
	// MethodOptions is OPTIONS.
	MethodOptions Method = httpx.MethodOptions
)
