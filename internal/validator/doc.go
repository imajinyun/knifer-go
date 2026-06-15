// Package validator provides value validation helpers.
//
// It includes common format validators for email, mobile phone, URL,
// IPv4, Chinese characters, and numeric strings. Some validators delegate
// to internal/net, internal/url, and internal/identity for domain-specific
// checks. This package is exposed through the vform facade.
package validator
