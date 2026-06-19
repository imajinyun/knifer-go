package vrand

import (
	"io"
	mathrand "math/rand"

	randimpl "github.com/imajinyun/go-knifer/internal/rand"
)

const (
	BaseNumber       = randimpl.BaseNumber
	BaseChar         = randimpl.BaseChar
	BaseCharNumber   = randimpl.BaseCharNumber
	BaseCharNumberUC = randimpl.BaseCharNumberUC
)

// RandomOption customizes per-call random helpers.
type RandomOption = randimpl.RandomOption

// Int returns a pseudo-random integer in [0, max), or 0 when max is non-positive.
func Int(max int) int { return IntWithOptions(max) }

// IntRange returns a pseudo-random integer in [min, max), or min when max <= min.
func IntRange(min, max int) int { return IntRangeWithOptions(min, max) }

// Long returns a non-negative pseudo-random int64.
func Long() int64 { return LongWithOptions() }

// Float returns a pseudo-random float64 in [0.0, 1.0).
func Float() float64 { return FloatWithOptions() }

// Bool returns a pseudo-random boolean.
func Bool() bool { return BoolWithOptions() }

// String returns a pseudo-random lowercase alphanumeric string of length n.
func String(n int) string { return StringWithOptions(n) }

// Numbers returns a pseudo-random numeric string of length n.
func Numbers(n int) string { return NumbersWithOptions(n) }

// StringUpper returns a pseudo-random mixed-case alphanumeric string of length n.
func StringUpper(n int) string { return StringUpperWithOptions(n) }

// StringFrom builds a pseudo-random string by sampling runes from charset.
func StringFrom(charset string, n int) string { return StringFromWithOptions(charset, n) }

// Ele returns a pseudo-random element from a, or the zero value when a is empty.
func Ele[T any](a []T) T { return EleWithOptions(a) }

// WithRandomSource sets the pseudo-random source used by numeric, string,
// element, and compatibility fallback byte helpers. Use SecureBytes for
// secrets, tokens, keys, and nonces.
func WithRandomSource(source *mathrand.Rand) RandomOption { return randimpl.WithRandomSource(source) }

// WithRandomReader sets the byte source used by BytesWithOptions and SecureBytesWithOptions.
func WithRandomReader(reader io.Reader) RandomOption { return randimpl.WithRandomReader(reader) }

// WithStrictCryptoRandom makes BytesWithOptions return reader errors instead of falling back to pseudo-random bytes.
// Prefer SecureBytes for security-sensitive bytes.
func WithStrictCryptoRandom() RandomOption { return randimpl.WithStrictCryptoRandom() }

func IntWithOptions(max int, opts ...RandomOption) int {
	return randimpl.RandomIntWithOptions(max, opts...)
}

func IntRangeWithOptions(min, max int, opts ...RandomOption) int {
	return randimpl.RandomIntRangeWithOptions(min, max, opts...)
}

func LongWithOptions(opts ...RandomOption) int64 { return randimpl.RandomLongWithOptions(opts...) }

func FloatWithOptions(opts ...RandomOption) float64 { return randimpl.RandomFloatWithOptions(opts...) }

func BoolWithOptions(opts ...RandomOption) bool { return randimpl.RandomBoolWithOptions(opts...) }

func BytesWithOptions(n int, opts ...RandomOption) ([]byte, error) {
	return randimpl.RandomBytesWithOptions(n, opts...)
}

// SecureBytes returns n cryptographically secure random bytes and fails closed on entropy errors.
func SecureBytes(n int) ([]byte, error) { return SecureBytesWithOptions(n) }

// SecureBytesWithOptions returns n cryptographically secure random bytes with per-call options.
func SecureBytesWithOptions(n int, opts ...RandomOption) ([]byte, error) {
	return randimpl.SecureRandomBytesWithOptions(n, opts...)
}

func StringWithOptions(n int, opts ...RandomOption) string {
	return randimpl.RandomStringWithOptions(n, opts...)
}

func NumbersWithOptions(n int, opts ...RandomOption) string {
	return randimpl.RandomNumbersWithOptions(n, opts...)
}

func StringUpperWithOptions(n int, opts ...RandomOption) string {
	return randimpl.RandomStringUpperWithOptions(n, opts...)
}

func StringFromWithOptions(charset string, n int, opts ...RandomOption) string {
	return randimpl.RandomStringFromWithOptions(charset, n, opts...)
}

func EleWithOptions[T any](a []T, opts ...RandomOption) T {
	return randimpl.RandomEleWithOptions(a, opts...)
}

// SetSeed resets the package-level pseudo-random source seed for reproducible
// non-security helpers. It must not be used for secrets, tokens, keys, or nonces.
func SetSeed(seed int64) { randimpl.SetSeed(seed) }

// ConfigureDefaultRandomSourceProvider sets the provider used to lazily create the package-level pseudo-random source.
// It is intended for tests and process bootstrap; tests should call
// ResetDefaultRandomSource from t.Cleanup to avoid cross-test state coupling.
func ConfigureDefaultRandomSourceProvider(provider func() *mathrand.Rand) {
	randimpl.ConfigureDefaultRandomSourceProvider(provider)
}

// ResetDefaultRandomSource restores the time-seeded default source provider and clears cached state.
func ResetDefaultRandomSource() { randimpl.ResetDefaultRandomSource() }
