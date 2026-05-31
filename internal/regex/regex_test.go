package regex

import (
	"reflect"
	"regexp"
	"testing"
)

func TestBasicRegexCompatibility(t *testing.T) {
	if !ReMatch(`^\d+$`, "123") || ReMatch(`^\d+$`, "12a") || ReMatch(`(`, "x") {
		t.Fatalf("ReMatch failed")
	}
	if ReFind(`\d+`, "ab123cd") != "123" || ReFind(`(`, "x") != "" {
		t.Fatalf("ReFind failed")
	}
	all := ReFindAll(`\d+`, "a1b22c333")
	if !reflect.DeepEqual(all, []string{"1", "22", "333"}) {
		t.Fatalf("ReFindAll failed: %v", all)
	}
	if ReReplace(`\d`, "a1b2", "*") != "a*b*" || ReReplace(`(`, "x", "*") != "x" {
		t.Fatalf("ReReplace failed")
	}
}

func TestGetGroupsAndNamedGroups(t *testing.T) {
	pattern := `(?<year>\d{4})-(?<month>\d{2})-(?<day>\d{2})`
	content := "date=2026-05-31"

	if got := GetGroup0(pattern, content); got != "2026-05-31" {
		t.Fatalf("GetGroup0 = %q", got)
	}
	if got := GetGroup1(pattern, content); got != "2026" {
		t.Fatalf("GetGroup1 = %q", got)
	}
	if got := GetByName(pattern, content, "month"); got != "05" {
		t.Fatalf("GetByName = %q", got)
	}
	groups := GetAllGroups(pattern, content, true, false)
	if !reflect.DeepEqual(groups, []string{"2026-05-31", "2026", "05", "31"}) {
		t.Fatalf("GetAllGroups = %#v", groups)
	}
	names := GetAllGroupNames(pattern, content)
	if names["year"] != "2026" || names["month"] != "05" || names["day"] != "31" {
		t.Fatalf("GetAllGroupNames = %#v", names)
	}
}

func TestExtractAndDelete(t *testing.T) {
	if got := ExtractMulti(`(.*?)年(.*?)月`, "2026年5月", "$1-$2"); got != "2026-5" {
		t.Fatalf("ExtractMulti = %q", got)
	}
	content := "prefix 2026年5月 suffix"
	if got := ExtractMultiAndDelPre(`(\d+)年(\d+)月`, &content, "$1-$2"); got != "2026-5" {
		t.Fatalf("ExtractMultiAndDelPre = %q", got)
	}
	if content != " suffix" {
		t.Fatalf("content after delete = %q", content)
	}

	if got := DelFirst(`\d+`, "a123b456"); got != "ab456" {
		t.Fatalf("DelFirst = %q", got)
	}
	if got := DelLast(`\d+`, "a123b456"); got != "a123b" {
		t.Fatalf("DelLast = %q", got)
	}
	if got := DelAll(`\d+`, "a123b456"); got != "ab" {
		t.Fatalf("DelAll = %q", got)
	}
	if got := DelPre(`\d+`, "abc123xyz"); got != "xyz" {
		t.Fatalf("DelPre = %q", got)
	}
}

func TestFindCountIndexMatchReplaceAndEscape(t *testing.T) {
	if got := FindAllGroup1(`(\d+)`, "a1b22c333"); !reflect.DeepEqual(got, []string{"1", "22", "333"}) {
		t.Fatalf("FindAllGroup1 = %#v", got)
	}
	if got := Count(`\d+`, "a1b22c333"); got != 3 {
		t.Fatalf("Count = %d", got)
	}
	if !Contains(`\d+`, "abc1") || Contains(`\d+`, "abc") {
		t.Fatalf("Contains failed")
	}
	first := IndexOf(`\d+`, "ab12cd34")
	last := LastIndexOf(`\d+`, "ab12cd34")
	if first == nil || first.Start != 2 || first.End != 4 || last == nil || last.Start != 6 || last.End != 8 {
		t.Fatalf("IndexOf/LastIndexOf failed: %#v %#v", first, last)
	}
	if n, ok := GetFirstNumber("abc123def"); !ok || n != 123 {
		t.Fatalf("GetFirstNumber = %d %v", n, ok)
	}
	if !IsMatch(`\d+`, "123") || IsMatch(`\d+`, "a123") {
		t.Fatalf("IsMatch failed")
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

func TestCompiledRegexpHelpers(t *testing.T) {
	re := regexp.MustCompile(`(\w+)=(\d+)`)
	var seen []string
	Each(re, "a=1 b=22", func(m MatchResult) { seen = append(seen, m.Groups[2]) })
	if !reflect.DeepEqual(seen, []string{"1", "22"}) {
		t.Fatalf("Each = %#v", seen)
	}
	if got := ReplaceFirstRe(re, "a=1 b=22", `$1:x`); got != "a:x b=22" {
		t.Fatalf("ReplaceFirstRe = %q", got)
	}
}
