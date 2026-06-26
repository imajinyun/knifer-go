package vxml

import (
	"io"

	xmlimpl "github.com/imajinyun/knifer-go/internal/xml"
)

// WriteTo serializes a document or element to writer.
func WriteTo(w io.Writer, v any, opts ...WriteOption) error {
	return xmlimpl.WriteTo(w, v, opts...)
}

// MarshalString serializes a document or element to string.
func MarshalString(v any, opts ...WriteOption) (string, error) {
	return xmlimpl.MarshalString(v, opts...)
}

// WriteFile writes a document or element to path.
func WriteFile(path string, v any, opts ...WriteOption) error {
	return xmlimpl.WriteFile(path, v, opts...)
}

// MarshalMap serializes map data to an XML string.
func MarshalMap(data map[string]any, opts ...WriteOption) (string, error) {
	return xmlimpl.MarshalMap(data, opts...)
}

// MarshalBean serializes a struct or map-like value to an XML string.
func MarshalBean(bean any, opts ...WriteOption) (string, error) {
	return xmlimpl.MarshalBean(bean, opts...)
}

// TransformWith copies XML from source to result with per-call options.
func TransformWith(source io.Reader, result io.Writer, opts ...WriteOption) error {
	return xmlimpl.TransformWith(source, result, opts...)
}

// TransformWithOptions copies XML from source to result with parser and writer options.
func TransformWithOptions(source io.Reader, result io.Writer, opts ...TransformOption) error {
	return xmlimpl.TransformWithOptions(source, result, opts...)
}

// Format pretty prints XML content.
func Format(xmlStr string) (string, error) { return xmlimpl.Format(xmlStr) }

// FormatWithOptions pretty prints XML content with parser and writer options.
func FormatWithOptions(xmlStr string, opts ...FormatOption) (string, error) {
	return xmlimpl.FormatWithOptions(xmlStr, opts...)
}
