package vxml

import (
	stdxml "encoding/xml"

	xmlimpl "github.com/imajinyun/go-knifer/internal/xml"
)

// CreateXML creates an empty XML document.
func CreateXML() *Document { return xmlimpl.CreateXML() }

// CreateXMLWithRoot creates an XML document with root element.
func CreateXMLWithRoot(rootElementName string) *Document {
	return xmlimpl.CreateXMLWithRoot(rootElementName)
}

// CreateXMLWithRootNS creates an XML document with root element and namespace URI.
func CreateXMLWithRootNS(rootElementName, namespace string) *Document {
	return xmlimpl.CreateXMLWithRootNS(rootElementName, namespace)
}

// GetRootElement returns the document root element.
func GetRootElement(doc *Document) *Element { return xmlimpl.GetRootElement(doc) }

// GetOwnerDocument returns the document that owns node by walking to the root.
func GetOwnerDocument(node *Element) *Document { return xmlimpl.GetOwnerDocument(node) }

// CleanInvalid removes XML 1.0 invalid control characters.
func CleanInvalid(xmlContent string) string { return xmlimpl.CleanInvalid(xmlContent) }

// CleanComment removes XML comments.
func CleanComment(xmlContent string) string { return xmlimpl.CleanComment(xmlContent) }

// GetElements returns child elements with tag name. Empty tagName returns all direct children.
func GetElements(element *Element, tagName string) []*Element {
	return xmlimpl.GetElements(element, tagName)
}

// GetElement returns the first child element with tag name.
func GetElement(element *Element, tagName string) *Element {
	return xmlimpl.GetElement(element, tagName)
}

// ElementText returns child text or defaultValue when missing.
func ElementText(element *Element, tagName string, defaultValue ...string) string {
	return xmlimpl.ElementText(element, tagName, defaultValue...)
}

// TransElements returns the input list without nil elements.
func TransElements(nodes []*Element) []*Element { return xmlimpl.TransElements(nodes) }

// IsElement reports whether node is not nil.
func IsElement(node *Element) bool { return xmlimpl.IsElement(node) }

// AppendChild appends and returns a child element.
func AppendChild(node *Element, tagName string, namespace ...string) *Element {
	return xmlimpl.AppendChild(node, tagName, namespace...)
}

// AppendText appends text to an element.
func AppendText(node *Element, text any) *Element { return xmlimpl.AppendText(node, text) }

// Append appends map, slice, struct, or scalar data to node.
func Append(node *Element, data any) { xmlimpl.Append(node, data) }

// XMLName builds a stdxml.Name from local name.
func XMLName(local string) stdxml.Name { return stdxml.Name{Local: local} }
