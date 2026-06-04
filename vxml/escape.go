package vxml

import xmlimpl "github.com/imajinyun/go-knifer/internal/xml"

// Escape escapes XML text.
func Escape(s string) string { return xmlimpl.Escape(s) }

// Unescape unescapes XML/HTML entities.
func Unescape(s string) string { return xmlimpl.Unescape(s) }
