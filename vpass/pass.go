package vpass

import passimpl "github.com/imajinyun/knifer-go/internal/pass"

type (
	// Strength classifies a password score into broad buckets.
	Strength = passimpl.Strength
	// Analysis describes the rule-level result of a password strength check.
	Analysis = passimpl.Analysis
)

const (
	StrengthUnknown    = passimpl.StrengthUnknown
	StrengthVeryWeak   = passimpl.StrengthVeryWeak
	StrengthWeak       = passimpl.StrengthWeak
	StrengthMedium     = passimpl.StrengthMedium
	StrengthStrong     = passimpl.StrengthStrong
	StrengthVeryStrong = passimpl.StrengthVeryStrong
)

func Analyze(password string) Analysis    { return passimpl.Analyze(password) }
func Score(password string) int           { return passimpl.Score(password) }
func StrengthOf(password string) Strength { return passimpl.StrengthOf(password) }
func IsStrong(password string) bool       { return passimpl.IsStrong(password) }
func IsWeak(password string) bool         { return passimpl.IsWeak(password) }
