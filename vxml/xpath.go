package vxml

import xmlimpl "github.com/imajinyun/go-knifer/internal/xml"

// GetElementByXPath returns the first element matched by a simple expression.
func GetElementByXPath(expression string, source any) *Element {
	return xmlimpl.GetElementByXPath(expression, source)
}

// GetNodeListByXPath returns all elements matched by a simple expression.
func GetNodeListByXPath(expression string, source any) []*Element {
	return xmlimpl.GetNodeListByXPath(expression, source)
}

// GetNodeByXPath returns the first node matched by a simple expression.
func GetNodeByXPath(expression string, source any) *Element {
	return xmlimpl.GetNodeByXPath(expression, source)
}

// GetByXPath returns matched text, element, or list based on returnType.
func GetByXPath(expression string, source any, returnType string) any {
	return xmlimpl.GetByXPath(expression, source, returnType)
}
