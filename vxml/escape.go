package vxml

import xmlimpl "github.com/imajinyun/knifer-go/internal/xml"

// Escape escapes XML text.
func Escape(s string) string { return xmlimpl.Escape(s) }

// Unescape unescapes XML/HTML entities.
func Unescape(s string) string { return xmlimpl.Unescape(s) }
