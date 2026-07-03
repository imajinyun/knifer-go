// Package vhttp provides public APIs for HTTP utilities.
//
// Prefer construction-time RequestOption values such as WithTimeout,
// WithHeader, WithFollowRedirects, WithCookieJar, and WithUserAgent for
// request-specific behavior. Global defaults are still available for process
// wide compatibility, but per-call options keep each request explicit.
//
// This package only acts as a facade. Concrete implementations live in the
// corresponding internal subpackage.
//
// Start here:
//   - Quickstart: https://github.com/imajinyun/knifer-go/blob/main/docs/doc/22-vhttp.md
//   - Safe HTTP cookbook: https://github.com/imajinyun/knifer-go/blob/main/docs/doc/safe-http-cookbook.md
package vhttp
