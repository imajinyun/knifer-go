package vurl_test

import (
	"net/url"
	"strings"
	"testing"

	"github.com/imajinyun/knifer-go/vurl"
)

func TestFacadeQueryAndNormalize(t *testing.T) {
	if got := vurl.Normalize("example.com/a b", true, false); got != "http://example.com/a%20b" {
		t.Fatalf("Normalize: %q", got)
	}
	encoded := vurl.URLEncode("a b+c/中文")
	decoded, err := vurl.URLDecode(encoded)
	if err != nil || decoded != "a b+c/中文" {
		t.Fatalf("URL query roundtrip = %q, %v", decoded, err)
	}
	query := vurl.BuildQuery(map[string]any{"a": "1", "b": "x y"})
	if !strings.Contains(query, "a=1") || !strings.Contains(query, "b=x+y") {
		t.Fatalf("BuildQuery: %q", query)
	}
	completed, err := vurl.Complete("example.com/base/", "next")
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if completed != "http://example.com/base/next" {
		t.Fatalf("Complete: %q", completed)
	}
}

func TestFacadeNormalizeWithOptions(t *testing.T) {
	got := vurl.NormalizeWithOptions("example.com/a b", true, false, vurl.WithDefaultScheme("https"))
	if got != "https://example.com/a%20b" {
		t.Fatalf("NormalizeWithOptions = %q", got)
	}
	got = vurl.NormalizeUsingOptions("example.com//a b", vurl.WithDefaultScheme("https"), vurl.WithEncodePath(true), vurl.WithReplaceSlash(true))
	if got != "https://example.com/a%20b" {
		t.Fatalf("NormalizeUsingOptions = %q", got)
	}
}

func TestFacadePathAndSchemeHelpers(t *testing.T) {
	path, err := vurl.Path("https://example.com/a%20b/file.txt?x=1")
	if err != nil || path != "/a b/file.txt" {
		t.Fatalf("Path = %q, %v", path, err)
	}
	u, err := url.Parse("file:///tmp/demo.jar")
	if err != nil {
		t.Fatal(err)
	}
	if got := vurl.DecodedPath(u); got != "/tmp/demo.jar" {
		t.Fatalf("DecodedPath = %q", got)
	}
	if !vurl.IsFileURL(u) {
		t.Fatal("IsFileURL(file URL) = false")
	}
	if !vurl.IsJarFileURL(u) {
		t.Fatal("IsJarFileURL(file .jar URL) = false")
	}
	jar, err := url.Parse("jar:file:///tmp/demo.jar!/BOOT-INF/classes")
	if err != nil {
		t.Fatal(err)
	}
	if !vurl.IsJarURL(jar) {
		t.Fatal("IsJarURL(jar URL) = false")
	}
	if uri, err := vurl.ToURI("https://example.com/a b", true); err != nil || uri.String() != "https://example.com/a%20b" {
		t.Fatalf("ToURI = %v, %v", uri, err)
	}
	if !vurl.IsHTTP("http://example.com") || !vurl.IsHTTPS("https://example.com") || !vurl.IsHTTPSURL("https://example.com") {
		t.Fatal("HTTP/HTTPS scheme helpers failed")
	}
	if got := vurl.DataURIBase64("text/plain", "aGVsbG8="); got != "data:text/plain;base64,aGVsbG8=" {
		t.Fatalf("DataURIBase64 = %q", got)
	}
}

func BenchmarkNormalizeUsingOptions(b *testing.B) {
	var out string
	for b.Loop() {
		out = vurl.NormalizeUsingOptions("example.com//a b", vurl.WithDefaultScheme("https"), vurl.WithEncodePath(true), vurl.WithReplaceSlash(true))
	}
	_ = out
}
