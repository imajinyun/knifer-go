package vurl_test

import (
	"strings"
	"testing"

	"github.com/imajinyun/knifer-go/vurl"
)

func TestFacadeEncodeAndURLBuilder(t *testing.T) {
	if got := vurl.EncodePath("/a b/c+d"); got != "/a%20b/c+d" {
		t.Fatalf("EncodePath = %q", got)
	}
	if got := vurl.EncodePathSegment("a/b"); got != "a%2Fb" {
		t.Fatalf("EncodePathSegment = %q", got)
	}
	if got := vurl.EncodeQuery("a b+c"); got != "a+b%2Bc" {
		t.Fatalf("EncodeQuery = %q", got)
	}
	if got, _ := vurl.DecodeForPath("a+b%2Bc"); got != "a+b+c" {
		t.Fatalf("DecodeForPath = %q", got)
	}
	if got, _ := vurl.DecodeWithOptions("a+b%2Bc", vurl.WithPlusAsSpace(false)); got != "a+b+c" {
		t.Fatalf("DecodeWithOptions = %q", got)
	}
	built := vurl.NewHTTPURLBuilder("example.com").AddPathSegment("a b").AddQuery("q", "go net").SetFragment("top 1").Build()
	if built != "http://example.com/a%20b?q=go+net#top%201" {
		t.Fatalf("URLBuilder = %q", built)
	}
}

func TestFacadeAdditionalEncodingAndQueryHelpers(t *testing.T) {
	if got := vurl.EncodeWithOptions("a b", vurl.WithQueryEscapeFunc(func(s string) string { return "escaped:" + s })); got != "escaped:a b" {
		t.Fatalf("EncodeWithOptions = %q", got)
	}
	if got := vurl.URLEncodeWithOptions("a b", vurl.WithQueryEscapeFunc(strings.ToUpper)); got != "A B" {
		t.Fatalf("URLEncodeWithOptions = %q", got)
	}
	if got := vurl.EncodeQueryWithOptions("a b", vurl.WithQueryEscapeFunc(func(s string) string { return "q:" + s })); got != "q:a b" {
		t.Fatalf("EncodeQueryWithOptions = %q", got)
	}
	if got := vurl.EncodePathSegmentWithOptions("a/b", vurl.WithPathEscapeFunc(func(s string) string { return "p:" + s })); got != "p:a/b" {
		t.Fatalf("EncodePathSegmentWithOptions = %q", got)
	}
	if got := vurl.FormURLEncodeWithOptions("a b", vurl.WithQueryEscapeFunc(func(s string) string { return "form:" + s })); got != "form:a b" {
		t.Fatalf("FormURLEncodeWithOptions = %q", got)
	}
	if got := vurl.EncodeAll("a b"); got != "a%20b" {
		t.Fatalf("EncodeAll = %q", got)
	}
	if got := vurl.EncodeFragment("a b#c"); got != "a%20b%23c" {
		t.Fatalf("EncodeFragment = %q", got)
	}

	decoded, err := vurl.Decode("a+b%2Fc")
	if err != nil || decoded != "a b/c" {
		t.Fatalf("Decode = %q, %v", decoded, err)
	}
	decoded, err = vurl.DecodePlus("a+b%2Fc", false)
	if err != nil || decoded != "a+b/c" {
		t.Fatalf("DecodePlus(false) = %q, %v", decoded, err)
	}

	if got := vurl.EncodeParams("https://example.com/search?q=a b&lang=go"); got != "https://example.com/search?lang=go&q=a+b" {
		t.Fatalf("EncodeParams = %q", got)
	}
	if got := vurl.DecodeQueryFirst("a=1&a=2&b=x+y"); got["a"] != "1" || got["b"] != "x y" {
		t.Fatalf("DecodeQueryFirst = %#v", got)
	}
	if got := vurl.DecodeQuery("a=1&a=2"); len(got["a"]) != 2 || got["a"][0] != "1" || got["a"][1] != "2" {
		t.Fatalf("DecodeQuery = %#v", got)
	}
	if got := vurl.AppendQuery("https://example.com/path?x=1", map[string]any{"q": "go net"}); got != "https://example.com/path?x=1&q=go+net" {
		t.Fatalf("AppendQuery = %q", got)
	}
}

func BenchmarkEncodeQueryMap(b *testing.B) {
	values := map[string]any{"page": 2, "q": "go knifer", "tag": "safe url"}
	var out string
	for b.Loop() {
		out = vurl.EncodeQueryMap(values)
	}
	_ = out
}
