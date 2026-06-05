package vrand

import randimpl "github.com/imajinyun/go-knifer/internal/rand"

const (
	BaseNumber       = randimpl.BaseNumber
	BaseChar         = randimpl.BaseChar
	BaseCharNumber   = randimpl.BaseCharNumber
	BaseCharNumberUC = randimpl.BaseCharNumberUC
)

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

// SetSeed resets the package-level pseudo-random source seed.
func SetSeed(seed int64) { randimpl.SetSeed(seed) }
