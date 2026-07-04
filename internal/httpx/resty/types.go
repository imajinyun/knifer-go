package resty

import (
	knifer "github.com/imajinyun/knifer-go"
	"github.com/imajinyun/knifer-go/internal/httpx/internal/shared"
)

// Method represents an HTTP request method.
type Method = shared.Method

const (
	MethodGet     = shared.MethodGet
	MethodPost    = shared.MethodPost
	MethodHead    = shared.MethodHead
	MethodOptions = shared.MethodOptions
	MethodPut     = shared.MethodPut
	MethodDelete  = shared.MethodDelete
	MethodTrace   = shared.MethodTrace
	MethodConnect = shared.MethodConnect
	MethodPatch   = shared.MethodPatch
)

// Header defines common HTTP header names.
type Header = shared.Header

const (
	HeaderAuthorization      = shared.HeaderAuthorization
	HeaderProxyAuthorization = shared.HeaderProxyAuthorization
	HeaderDate               = shared.HeaderDate
	HeaderConnection         = shared.HeaderConnection
	HeaderMimeVersion        = shared.HeaderMimeVersion
	HeaderTrailer            = shared.HeaderTrailer
	HeaderTransferEncoding   = shared.HeaderTransferEncoding
	HeaderUpgrade            = shared.HeaderUpgrade
	HeaderVia                = shared.HeaderVia
	HeaderCacheControl       = shared.HeaderCacheControl
	HeaderPragma             = shared.HeaderPragma
	HeaderContentType        = shared.HeaderContentType

	HeaderHost           = shared.HeaderHost
	HeaderReferer        = shared.HeaderReferer
	HeaderOrigin         = shared.HeaderOrigin
	HeaderUserAgent      = shared.HeaderUserAgent
	HeaderAccept         = shared.HeaderAccept
	HeaderAcceptLanguage = shared.HeaderAcceptLanguage
	HeaderAcceptEncoding = shared.HeaderAcceptEncoding
	HeaderAcceptCharset  = shared.HeaderAcceptCharset
	HeaderCookie         = shared.HeaderCookie
	HeaderContentLength  = shared.HeaderContentLength

	HeaderWWWAuthenticate    = shared.HeaderWWWAuthenticate
	HeaderSetCookie          = shared.HeaderSetCookie
	HeaderContentEncoding    = shared.HeaderContentEncoding
	HeaderContentDisposition = shared.HeaderContentDisposition
	HeaderETag               = shared.HeaderETag
	HeaderLocation           = shared.HeaderLocation
)

// ContentType defines common Content-Type values.
type ContentType = shared.ContentType

const (
	ContentTypeFormURLEncoded = shared.ContentTypeFormURLEncoded
	ContentTypeMultipart      = shared.ContentTypeMultipart
	ContentTypeJSON           = shared.ContentTypeJSON
	ContentTypeXML            = shared.ContentTypeXML
	ContentTypeTextPlain      = shared.ContentTypeTextPlain
	ContentTypeTextXML        = shared.ContentTypeTextXML
	ContentTypeTextHTML       = shared.ContentTypeTextHTML
	ContentTypeOctetStream    = shared.ContentTypeOctetStream
	ContentTypeEventStream    = shared.ContentTypeEventStream
)

// BuildContentType builds a Content-Type string with charset.
func BuildContentType(contentType, charset string) string {
	return shared.BuildContentType(contentType, charset)
}

// IsDefaultContentType reports whether the value is a default Content-Type.
func IsDefaultContentType(contentType string) bool {
	return shared.IsDefaultContentType(contentType)
}

// IsFormURLEncoded reports whether the value is application/x-www-form-urlencoded.
func IsFormURLEncoded(contentType string) bool {
	return shared.IsFormURLEncoded(contentType)
}

// GuessContentType guesses Content-Type from the first body character.
func GuessContentType(body string) ContentType {
	return shared.GuessContentType(body)
}

// HTTPError represents an error during HTTP operations.
type HTTPError = shared.HTTPError

// NewHTTPError creates an HTTP error.
func NewHTTPError(msg string, cause error) *HTTPError {
	return shared.NewHTTPError(msg, cause)
}

// NewHTTPErrorWithCode creates an HTTP error with an explicit knifer-go code.
func NewHTTPErrorWithCode(code knifer.ErrCode, msg string, cause error) *HTTPError {
	return shared.NewHTTPErrorWithCode(code, msg, cause)
}

// HTTPErrorf creates an HTTP error with a formatted message.
func HTTPErrorf(format string, args ...any) *HTTPError {
	return shared.HTTPErrorf(format, args...)
}

// HTTPErrorfWithCode creates an HTTP error with an explicit code and formatted message.
func HTTPErrorfWithCode(code knifer.ErrCode, format string, args ...any) *HTTPError {
	return shared.HTTPErrorfWithCode(code, format, args...)
}
