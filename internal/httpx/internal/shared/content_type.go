package shared

import (
	"fmt"
	"strings"
)

// ContentType defines common Content-Type values.
type ContentType string

const (
	ContentTypeFormURLEncoded ContentType = "application/x-www-form-urlencoded"
	ContentTypeMultipart      ContentType = "multipart/form-data"
	ContentTypeJSON           ContentType = "application/json"
	ContentTypeXML            ContentType = "application/xml"
	ContentTypeTextPlain      ContentType = "text/plain"
	ContentTypeTextXML        ContentType = "text/xml"
	ContentTypeTextHTML       ContentType = "text/html"
	ContentTypeOctetStream    ContentType = "application/octet-stream"
	ContentTypeEventStream    ContentType = "text/event-stream"
)

// String returns the Content-Type string.
func (c ContentType) String() string { return string(c) }

// WithCharset returns the Content-Type string with charset, such as "application/json;charset=UTF-8".
func (c ContentType) WithCharset(charset string) string {
	return BuildContentType(string(c), charset)
}

// BuildContentType builds a Content-Type string with charset.
func BuildContentType(contentType, charset string) string {
	if charset == "" {
		return contentType
	}
	return fmt.Sprintf("%s;charset=%s", contentType, charset)
}

// IsDefaultContentType reports whether the value is a default Content-Type, including "" and form-urlencoded.
func IsDefaultContentType(contentType string) bool {
	return contentType == "" || IsFormURLEncoded(contentType)
}

// IsFormURLEncoded reports whether the value is application/x-www-form-urlencoded.
func IsFormURLEncoded(contentType string) bool {
	return strings.HasPrefix(strings.ToLower(contentType), string(ContentTypeFormURLEncoded))
}

// GuessContentType guesses Content-Type from the first body character; only JSON/XML are supported.
func GuessContentType(body string) ContentType {
	body = strings.TrimSpace(body)
	if body == "" {
		return ""
	}
	switch body[0] {
	case '{', '[':
		return ContentTypeJSON
	case '<':
		return ContentTypeXML
	}
	return ""
}
