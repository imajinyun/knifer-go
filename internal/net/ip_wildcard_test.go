package net

import (
	"regexp"
	"testing"
)

func TestMatchesWildcard(t *testing.T) {
	if !MatchesWildcard("192.168.*.*", "192.168.1.2") {
		t.Fatal("MatchesWildcard failed")
	}
}

func TestMatchesWildcardWithOptions(t *testing.T) {
	var compiled string
	if !MatchesWildcardWithOptions("10.0.*.2", "10.0.1.2", WithWildcardCompileFunc(func(pattern string) (*regexp.Regexp, error) {
		compiled = pattern
		return regexp.Compile(pattern)
	})) {
		t.Fatal("MatchesWildcardWithOptions failed")
	}
	if compiled != `^10\.0\.\d{1,3}\.2$` {
		t.Fatalf("compiled wildcard pattern = %q", compiled)
	}
}
