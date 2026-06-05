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

func Int(max int) int                         { return randimpl.RandomInt(max) }
func IntRange(min, max int) int               { return randimpl.RandomIntRange(min, max) }
func Long() int64                             { return randimpl.RandomLong() }
func Float() float64                          { return randimpl.RandomFloat() }
func Bool() bool                              { return randimpl.RandomBool() }
func Bytes(n int) []byte                      { return randimpl.RandomBytes(n) }
func String(n int) string                     { return randimpl.RandomString(n) }
func Numbers(n int) string                    { return randimpl.RandomNumbers(n) }
func StringUpper(n int) string                { return randimpl.RandomStringUpper(n) }
func StringFrom(charset string, n int) string { return randimpl.RandomStringFrom(charset, n) }
func Ele[T any](a []T) T                      { return randimpl.RandomEle(a) }

// WithRandomSource sets the pseudo-random source used by numeric, string, element, and fallback byte helpers.
func WithRandomSource(source *mathrand.Rand) RandomOption { return randimpl.WithRandomSource(source) }

// WithRandomReader sets the byte source used by BytesWithOptions.
func WithRandomReader(reader io.Reader) RandomOption { return randimpl.WithRandomReader(reader) }

// WithStrictCryptoRandom makes BytesWithOptions return reader errors instead of falling back to pseudo-random bytes.
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

// SetSeed resets the package-level pseudo-random source seed.
func SetSeed(seed int64) { randimpl.SetSeed(seed) }
