package vhash

import hashimpl "github.com/imajinyun/go-knifer/internal/hash"

func AdditiveHash(s string, prime int) int { return hashimpl.AdditiveHash(s, prime) }
func FnvHash(s string) uint32              { return hashimpl.FnvHash(s) }

// MD5Hex returns the MD5 hex digest of s.
// For security-sensitive digest workflows, prefer vcrypto.MD5Hex.
func MD5Hex(s string) string { return hashimpl.MD5Hex(s) }

// SHA1Hex returns the SHA-1 hex digest of s.
// For security-sensitive digest workflows, prefer vcrypto.SHA1Hex.
func SHA1Hex(s string) string { return hashimpl.SHA1Hex(s) }

// SHA256Hex returns the SHA-256 hex digest of s.
// For security-sensitive digest workflows, prefer vcrypto.SHA256Hex.
func SHA256Hex(s string) string { return hashimpl.SHA256Hex(s) }
