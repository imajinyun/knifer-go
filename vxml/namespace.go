package vxml

import xmlimpl "github.com/imajinyun/knifer-go/internal/xml"

// NewNamespaceCache collects namespace declarations from doc.
func NewNamespaceCache(doc *Document) *NamespaceCache { return xmlimpl.NewNamespaceCache(doc) }
