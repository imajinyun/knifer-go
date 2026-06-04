// Package shared holds engine-agnostic HTTP protocol types and helpers
// (methods, headers, content types, and errors).
//
// It is scoped to the internal/httpx subtree, so only the http and resty
// implementation packages may import it. Other modules and sibling internal
// packages cannot reference it.
package shared
