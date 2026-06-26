package vurl_test

import (
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/imajinyun/knifer-go/vurl"
)

func TestFacadeChecksAndDataURI(t *testing.T) {
	if !vurl.IsWebURL("https://example.com") || vurl.IsWebURL("ftp://example.com") {
		t.Fatal("IsWebURL failed")
	}
	if !vurl.IsAbsoluteURL("ftp://example.com") {
		t.Fatal("IsAbsoluteURL failed")
	}
	if got := vurl.DataURI("text/plain", "utf-8", "base64", "aGVsbG8="); got != "data:text/plain;charset=utf-8;base64,aGVsbG8=" {
		t.Fatalf("DataURI: %q", got)
	}
}

func TestFacadeAdditionalParseFileAndBuilderHelpers(t *testing.T) {
	if builder := vurl.NewURLBuilder().SetScheme("https").SetHost("example.com").AddPathSegment("a b").AddQuery("q", "go net").SetFragment("top"); builder.Build() != "https://example.com/a%20b?q=go+net#top" {
		t.Fatalf("NewURLBuilder Build = %q", builder.Build())
	}
	parsedBuilder, err := vurl.ParseURLBuilder("https://example.com/base?x=1#frag")
	if err != nil {
		t.Fatalf("ParseURLBuilder: %v", err)
	}
	if got := parsedBuilder.AddPathSegment("next").Build(); got != "https://example.com/base/next?x=1#frag" {
		t.Fatalf("ParseURLBuilder Build = %q", got)
	}
	if u, err := vurl.Parse(""); err != nil || u != nil {
		t.Fatalf("Parse blank = %v, %v", u, err)
	}
	if u, err := vurl.ParseHTTP("https://example.com/a b"); err != nil || u.String() != "https://example.com/a%20b" {
		t.Fatalf("ParseHTTP = %v, %v", u, err)
	}
	if _, err := vurl.ParseHTTP("/relative"); err == nil {
		t.Fatal("ParseHTTP relative error = nil")
	}
	if got := vurl.StringURI("payload"); got != "string:///payload" {
		t.Fatalf("StringURI = %q", got)
	}
	if got := vurl.StringURI("string:///payload"); got != "string:///payload" {
		t.Fatalf("StringURI existing = %q", got)
	}
	if got := vurl.EncodeBlank("a\tb\nc"); got != "a%20b%20c" {
		t.Fatalf("EncodeBlank = %q", got)
	}
	tmp := t.TempDir()
	file := filepath.Join(tmp, "data.txt")
	if err := os.WriteFile(file, []byte("file-data"), 0o600); err != nil {
		t.Fatal(err)
	}
	fileURL, err := vurl.FileURL(file)
	if err != nil {
		t.Fatalf("FileURL: %v", err)
	}
	if fileURL.Scheme != vurl.URLProtocolFile {
		t.Fatalf("FileURL scheme = %q", fileURL.Scheme)
	}
	if urls, err := vurl.FileURLs(file); err != nil || len(urls) != 1 || urls[0].Scheme != vurl.URLProtocolFile {
		t.Fatalf("FileURLs = %#v, %v", urls, err)
	}
	if _, err := vurl.FileURL(""); err == nil {
		t.Fatal("FileURL blank error = nil")
	}
	host := vurl.Host(&url.URL{Scheme: "https", Host: "example.com", Path: "/ignored"})
	if host.String() != "https://example.com" || vurl.Host(nil) != nil {
		t.Fatalf("Host helper = %v", host)
	}
	if got := vurl.FormURLEncode("a b+c"); got != "a+b%2Bc" {
		t.Fatalf("FormURLEncode = %q", got)
	}
}
