package hash

import (
	"hash"
	"hash/fnv"
)

// AdditiveHash calculates an additive hash modulo prime. Non-positive prime falls back to 31.
func AdditiveHash(s string, prime int) int {
	if prime <= 0 {
		prime = 31
	}
	h := len(s)
	for _, r := range s {
		h += int(r)
	}
	return h % prime
}

// FnvHash calculates a 32-bit FNV-1 hash.
func FnvHash(s string) uint32 {
	return Hash32(s, fnv.New32)
}

// Hash32 calculates a 32-bit hash using newHash. nil falls back to FNV-1.
func Hash32(s string, newHash func() hash.Hash32) uint32 {
	if newHash == nil {
		newHash = fnv.New32
	}
	h := newHash()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}
