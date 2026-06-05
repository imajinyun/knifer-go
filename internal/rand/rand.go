// Package rand provides random value helpers.
package rand

import (
	cryptorand "crypto/rand"
	mathrand "math/rand"
	"sync"
	"time"
)

// This file provides random-value helpers aligned with the utility toolkit-core RandomUtil.

// Character set constants used by random string helpers.
const (
	BaseNumber       = "0123456789"
	BaseChar         = "abcdefghijklmnopqrstuvwxyz"
	BaseCharNumber   = BaseChar + BaseNumber
	BaseCharNumberUC = BaseChar + "ABCDEFGHIJKLMNOPQRSTUVWXYZ" + BaseNumber
)

var (
	defaultRandMu sync.Mutex
	defaultRand   = mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
)

// SetSeed resets the package-level pseudo-random source seed.
func SetSeed(seed int64) {
	defaultRandMu.Lock()
	defer defaultRandMu.Unlock()
	defaultRand.Seed(seed)
}

// RandomInt returns a random integer in [0, max). Non-positive max returns 0.
func RandomInt(max int) int {
	if max <= 0 {
		return 0
	}
	defaultRandMu.Lock()
	defer defaultRandMu.Unlock()
	return defaultRand.Intn(max)
}

// RandomIntRange returns a random integer in [min, max). If max <= min, min is returned.
func RandomIntRange(min, max int) int {
	if max <= min {
		return min
	}
	defaultRandMu.Lock()
	defer defaultRandMu.Unlock()
	return min + defaultRand.Intn(max-min)
}

// RandomLong returns a non-negative random int64.
func RandomLong() int64 {
	defaultRandMu.Lock()
	defer defaultRandMu.Unlock()
	return defaultRand.Int63()
}

// RandomFloat returns a random float64 in [0.0, 1.0).
func RandomFloat() float64 {
	defaultRandMu.Lock()
	defer defaultRandMu.Unlock()
	return defaultRand.Float64()
}

// RandomBool returns a random boolean.
func RandomBool() bool {
	defaultRandMu.Lock()
	defer defaultRandMu.Unlock()
	return defaultRand.Intn(2) == 0
}

// RandomBytes returns n cryptographically secure random bytes when possible.
func RandomBytes(n int) []byte {
	if n <= 0 {
		return []byte{}
	}
	buf := make([]byte, n)
	fillRandomBytes(buf)
	return buf
}

func fillRandomBytes(buf []byte) {
	if _, err := cryptorand.Read(buf); err != nil {
		// Fall back to math/rand when crypto/rand is unavailable.
		defaultRandMu.Lock()
		defer defaultRandMu.Unlock()
		for i := range buf {
			buf[i] = byte(defaultRand.Intn(256))
		}
	}
}

// RandomString returns a random string from BaseCharNumber, using lowercase letters and digits.
func RandomString(n int) string { return RandomStringFrom(BaseCharNumber, n) }

// RandomNumbers returns a random numeric string.
func RandomNumbers(n int) string { return RandomStringFrom(BaseNumber, n) }

// RandomStringUpper returns a random string with lowercase letters, uppercase letters, and digits.
func RandomStringUpper(n int) string { return RandomStringFrom(BaseCharNumberUC, n) }

// RandomStringFrom builds a random string by sampling runes from charset.
func RandomStringFrom(charset string, n int) string {
	if n <= 0 || len(charset) == 0 {
		return ""
	}
	rs := []rune(charset)
	out := make([]rune, n)
	defaultRandMu.Lock()
	defer defaultRandMu.Unlock()
	for i := 0; i < n; i++ {
		out[i] = rs[defaultRand.Intn(len(rs))]
	}
	return string(out)
}

// RandomEle returns a random element, or the zero value for an empty slice.
func RandomEle[T any](a []T) T {
	if len(a) == 0 {
		var zero T
		return zero
	}
	defaultRandMu.Lock()
	defer defaultRandMu.Unlock()
	return a[defaultRand.Intn(len(a))]
}
