package vregex

import regeximpl "github.com/imajinyun/go-knifer/internal/regex"

// EscapeChar escapes a single regular-expression keyword character.
func EscapeChar(c rune) string { return regeximpl.EscapeChar(c) }

// Escape escapes regular-expression keyword characters in content.
func Escape(content string) string { return regeximpl.Escape(content) }
