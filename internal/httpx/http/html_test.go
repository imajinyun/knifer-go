package http

import (
	"regexp"
	"strings"
	"testing"
)

// Covers the utility toolkit-http HtmlUtilTest.

func TestHTMLEscape(t *testing.T) {
	html := "<html><body>123'123'</body></html>"
	escape := HTMLEscape(html)
	if !strings.Contains(escape, "&lt;html&gt;") || !strings.Contains(escape, "&lt;/body&gt;") {
		t.Fatalf("escape output: %q", escape)
	}
	if HTMLUnescape(escape) != html {
		t.Fatalf("unescape failed: %q", HTMLUnescape(escape))
	}
	if HTMLUnescape("&apos;") != "'" {
		t.Fatalf("unescape apos failed")
	}
}

func TestHTMLEscapeNbsp(t *testing.T) {
	if HTMLUnescape("&nbsp;") != "\u00A0" {
		t.Fatalf("nbsp unescape failed: %q", HTMLUnescape("&nbsp;"))
	}
}

func TestCleanHTMLTag(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{`pre<img src="xxx/dfdsfds/test.jpg">`, "pre"},
		{`pre<img>`, "pre"},
		{`pre<img src="xxx/dfdsfds/test.jpg" />`, "pre"},
		{`pre<img />`, "pre"},
		{`pre<div class="test_div">dfdsfdsfdsf</div>`, "predfdsfdsfdsf"},
	}
	for _, c := range cases {
		if got := CleanHTML(c.in); got != c.want {
			t.Fatalf("CleanHTML(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestFilterHTMLTag(t *testing.T) {
	str := `pre<img src="xxx/dfdsfds/test.jpg">`
	if got := FilterHTMLTag(str, "img"); got != "pre" {
		t.Fatalf("FilterHTMLTag img: %q", got)
	}
	str = `pre<div class="test_div">dfdsfdsfdsf</div>`
	if got := FilterHTMLTag(str, "div"); got != "pre" {
		t.Fatalf("FilterHTMLTag div: %q", got)
	}
	// Multiple tags.
	got := FilterHTMLTag(`<html><img src='x'><i>测试</i></html>`, "i", "br")
	if got != `<html><img src='x'></html>` {
		t.Fatalf("FilterHTMLTag multi: %q", got)
	}
}

func TestNilHTMLFilterCompileFuncDoesNotOverwriteConfiguredProvider(t *testing.T) {
	cfg := applyHTMLFilterOptions([]HTMLFilterOption{
		WithHTMLFilterCompileFunc(regexp.Compile),
		WithHTMLFilterCompileFunc(nil),
	})
	if cfg.compile == nil {
		t.Fatal("nil WithHTMLFilterCompileFunc should not overwrite configured compiler")
	}
}

// Covers HTMLFilterTest issue3433Test with a simplified removal of unsafe attributes and tags.
func TestCleanHTMLPreservesText(t *testing.T) {
	got := CleanHTML(`<p onclick="bbbb">a</p>`)
	if got != "a" {
		t.Fatalf("CleanHTML onclick: %q", got)
	}
}
