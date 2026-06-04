package vregex

import (
	"regexp"

	regeximpl "github.com/imajinyun/go-knifer/internal/regex"
)

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

// TemplateVars returns numeric placeholders referenced by a replacement template, longest first.
func TemplateVars(template string) []int { return regeximpl.TemplateVars(template) }
