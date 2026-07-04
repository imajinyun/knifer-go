package regex

import (
	"math"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func TestRegexHelpersWithOptions(t *testing.T) {
	compiler := func(pattern string) (*regexp.Regexp, error) {
		return regexp.Compile(strings.ReplaceAll(pattern, "TOKEN", `\d+`))
	}
	opt := WithCompileFunc(compiler)

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
	if got := FindAllWithOptions(`x(TOKEN)`, "x1x22", 1, opt); !reflect.DeepEqual(got, []string{"1", "22"}) {
		t.Fatalf("FindAllWithOptions = %#v", got)
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

func TestSpecializedRegexOptions(t *testing.T) {
	if n, ok := GetFirstNumberWithOptions("v12.34", WithNumbersRegexp(regexp.MustCompile(`[3-9]\d`))); !ok || n != 34 {
		t.Fatalf("GetFirstNumberWithOptions = %d %v", n, ok)
	}
	if got := TemplateVarsWithOptions("${3} $1", WithGroupVarRegexp(regexp.MustCompile(`\$\{(\d+)\}`))); !reflect.DeepEqual(got, []int{3}) {
		t.Fatalf("TemplateVarsWithOptions = %#v", got)
	}
	if got := GetByNameWithOptions(`(?<word>\w+)`, "abc", "word", WithNamedGroupNormalizer(func(pattern string) string {
		return strings.ReplaceAll(pattern, `(?<`, `(?P<`)
	})); got != "abc" {
		t.Fatalf("GetByNameWithOptions custom normalizer = %q", got)
	}
}

func TestRegexOptionFallbacksAndNumberBoundaries(t *testing.T) {
	clearProviders := func(c *regexConfig) {
		c.compile = nil
		c.groupVarRegexp = nil
		c.numbersRegexp = nil
		c.namedGroupRegexp = nil
	}

	if !ReMatchWithOptions(`\d+`, "123", clearProviders) {
		t.Fatal("nil compile provider should fall back to regexp.Compile")
	}
	if got := TemplateVarsWithOptions("$2 $1", clearProviders); !reflect.DeepEqual(got, []int{2, 1}) {
		t.Fatalf("TemplateVarsWithOptions fallback = %#v", got)
	}
	if n, ok := GetFirstNumberWithOptions("v42", clearProviders); !ok || n != 42 {
		t.Fatalf("GetFirstNumberWithOptions fallback = %d %v", n, ok)
	}
	if got := GetByNameWithOptions(`(?<word>\w+)`, "abc", "word", clearProviders); got != "abc" {
		t.Fatalf("GetByNameWithOptions fallback = %q", got)
	}

	tooLarge := regexp.MustCompile(regexp.QuoteMeta(strconv.Itoa(math.MaxInt)) + `\d`)
	if n, ok := GetFirstNumberWithOptions(strconv.Itoa(math.MaxInt)+"9", WithNumbersRegexp(tooLarge)); ok || n != 0 {
		t.Fatalf("GetFirstNumberWithOptions overflow = %d %v", n, ok)
	}
}

func TestNilRegexProviderOptionsDoNotOverwriteConfiguredProviders(t *testing.T) {
	compile := func(pattern string) (*regexp.Regexp, error) {
		return regexp.Compile(regexp.QuoteMeta(pattern))
	}
	groupRe := regexp.MustCompile(`\$\{(\d+)\}`)
	numberRe := regexp.MustCompile(`[4-9]\d`)
	cfg := applyOptions([]Option{
		WithCompileFunc(compile),
		WithCompileFunc(nil),
		WithGroupVarRegexp(groupRe),
		WithGroupVarRegexp(nil),
		WithNumbersRegexp(numberRe),
		WithNumbersRegexp(nil),
	})
	if cfg.compile == nil || cfg.groupVarRegexp != groupRe || cfg.numbersRegexp != numberRe {
		t.Fatalf("nil regex provider option overwrote configured provider: %#v", cfg)
	}
}

func TestIsMatchEmptyPatternContract(t *testing.T) {
	if IsMatch("", "") {
		t.Fatal("empty pattern should not match empty content")
	}
	if !IsMatch("", "non-empty") {
		t.Fatal("empty pattern should match non-empty content")
	}
}
