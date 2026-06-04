package vxml

import (
	"io"

	xmlimpl "github.com/imajinyun/go-knifer/internal/xml"
)

// ReadXML parses XML content directly, or treats the input as a file path when
// it does not start with '<'.
func ReadXML(pathOrContent string, opts ...ParseOption) (*Document, error) {
	return xmlimpl.ReadXML(pathOrContent, opts...)
}

// ReadXMLFile parses an XML file.
func ReadXMLFile(path string, opts ...ParseOption) (*Document, error) {
	return xmlimpl.ReadXMLFile(path, opts...)
}

// ReadXMLBytes parses XML bytes.
func ReadXMLBytes(data []byte, opts ...ParseOption) (*Document, error) {
	return xmlimpl.ReadXMLBytes(data, opts...)
}

// ReadXMLReader parses XML from reader.
func ReadXMLReader(r io.Reader, opts ...ParseOption) (*Document, error) {
	return xmlimpl.ReadXMLReader(r, opts...)
}

// ParseXML parses an XML string.
func ParseXML(xmlStr string, opts ...ParseOption) (*Document, error) {
	return xmlimpl.ParseXML(xmlStr, opts...)
}

// ReadBySAX streams XML tokens from reader to handler.
func ReadBySAX(r io.Reader, handler TokenHandler) error { return xmlimpl.ReadBySAX(r, handler) }

// ReadBySAXFile streams XML tokens from file.
func ReadBySAXFile(path string, handler TokenHandler) error {
	return xmlimpl.ReadBySAXFile(path, handler)
}
