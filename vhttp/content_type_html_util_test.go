package vhttp_test

import (
	"net/http"
	"regexp"
	"testing"

	"github.com/imajinyun/go-knifer/vhttp"
)

func TestFacadeContentHTMLAndUserAgentHelpers(t *testing.T) {
	if got := vhttp.BuildContentType("text/plain", "utf-8"); got != "text/plain;charset=utf-8" {
		t.Fatalf("BuildContentType = %q", got)
	}
	if !vhttp.IsDefaultContentType("") || !vhttp.IsFormURLEncoded("application/x-www-form-urlencoded; charset=utf-8") {
		t.Fatal("content type predicates returned unexpected result")
	}
	if got := vhttp.GuessContentType(`{"ok":true}`); got != vhttp.ContentTypeJSON {
		t.Fatalf("GuessContentType json = %q", got)
	}
	if got := vhttp.GetCharsetFromContentType("text/plain; charset=gbk"); got != "gbk" {
		t.Fatalf("GetCharsetFromContentType = %q", got)
	}
	if got := vhttp.GetCharsetFromContentTypeWithOptions("text/plain; enc=big5", vhttp.WithCharsetRegexp(regexp.MustCompile(`enc=([a-z0-9-]+)`))); got != "big5" {
		t.Fatalf("GetCharsetFromContentTypeWithOptions = %q", got)
	}
	if got := vhttp.GetCharsetFromHTML(`<meta charset="utf-8"><p>x</p>`); got != "utf-8" {
		t.Fatalf("GetCharsetFromHTML = %q", got)
	}
	if got := vhttp.GetCharsetFromHTMLWithOptions(`<meta data-charset="gb2312">`, vhttp.WithMetaCharsetRegexp(regexp.MustCompile(`data-charset="([^"]+)"`))); got != "gb2312" {
		t.Fatalf("GetCharsetFromHTMLWithOptions = %q", got)
	}
	if got := vhttp.GetMimeType("payload.JSON"); got != "application/json" {
		t.Fatalf("GetMimeType = %q", got)
	}

	if got := vhttp.HTMLEscape(`<b>"go"</b>`); got != `&lt;b&gt;&#34;go&#34;&lt;/b&gt;` {
		t.Fatalf("HTMLEscape = %q", got)
	}
	if got := vhttp.HTMLUnescape("&lt;b&gt;go&lt;/b&gt;"); got != "<b>go</b>" {
		t.Fatalf("HTMLUnescape = %q", got)
	}
	if got := vhttp.CleanHTML("<p>Hello</p><!--drop-->"); got != "Hello" {
		t.Fatalf("CleanHTML = %q", got)
	}
	if got := vhttp.CleanHTMLWithOptions("a[drop]b", vhttp.WithHTMLTagRegexp(regexp.MustCompile(`\[.*?\]`)), vhttp.WithHTMLCommentRegexp(regexp.MustCompile(`$^`))); got != "ab" {
		t.Fatalf("CleanHTMLWithOptions = %q", got)
	}
	if got := vhttp.FilterHTMLTag("<div>ok</div><script>x</script>", "script"); got != "<div>ok</div>" {
		t.Fatalf("FilterHTMLTag = %q", got)
	}
	if got := vhttp.FilterHTMLTagWithOptions("<custom>drop</custom><p>keep</p>", []string{"custom"}, vhttp.WithHTMLFilterCompileFunc(regexp.Compile)); got != "<p>keep</p>" {
		t.Fatalf("FilterHTMLTagWithOptions = %q", got)
	}
	if ua := vhttp.ParseUserAgent("Mozilla/5.0 Chrome/120.0"); ua == nil {
		t.Fatal("ParseUserAgent returned nil")
	}
	if !vhttp.IsRedirected(http.StatusTemporaryRedirect) || vhttp.IsRedirected(http.StatusOK) {
		t.Fatal("IsRedirected returned unexpected result")
	}
}
