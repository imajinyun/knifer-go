package vhash

import (
	"hash"

	hashimpl "github.com/imajinyun/go-knifer/internal/hash"
)

// AdditiveHash calculates an additive hash modulo prime.
func AdditiveHash(s string, prime int) int { return hashimpl.AdditiveHash(s, prime) }

// FnvHash calculates a 32-bit FNV-1 hash using the standard library hash/fnv.
func FnvHash(s string) uint32 { return hashimpl.FnvHash(s) }

// Hash32 calculates a 32-bit hash using newHash. nil falls back to FNV-1.
func Hash32(s string, newHash func() hash.Hash32) uint32 { return hashimpl.Hash32(s, newHash) }

// FnvHashString calculates the improved 32-bit FNV-1 hash for strings.
// This differs from FnvHash, which uses the standard library hash/fnv FNV-1.
func FnvHashString(s string) int32 { return hashimpl.FnvHashString(s) }

// RsHash calculates a hash using the RS algorithm.
func RsHash(s string) int32 { return hashimpl.RsHash(s) }

// JsHash calculates a hash using the JS algorithm.
func JsHash(s string) int32 { return hashimpl.JsHash(s) }

// PjwHash calculates a hash using the PJW algorithm.
func PjwHash(s string) int32 { return hashimpl.PjwHash(s) }

// ElfHash calculates a hash using the ELF algorithm.
func ElfHash(s string) int32 { return hashimpl.ElfHash(s) }

// BkdrHash calculates a hash using the BKDR algorithm.
func BkdrHash(s string) int32 { return hashimpl.BkdrHash(s) }

// SdbmHash calculates a hash using the SDBM algorithm.
func SdbmHash(s string) int32 { return hashimpl.SdbmHash(s) }

// DjbHash calculates a hash using the DJB algorithm.
func DjbHash(s string) int32 { return hashimpl.DjbHash(s) }

// ApHash calculates a hash using the AP algorithm.
func ApHash(s string) int32 { return hashimpl.ApHash(s) }

// HfHash calculates a hash using the HF algorithm.
func HfHash(s string) int64 { return hashimpl.HfHash(s) }

// HfIpHash calculates a hash using the HFIP algorithm.
func HfIpHash(s string) int64 { return hashimpl.HfIpHash(s) }

// TianlHash calculates a hash using the TianL algorithm.
func TianlHash(s string) int64 { return hashimpl.TianlHash(s) }

// JavaDefaultHash calculates a hash equivalent to Java String.hashCode.
func JavaDefaultHash(s string) int32 { return hashimpl.JavaDefaultHash(s) }
