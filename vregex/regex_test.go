package vregex

import (
	"reflect"
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
