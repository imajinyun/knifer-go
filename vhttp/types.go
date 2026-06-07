package vhttp

import httpx "github.com/imajinyun/go-knifer/internal/httpx/http"

// Request is a chainable HTTP request builder.
type Request = httpx.HTTPRequest

// RequestOption customizes one HTTP request at construction time.
type RequestOption = httpx.RequestOption

// GlobalConfig is an immutable snapshot of package-level HTTP defaults.
type GlobalConfig = httpx.GlobalConfig

// Response wraps an HTTP response.
type Response = httpx.HTTPResponse

// SaveOption customizes response file saving.
type SaveOption = httpx.SaveOption

// ContentDecoder decodes a response body for a Content-Encoding value.
type ContentDecoder = httpx.ContentDecoder

// NewRequestFunc creates an outgoing HTTP request.
type NewRequestFunc = httpx.NewRequestFunc

// MultipartWriterFactory creates a multipart writer for request bodies.
type MultipartWriterFactory = httpx.MultipartWriterFactory

// MultipartWriter is the subset of multipart.Writer used by request construction.
type MultipartWriter = httpx.MultipartWriter

// ServerOption customizes SimpleServer construction.
type ServerOption = httpx.ServerOption

// StaticOption customizes SimpleServer static file registration.
type StaticOption = httpx.StaticOption

// ListenAndServeFunc starts serving with the provided HTTP server.
type ListenAndServeFunc = httpx.ListenAndServeFunc

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
