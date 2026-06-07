package vregex

import (
	"regexp"

	regeximpl "github.com/imajinyun/go-knifer/internal/regex"
)

// GetGroup0 returns the full text of the first match.
func GetGroup0(pattern, content string) string { return regeximpl.GetGroup0(pattern, content) }

// GetGroup0WithOptions returns the full text of the first match with options.
func GetGroup0WithOptions(pattern, content string, opts ...Option) string {
	return regeximpl.GetGroup0WithOptions(pattern, content, opts...)
}

// GetGroup1 returns the first capture group of the first match.
func GetGroup1(pattern, content string) string { return regeximpl.GetGroup1(pattern, content) }

// GetGroup1WithOptions returns the first capture group of the first match with options.
func GetGroup1WithOptions(pattern, content string, opts ...Option) string {
	return regeximpl.GetGroup1WithOptions(pattern, content, opts...)
}

// Get returns a capture group from the first match.
func Get(pattern, content string, groupIndex int) string {
	return regeximpl.Get(pattern, content, groupIndex)
}

// GetWithOptions returns a capture group from the first match with options.
func GetWithOptions(pattern, content string, groupIndex int, opts ...Option) string {
	return regeximpl.GetWithOptions(pattern, content, groupIndex, opts...)
}

// GetOK returns a capture group from the first match and reports whether it exists.
func GetOK(pattern, content string, groupIndex int) (string, bool) {
	return regeximpl.GetOK(pattern, content, groupIndex)
}

// GetOKWithOptions returns a capture group from the first match with options and reports whether it exists.
func GetOKWithOptions(pattern, content string, groupIndex int, opts ...Option) (string, bool) {
	return regeximpl.GetOKWithOptions(pattern, content, groupIndex, opts...)
}

// GetByName returns a named capture group from the first match.
func GetByName(pattern, content, groupName string) string {
	return regeximpl.GetByName(pattern, content, groupName)
}

// GetByNameWithOptions returns a named capture group from the first match with options.
func GetByNameWithOptions(pattern, content, groupName string, opts ...Option) string {
	return regeximpl.GetByNameWithOptions(pattern, content, groupName, opts...)
}

// GetRe returns a capture group from the first match of a compiled expression.
func GetRe(re *regexp.Regexp, content string, groupIndex int) string {
	return regeximpl.GetRe(re, content, groupIndex)
}

// GetByNameRe returns a named capture group from the first match of a compiled expression.
func GetByNameRe(re *regexp.Regexp, content, groupName string) string {
	return regeximpl.GetByNameRe(re, content, groupName)
}

// GetAllGroups returns capture groups from matches.
func GetAllGroups(pattern, content string, withGroup0 bool, findAll bool) []string {
	return regeximpl.GetAllGroups(pattern, content, withGroup0, findAll)
}

// GetAllGroupsWithOptions returns capture groups from matches with options.
func GetAllGroupsWithOptions(
	pattern string,
	content string,
	withGroup0 bool,
	findAll bool,
	opts ...Option,
) []string {
	return regeximpl.GetAllGroupsWithOptions(pattern, content, withGroup0, findAll, opts...)
}

// GetAllGroupsRe returns capture groups from matches of a compiled expression.
func GetAllGroupsRe(re *regexp.Regexp, content string, withGroup0 bool, findAll bool) []string {
	return regeximpl.GetAllGroupsRe(re, content, withGroup0, findAll)
}

// GetAllGroupNames returns named capture groups from the first match.
func GetAllGroupNames(pattern, content string) map[string]string {
	return regeximpl.GetAllGroupNames(pattern, content)
}

// GetAllGroupNamesWithOptions returns named capture groups from the first match with options.
func GetAllGroupNamesWithOptions(pattern, content string, opts ...Option) map[string]string {
	return regeximpl.GetAllGroupNamesWithOptions(pattern, content, opts...)
}

// GetAllGroupNamesRe returns named capture groups from the first match of a compiled expression.
func GetAllGroupNamesRe(re *regexp.Regexp, content string) map[string]string {
	return regeximpl.GetAllGroupNamesRe(re, content)
}

// FindAllGroup0 returns all full-match strings.
func FindAllGroup0(pattern, content string) []string {
	return regeximpl.FindAllGroup0(pattern, content)
}

// FindAllGroup0WithOptions returns all full-match strings with options.
func FindAllGroup0WithOptions(pattern, content string, opts ...Option) []string {
	return regeximpl.FindAllGroup0WithOptions(pattern, content, opts...)
}

// FindAllGroup1 returns all first capture groups.
func FindAllGroup1(pattern, content string) []string {
	return regeximpl.FindAllGroup1(pattern, content)
}

// FindAllGroup1WithOptions returns all first capture groups with options.
func FindAllGroup1WithOptions(pattern, content string, opts ...Option) []string {
	return regeximpl.FindAllGroup1WithOptions(pattern, content, opts...)
}

// FindAllGroup returns all values for a capture group.
func FindAllGroup(pattern, content string, group int) []string {
	return regeximpl.FindAll(pattern, content, group)
}

// FindAllGroupWithOptions returns all values for a capture group with options.
func FindAllGroupWithOptions(pattern, content string, group int, opts ...Option) []string {
	return regeximpl.FindAllWithOptions(pattern, content, group, opts...)
}

// FindAllGroupRe returns all values for a capture group of a compiled expression.
func FindAllGroupRe(re *regexp.Regexp, content string, group int) []string {
	return regeximpl.FindAllRe(re, content, group)
}
