package vregex

import (
	"regexp"

	regeximpl "github.com/imajinyun/go-knifer/internal/regex"
)

const (
	// REChinese matches a single Chinese Han character.
	REChinese = regeximpl.REChinese
	// REChineses matches a non-empty string made only of Chinese Han characters.
	REChineses = regeximpl.REChineses
)

// MatchResult describes a single regular-expression match.
type MatchResult = regeximpl.MatchResult

// Match reports whether s contains a match for pattern.
func Match(pattern, s string) bool { return regeximpl.ReMatch(pattern, s) }

// Find returns the first whole-match result.
func Find(pattern, s string) string { return regeximpl.ReFind(pattern, s) }

// FindAll returns all whole-match results.
func FindAll(pattern, s string) []string { return regeximpl.ReFindAll(pattern, s) }

// Replace replaces all matches with replacement.
func Replace(pattern, s, replacement string) string {
	return regeximpl.ReReplace(pattern, s, replacement)
}

// GetGroup0 returns the full text of the first match.
func GetGroup0(pattern, content string) string { return regeximpl.GetGroup0(pattern, content) }

// GetGroup1 returns the first capture group of the first match.
func GetGroup1(pattern, content string) string { return regeximpl.GetGroup1(pattern, content) }

// Get returns a capture group from the first match.
func Get(pattern, content string, groupIndex int) string {
	return regeximpl.Get(pattern, content, groupIndex)
}

// GetOK returns a capture group from the first match and reports whether it exists.
func GetOK(pattern, content string, groupIndex int) (string, bool) {
	return regeximpl.GetOK(pattern, content, groupIndex)
}

// GetByName returns a named capture group from the first match.
func GetByName(pattern, content, groupName string) string {
	return regeximpl.GetByName(pattern, content, groupName)
}

// GetRe returns a capture group from the first match of a compiled expression.
func GetRe(re *regexp.Regexp, content string, groupIndex int) string {
	return regeximpl.GetRe(re, content, groupIndex)
}

// GetByNameRe returns a named capture group from the first match of a compiled expression.
func GetByNameRe(re *regexp.Regexp, content, groupName string) string {
	return regeximpl.GetByNameRe(re, content, groupName)
}

// First calls consumer with the first match of re.
func First(re *regexp.Regexp, content string, consumer func(MatchResult)) {
	regeximpl.First(re, content, consumer)
}

// GetAllGroups returns capture groups from matches.
func GetAllGroups(pattern, content string, withGroup0 bool, findAll bool) []string {
	return regeximpl.GetAllGroups(pattern, content, withGroup0, findAll)
}

// GetAllGroupsRe returns capture groups from matches of a compiled expression.
func GetAllGroupsRe(re *regexp.Regexp, content string, withGroup0 bool, findAll bool) []string {
	return regeximpl.GetAllGroupsRe(re, content, withGroup0, findAll)
}

// GetAllGroupNames returns named capture groups from the first match.
func GetAllGroupNames(pattern, content string) map[string]string {
	return regeximpl.GetAllGroupNames(pattern, content)
}

// GetAllGroupNamesRe returns named capture groups from the first match of a compiled expression.
func GetAllGroupNamesRe(re *regexp.Regexp, content string) map[string]string {
	return regeximpl.GetAllGroupNamesRe(re, content)
}

// ExtractMulti builds a string from the first match using $1, $2, ... placeholders.
func ExtractMulti(pattern, content, template string) string {
	return regeximpl.ExtractMulti(pattern, content, template)
}

// ExtractMultiRe builds a string from the first match of a compiled expression.
func ExtractMultiRe(re *regexp.Regexp, content, template string) string {
	return regeximpl.ExtractMultiRe(re, content, template)
}

// ExtractMultiAndDelPre extracts with a template and removes the consumed prefix from contentHolder.
func ExtractMultiAndDelPre(pattern string, contentHolder *string, template string) string {
	return regeximpl.ExtractMultiAndDelPre(pattern, contentHolder, template)
}

// ExtractMultiAndDelPreRe extracts with a template and removes the consumed prefix from contentHolder.
func ExtractMultiAndDelPreRe(re *regexp.Regexp, contentHolder *string, template string) string {
	return regeximpl.ExtractMultiAndDelPreRe(re, contentHolder, template)
}

// DelFirst deletes the first match.
func DelFirst(pattern, content string) string { return regeximpl.DelFirst(pattern, content) }

// DelFirstRe deletes the first match of a compiled expression.
func DelFirstRe(re *regexp.Regexp, content string) string { return regeximpl.DelFirstRe(re, content) }

// ReplaceFirst replaces the first match.
func ReplaceFirst(pattern, content, replacement string) string {
	return regeximpl.ReplaceFirst(pattern, content, replacement)
}

// ReplaceFirstRe replaces the first match of a compiled expression.
func ReplaceFirstRe(re *regexp.Regexp, content, replacement string) string {
	return regeximpl.ReplaceFirstRe(re, content, replacement)
}

// DelLast deletes the last match.
func DelLast(pattern, content string) string { return regeximpl.DelLast(pattern, content) }

// DelLastRe deletes the last match of a compiled expression.
func DelLastRe(re *regexp.Regexp, content string) string { return regeximpl.DelLastRe(re, content) }

// DelAll deletes every match.
func DelAll(pattern, content string) string { return regeximpl.DelAll(pattern, content) }

// DelAllRe deletes every match of a compiled expression.
func DelAllRe(re *regexp.Regexp, content string) string { return regeximpl.DelAllRe(re, content) }

// DelPre deletes everything through the first match.
func DelPre(pattern, content string) string { return regeximpl.DelPre(pattern, content) }

// DelPreRe deletes everything through the first match of a compiled expression.
func DelPreRe(re *regexp.Regexp, content string) string { return regeximpl.DelPreRe(re, content) }

// FindAllGroup0 returns all full-match strings.
func FindAllGroup0(pattern, content string) []string {
	return regeximpl.FindAllGroup0(pattern, content)
}

// FindAllGroup1 returns all first capture groups.
func FindAllGroup1(pattern, content string) []string {
	return regeximpl.FindAllGroup1(pattern, content)
}

// FindAllGroup returns all values for a capture group.
func FindAllGroup(pattern, content string, group int) []string {
	return regeximpl.FindAll(pattern, content, group)
}

// FindAllGroupRe returns all values for a capture group of a compiled expression.
func FindAllGroupRe(re *regexp.Regexp, content string, group int) []string {
	return regeximpl.FindAllRe(re, content, group)
}

// Each calls consumer for every match.
func Each(re *regexp.Regexp, content string, consumer func(MatchResult)) {
	regeximpl.Each(re, content, consumer)
}

// Count returns the number of matches.
func Count(pattern, content string) int { return regeximpl.Count(pattern, content) }

// CountRe returns the number of matches for a compiled expression.
func CountRe(re *regexp.Regexp, content string) int { return regeximpl.CountRe(re, content) }

// Contains reports whether content contains a match.
func Contains(pattern, content string) bool { return regeximpl.Contains(pattern, content) }

// ContainsRe reports whether content contains a match for a compiled expression.
func ContainsRe(re *regexp.Regexp, content string) bool { return regeximpl.ContainsRe(re, content) }

// IndexOf returns the first match result.
func IndexOf(pattern, content string) *MatchResult { return regeximpl.IndexOf(pattern, content) }

// IndexOfRe returns the first match result for a compiled expression.
func IndexOfRe(re *regexp.Regexp, content string) *MatchResult {
	return regeximpl.IndexOfRe(re, content)
}

// LastIndexOf returns the last match result.
func LastIndexOf(pattern, content string) *MatchResult {
	return regeximpl.LastIndexOf(pattern, content)
}

// LastIndexOfRe returns the last match result for a compiled expression.
func LastIndexOfRe(re *regexp.Regexp, content string) *MatchResult {
	return regeximpl.LastIndexOfRe(re, content)
}

// GetFirstNumber returns the first integer in content.
func GetFirstNumber(content string) (int, bool) { return regeximpl.GetFirstNumber(content) }

// IsMatch reports whether the whole content matches pattern.
func IsMatch(pattern, content string) bool { return regeximpl.IsMatch(pattern, content) }

// IsMatchRe reports whether the whole content matches a compiled expression.
func IsMatchRe(re *regexp.Regexp, content string) bool { return regeximpl.IsMatchRe(re, content) }

// ReplaceAll replaces all matches using a template with $1, $2, ... placeholders.
func ReplaceAll(content, pattern, replacementTemplate string) string {
	return regeximpl.ReplaceAll(content, pattern, replacementTemplate)
}

// ReplaceAllRe replaces all matches of a compiled expression using a template.
func ReplaceAllRe(content string, re *regexp.Regexp, replacementTemplate string) string {
	return regeximpl.ReplaceAllRe(content, re, replacementTemplate)
}

// ReplaceAllFunc replaces all matches using a custom function.
func ReplaceAllFunc(content, pattern string, replaceFunc func(MatchResult) string) string {
	return regeximpl.ReplaceAllFunc(content, pattern, replaceFunc)
}

// ReplaceAllFuncRe replaces all matches of a compiled expression using a custom function.
func ReplaceAllFuncRe(content string, re *regexp.Regexp, replaceFunc func(MatchResult) string) string {
	return regeximpl.ReplaceAllFuncRe(content, re, replaceFunc)
}

// EscapeChar escapes a single regular-expression keyword character.
func EscapeChar(c rune) string { return regeximpl.EscapeChar(c) }

// Escape escapes regular-expression keyword characters in content.
func Escape(content string) string { return regeximpl.Escape(content) }

// TemplateVars returns numeric placeholders referenced by a replacement template, longest first.
func TemplateVars(template string) []int { return regeximpl.TemplateVars(template) }
