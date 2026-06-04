package vxml

import xmlimpl "github.com/imajinyun/go-knifer/internal/xml"

// WithNamespaceAware controls whether parsed element names keep namespace URIs.
func WithNamespaceAware(b bool) ParseOption { return xmlimpl.WithNamespaceAware(b) }

// WithCharset sets the XML declaration charset.
func WithCharset(s string) WriteOption { return xmlimpl.WithCharset(s) }

// WithIndent sets the indentation width in spaces (0 disables pretty printing).
func WithIndent(n int) WriteOption { return xmlimpl.WithIndent(n) }

// WithPretty enables pretty printing with the default indentation.
func WithPretty() WriteOption { return xmlimpl.WithPretty() }

// WithOmitDeclaration controls whether the <?xml ... ?> prolog is emitted.
func WithOmitDeclaration(b bool) WriteOption { return xmlimpl.WithOmitDeclaration(b) }

// WithIgnoreNullFields skips struct fields whose value is a typed nil.
func WithIgnoreNullFields(b bool) WriteOption { return xmlimpl.WithIgnoreNullFields(b) }

// WithRootName overrides the synthesized root element name for MarshalMap / MarshalBean.
func WithRootName(s string) WriteOption { return xmlimpl.WithRootName(s) }

// WithNamespace sets the xmlns attribute on the synthesized root element.
func WithNamespace(s string) WriteOption { return xmlimpl.WithNamespace(s) }
