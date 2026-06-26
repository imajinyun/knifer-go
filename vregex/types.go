package vregex

import regeximpl "github.com/imajinyun/knifer-go/internal/regex"

const (
	// REChinese matches a single Chinese Han character.
	REChinese = regeximpl.REChinese
	// REChineses matches a non-empty string made only of Chinese Han characters.
	REChineses = regeximpl.REChineses
)

// MatchResult describes a single regular-expression match.
type MatchResult = regeximpl.MatchResult
