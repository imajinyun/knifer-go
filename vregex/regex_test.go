package vregex

import (
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func TestRegexFacade(t *testing.T) {
	if !Match(`^\d+$`, "123") || Match(`^\d+$`, "12a") || Match(`(`, "x") {
		t.Fatal("Match failed")
	}
	if Find(`\d+`, "ab123cd") != "123" || Find(`(`, "x") != "" {
		t.Fatal("Find failed")
	}
	if all := FindAll(`\d+`, "a1b22c333"); len(all) != 3 || all[2] != "333" {
		t.Fatalf("FindAll failed")
	}
	if Replace(`\d`, "a1b2", "*") != "a*b*" || Replace(`(`, "x", "*") != "x" {
		t.Fatal("Replace failed")
	}
}

func TestExtendedRegexFacade(t *testing.T) {
	if got := GetGroup1(`(\d+)`, "abc123"); got != "123" {
		t.Fatalf("GetGroup1 = %q", got)
	}
	if got := GetByName(`(?<word>\w+)-(?<num>\d+)`, "abc-123", "num"); got != "123" {
		t.Fatalf("GetByName = %q", got)
	}
	if got := GetAllGroups(`(a)(b)`, "ab", true, false); !reflect.DeepEqual(got, []string{"ab", "a", "b"}) {
		t.Fatalf("GetAllGroups = %#v", got)
	}
	if got := ExtractMulti(`(\d+)年(\d+)月`, "2026年5月", `$1-$2`); got != "2026-5" {
		t.Fatalf("ExtractMulti = %q", got)
	}
	if got := DelLast(`\d+`, "a1b22"); got != "a1b" {
		t.Fatalf("DelLast = %q", got)
	}
	if got := FindAllGroup1(`(\d+)`, "a1b22"); !reflect.DeepEqual(got, []string{"1", "22"}) {
		t.Fatalf("FindAllGroup1 = %#v", got)
	}
	if Count(`\d+`, "a1b22") != 2 || !Contains(`\d+`, "a1") || Contains(`\d+`, "abc") {
		t.Fatal("Count/Contains failed")
	}
	if match := IndexOf(`\d+`, "ab12"); match == nil || match.Start != 2 || match.Text != "12" {
		t.Fatalf("IndexOf = %#v", match)
	}
	if n, ok := GetFirstNumber("a123"); !ok || n != 123 {
		t.Fatalf("GetFirstNumber = %d %v", n, ok)
	}
	if !IsMatch(`\d+`, "123") || IsMatch(`\d+`, "a123") {
		t.Fatal("IsMatch failed")
	}
	if got := ReplaceAll("中文1234", `(\d+)`, `($1)`); got != "中文(1234)" {
		t.Fatalf("ReplaceAll = %q", got)
	}
	if got := ReplaceAllFunc("a1b22", `\d+`, func(m MatchResult) string { return "[" + m.Text + "]" }); got != "a[1]b[22]" {
		t.Fatalf("ReplaceAllFunc = %q", got)
	}
	if got := Escape("a+b(c)"); got != `a\+b\(c\)` {
		t.Fatalf("Escape = %q", got)
	}
}

func TestRegexFacadeWithOptions(t *testing.T) {
	compiler := func(pattern string) (*regexp.Regexp, error) {
		return regexp.Compile(strings.ReplaceAll(pattern, "TOKEN", `\d+`))
	}
	opt := WithCompileFunc(compiler)

	if got := FindWithOptions(`TOKEN`, "ab123", opt); got != "123" {
		t.Fatalf("FindWithOptions = %q", got)
	}
	if got := FindAllWithOptions(`TOKEN`, "a1b22", opt); !reflect.DeepEqual(got, []string{"1", "22"}) {
		t.Fatalf("FindAllWithOptions = %#v", got)
	}
	if got := GetGroup0WithOptions(`TOKEN`, "a123", opt); got != "123" {
		t.Fatalf("GetGroup0WithOptions = %q", got)
	}
	if got := GetGroup1WithOptions(`x(TOKEN)`, "x123", opt); got != "123" {
		t.Fatalf("GetGroup1WithOptions = %q", got)
	}
	if got := GetWithOptions(`x(TOKEN)`, "x123", 1, opt); got != "123" {
		t.Fatalf("GetWithOptions = %q", got)
	}
	if got, ok := GetOKWithOptions(`x(TOKEN)`, "x123", 1, opt); !ok || got != "123" {
		t.Fatalf("GetOKWithOptions = %q %v", got, ok)
	}
	if got := GetByNameWithOptions(`x(?<num>TOKEN)`, "x123", "num", opt); got != "123" {
		t.Fatalf("GetByNameWithOptions = %q", got)
	}
	if got := GetAllGroupsWithOptions(`x(TOKEN)`, "x123", true, false, opt); !reflect.DeepEqual(got, []string{"x123", "123"}) {
		t.Fatalf("GetAllGroupsWithOptions = %#v", got)
	}
	if got := GetAllGroupNamesWithOptions(`x(?<num>TOKEN)`, "x123", opt); got["num"] != "123" {
		t.Fatalf("GetAllGroupNamesWithOptions = %#v", got)
	}
	if got := ExtractMultiWithOptions(`x(TOKEN)`, "x123", "$1", opt); got != "123" {
		t.Fatalf("ExtractMultiWithOptions = %q", got)
	}
	holder := "x123y"
	if got := ExtractMultiAndDelPreWithOptions(`x(TOKEN)`, &holder, "$1", opt); got != "123" || holder != "y" {
		t.Fatalf("ExtractMultiAndDelPreWithOptions = %q holder=%q", got, holder)
	}
	if got := DelFirstWithOptions(`TOKEN`, "a123b456", opt); got != "ab456" {
		t.Fatalf("DelFirstWithOptions = %q", got)
	}
	if got := ReplaceFirstWithOptions(`TOKEN`, "a123b456", "X", opt); got != "aXb456" {
		t.Fatalf("ReplaceFirstWithOptions = %q", got)
	}
	if got := DelLastWithOptions(`TOKEN`, "a123b456", opt); got != "a123b" {
		t.Fatalf("DelLastWithOptions = %q", got)
	}
	if got := DelAllWithOptions(`TOKEN`, "a123b456", opt); got != "ab" {
		t.Fatalf("DelAllWithOptions = %q", got)
	}
	if got := DelPreWithOptions(`TOKEN`, "a123b", opt); got != "b" {
		t.Fatalf("DelPreWithOptions = %q", got)
	}
	if got := FindAllGroup0WithOptions(`TOKEN`, "a1b22", opt); !reflect.DeepEqual(got, []string{"1", "22"}) {
		t.Fatalf("FindAllGroup0WithOptions = %#v", got)
	}
	if got := FindAllGroup1WithOptions(`x(TOKEN)`, "x1x22", opt); !reflect.DeepEqual(got, []string{"1", "22"}) {
		t.Fatalf("FindAllGroup1WithOptions = %#v", got)
	}
	if got := FindAllGroupWithOptions(`x(TOKEN)`, "x1x22", 1, opt); !reflect.DeepEqual(got, []string{"1", "22"}) {
		t.Fatalf("FindAllGroupWithOptions = %#v", got)
	}
	if got := CountWithOptions(`TOKEN`, "a1b22", opt); got != 2 {
		t.Fatalf("CountWithOptions = %d", got)
	}
	if got := IndexOfWithOptions(`TOKEN`, "ab12", opt); got == nil || got.Text != "12" || got.Start != 2 {
		t.Fatalf("IndexOfWithOptions = %#v", got)
	}
	if got := LastIndexOfWithOptions(`TOKEN`, "ab12cd34", opt); got == nil || got.Text != "34" || got.Start != 6 {
		t.Fatalf("LastIndexOfWithOptions = %#v", got)
	}
}
