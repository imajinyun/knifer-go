package vregex

import (
	"regexp"

	regeximpl "github.com/imajinyun/go-knifer/internal/regex"
)

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
