package vregex

import (
	"regexp"
	"strings"
	"testing"
)

func TestRegexMatchFacade(t *testing.T) {
	if !Match(`^\d+$`, "123") || Match(`^\d+$`, "12a") || Match(`(`, "x") {
		t.Fatal("Match failed")
	}
	if !Contains(`\d+`, "a1") || Contains(`\d+`, "abc") {
		t.Fatal("Contains failed")
	}
	if !ContainsRe(regexp.MustCompile(`\d+`), "a1") || ContainsRe(regexp.MustCompile(`\d+`), "abc") {
		t.Fatal("ContainsRe failed")
	}
	if !IsMatch(`\d+`, "123") || IsMatch(`\d+`, "a123") {
		t.Fatal("IsMatch failed")
	}
	if !IsMatchRe(regexp.MustCompile(`\d+`), "123") || IsMatchRe(regexp.MustCompile(`\d+`), "a123") {
		t.Fatal("IsMatchRe failed")
	}
}

func TestRegexMatchFacadeWithOptions(t *testing.T) {
	opt := WithCompileFunc(func(pattern string) (*regexp.Regexp, error) {
		return regexp.Compile(strings.ReplaceAll(pattern, "TOKEN", `\d+`))
	})

	if !MatchWithOptions(`TOKEN`, "abc123", opt) {
		t.Fatal("MatchWithOptions failed")
	}
	if !ContainsWithOptions(`TOKEN`, "abc123", opt) {
		t.Fatal("ContainsWithOptions failed")
	}
	if !IsMatchWithOptions(`TOKEN`, "123", opt) || IsMatchWithOptions(`TOKEN`, "abc123", opt) {
		t.Fatal("IsMatchWithOptions failed")
	}
}

func TestFacadeRegexOptionSetters(t *testing.T) {
	if WithDotAll(true) == nil {
		t.Fatal("WithDotAll returned nil")
	}
	if WithNamedGroupRegexp(regexp.MustCompile(`\(`)) == nil {
		t.Fatal("WithNamedGroupRegexp returned nil")
	}
	if WithNamedGroupNormalizer(strings.ToUpper) == nil {
		t.Fatal("WithNamedGroupNormalizer returned nil")
	}
}
