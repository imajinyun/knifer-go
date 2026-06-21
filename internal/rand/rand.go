// Package rand provides random value helpers.
package rand

import (
	cryptorand "crypto/rand"
	"io"
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
	defaultRandMu       sync.Mutex
	defaultRand         *mathrand.Rand
	defaultRandProvider = newDefaultRand
)

type randomConfig struct {
	source       *mathrand.Rand
	reader       io.Reader
	strictCrypto bool
}

// RandomOption customizes per-call random helpers.
type RandomOption func(*randomConfig)

// WithRandomSource sets the pseudo-random source used by numeric, string,
// element, and compatibility fallback byte helpers. Do not use pseudo-random
// sources for secrets, tokens, keys, or nonces; use SecureRandomBytes with a
// cryptographic reader for security-sensitive bytes.
func WithRandomSource(source *mathrand.Rand) RandomOption {
	return func(c *randomConfig) { c.source = source }
}

// WithRandomReader sets the byte source used by RandomBytesWithOptions and SecureRandomBytesWithOptions.
func WithRandomReader(reader io.Reader) RandomOption {
	return func(c *randomConfig) { c.reader = reader }
}

// WithStrictCryptoRandom makes RandomBytesWithOptions return reader errors instead of falling back to pseudo-random bytes.
// Prefer SecureRandomBytes for security-sensitive bytes.
func WithStrictCryptoRandom() RandomOption {
	return func(c *randomConfig) { c.strictCrypto = true }
}

func applyRandomOptions(opts []RandomOption) randomConfig {
	cfg := randomConfig{reader: cryptorand.Reader}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.reader == nil {
		cfg.reader = cryptorand.Reader
	}
	return cfg
}

// SetSeed resets the package-level pseudo-random source seed. It is intended
// for reproducible non-security helpers such as RandomInt, RandomString, and
// RandomEle; it must not be used for secrets, tokens, keys, or nonces.
func SetSeed(seed int64) {
	defaultRandMu.Lock()
	defer defaultRandMu.Unlock()
	defaultRand = mathrand.New(mathrand.NewSource(seed))
}

// ConfigureDefaultRandomSourceProvider sets the provider used to lazily create
// the package-level pseudo-random source. The source is for reproducible or
// convenience randomness only; SecureRandomBytes does not rely on it unless a
// caller explicitly supplies an insecure reader and disables strict behavior.
// Passing nil restores the time-seeded default provider and clears the cached source.
func ConfigureDefaultRandomSourceProvider(provider func() *mathrand.Rand) {
	defaultRandMu.Lock()
	defer defaultRandMu.Unlock()
	defaultRand = nil
	if provider == nil {
		defaultRandProvider = newDefaultRand
		return
	}
	defaultRandProvider = provider
}

// ResetDefaultRandomSource restores the time-seeded default source provider and clears cached state.
func ResetDefaultRandomSource() { ConfigureDefaultRandomSourceProvider(nil) }

// RandomInt returns a random integer in [0, max). Non-positive max returns 0.
func RandomInt(max int) int {
	return RandomIntWithOptions(max)
}

// RandomIntWithOptions returns a random integer in [0, max) with per-call options.
func RandomIntWithOptions(max int, opts ...RandomOption) int {
	if max <= 0 {
		return 0
	}
	cfg := applyRandomOptions(opts)
	return randomIntn(cfg, max)
}

// RandomIntRange returns a random integer in [min, max). If max <= min, min is returned.
func RandomIntRange(min, max int) int {
	return RandomIntRangeWithOptions(min, max)
}

// RandomIntRangeWithOptions returns a random integer in [min, max) with per-call options.
func RandomIntRangeWithOptions(min, max int, opts ...RandomOption) int {
	if max <= min {
		return min
	}
	cfg := applyRandomOptions(opts)
	return min + randomIntn(cfg, max-min)
}

// RandomLong returns a non-negative random int64.
func RandomLong() int64 {
	return RandomLongWithOptions()
}

// RandomLongWithOptions returns a non-negative random int64 with per-call options.
func RandomLongWithOptions(opts ...RandomOption) int64 {
	cfg := applyRandomOptions(opts)
	if cfg.source != nil {
		return cfg.source.Int63()
	}
	defaultRandMu.Lock()
	defer defaultRandMu.Unlock()
	return defaultRandLocked().Int63()
}

// RandomFloat returns a random float64 in [0.0, 1.0).
func RandomFloat() float64 {
	return RandomFloatWithOptions()
}

// RandomFloatWithOptions returns a random float64 in [0.0, 1.0) with per-call options.
func RandomFloatWithOptions(opts ...RandomOption) float64 {
	cfg := applyRandomOptions(opts)
	if cfg.source != nil {
		return cfg.source.Float64()
	}
	defaultRandMu.Lock()
	defer defaultRandMu.Unlock()
	return defaultRandLocked().Float64()
}

// RandomBool returns a random boolean.
func RandomBool() bool {
	return RandomBoolWithOptions()
}

// RandomBoolWithOptions returns a random boolean with per-call options.
func RandomBoolWithOptions(opts ...RandomOption) bool {
	cfg := applyRandomOptions(opts)
	return randomIntn(cfg, 2) == 0
}

// RandomBytes returns n random bytes.
//
// Compatibility: when the cryptographic reader fails, RandomBytesWithOptions can
// fall back to pseudo-random bytes unless WithStrictCryptoRandom is provided.
// Do not use this helper for secrets, tokens, keys, or nonces; use
// SecureRandomBytes or package vcrypto instead.
func RandomBytes(n int) []byte {
	b, _ := RandomBytesWithOptions(n)
	return b
}

// RandomBytesWithOptions returns n random bytes with per-call options.
//
// Compatibility: unless WithStrictCryptoRandom is provided, reader failures fall
// back to pseudo-random bytes. Use SecureRandomBytesWithOptions for
// security-sensitive bytes.
func RandomBytesWithOptions(n int, opts ...RandomOption) ([]byte, error) {
	if n <= 0 {
		return []byte{}, nil
	}
	cfg := applyRandomOptions(opts)
	buf := make([]byte, n)
	err := fillRandomBytesWithConfig(buf, cfg)
	if err != nil {
		return nil, err
	}
	return buf, err
}

// SecureRandomBytes returns n cryptographically secure random bytes and fails closed on entropy errors.
func SecureRandomBytes(n int) ([]byte, error) { return SecureRandomBytesWithOptions(n) }

// SecureRandomBytesWithOptions returns n cryptographically secure random bytes with per-call options.
func SecureRandomBytesWithOptions(n int, opts ...RandomOption) ([]byte, error) {
	strict := make([]RandomOption, 0, len(opts)+1)
	strict = append(strict, opts...)
	strict = append(strict, WithStrictCryptoRandom())
	return RandomBytesWithOptions(n, strict...)
}

func fillRandomBytes(buf []byte) {
	_ = fillRandomBytesWithConfig(buf, applyRandomOptions(nil))
}

func fillRandomBytesWithConfig(buf []byte, cfg randomConfig) error {
	if _, err := io.ReadFull(cfg.reader, buf); err != nil {
		if cfg.strictCrypto {
			return err
		}
		// Compatibility fallback: RandomBytes historically returned bytes even when
		// crypto/rand failed. This path is intentionally pseudo-random and must not
		// be used for secrets, tokens, keys, or nonces; SecureRandomBytes enables
		// strictCrypto and fails closed instead.
		for i := range buf {
			buf[i] = byte(randomIntn(cfg, 256))
		}
	}
	return nil
}

// RandomString returns a random string from BaseCharNumber, using lowercase letters and digits.
func RandomString(n int) string { return RandomStringWithOptions(n) }

// RandomStringWithOptions returns a random string from BaseCharNumber with per-call options.
func RandomStringWithOptions(n int, opts ...RandomOption) string {
	return RandomStringFromWithOptions(BaseCharNumber, n, opts...)
}

// RandomNumbers returns a random numeric string.
func RandomNumbers(n int) string { return RandomNumbersWithOptions(n) }

// RandomNumbersWithOptions returns a random numeric string with per-call options.
func RandomNumbersWithOptions(n int, opts ...RandomOption) string {
	return RandomStringFromWithOptions(BaseNumber, n, opts...)
}

// RandomStringUpper returns a random string with lowercase letters, uppercase letters, and digits.
func RandomStringUpper(n int) string { return RandomStringUpperWithOptions(n) }

// RandomStringUpperWithOptions returns a random mixed-case alphanumeric string with per-call options.
func RandomStringUpperWithOptions(n int, opts ...RandomOption) string {
	return RandomStringFromWithOptions(BaseCharNumberUC, n, opts...)
}

// RandomStringFrom builds a random string by sampling runes from charset.
func RandomStringFrom(charset string, n int) string {
	return RandomStringFromWithOptions(charset, n)
}

// RandomStringFromWithOptions builds a random string by sampling runes from charset with per-call options.
func RandomStringFromWithOptions(charset string, n int, opts ...RandomOption) string {
	if n <= 0 || len(charset) == 0 {
		return ""
	}
	cfg := applyRandomOptions(opts)
	rs := []rune(charset)
	out := make([]rune, n)
	for i := 0; i < n; i++ {
		out[i] = rs[randomIntn(cfg, len(rs))]
	}
	return string(out)
}

// RandomEle returns a random element, or the zero value for an empty slice.
func RandomEle[T any](a []T) T {
	return RandomEleWithOptions(a)
}

// RandomEleWithOptions returns a random element with per-call options, or the zero value for an empty slice.
func RandomEleWithOptions[T any](a []T, opts ...RandomOption) T {
	if len(a) == 0 {
		var zero T
		return zero
	}
	cfg := applyRandomOptions(opts)
	return a[randomIntn(cfg, len(a))]
}

func randomIntn(cfg randomConfig, n int) int {
	if cfg.source != nil {
		return cfg.source.Intn(n)
	}
	defaultRandMu.Lock()
	defer defaultRandMu.Unlock()
	return defaultRandLocked().Intn(n)
}

func defaultRandLocked() *mathrand.Rand {
	if defaultRand == nil {
		defaultRand = defaultRandProvider()
		if defaultRand == nil {
			defaultRand = newDefaultRand()
		}
	}
	return defaultRand
}

func newDefaultRand() *mathrand.Rand {
	return mathrand.New(mathrand.NewSource(time.Now().UnixNano())) // #nosec G404 -- non-security helpers and RandomBytes compatibility fallback only; SecureRandomBytes fails closed.
}
