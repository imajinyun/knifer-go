package url

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

func TestEncodeBlankAndParseHTTP(t *testing.T) {
	if got := EncodeBlank("https://example.com/a b"); got != "https://example.com/a%20b" {
		t.Fatalf("EncodeBlank: %q", got)
	}
	u, err := ParseHTTP("https://example.com/a b")
	if err != nil {
		t.Fatalf("ParseHTTP: %v", err)
	}
	if u.EscapedPath() != "/a%20b" {
		t.Fatalf("path: %q", u.EscapedPath())
	}
}

func TestNormalizeAndComplete(t *testing.T) {
	if got := Normalize("\\example.com\\a b", true, true); got != "http://example.com/a%20b" {
		t.Fatalf("Normalize: %q", got)
	}
	if got := NormalizeWithOptions("example.com/a", false, false, WithDefaultScheme("https")); got != "https://example.com/a" {
		t.Fatalf("NormalizeWithOptions: %q", got)
	}
	got, err := Complete("example.com/dir/", "a.html")
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if got != "http://example.com/dir/a.html" {
		t.Fatalf("Complete got %q", got)
	}
	if got := NormalizeUsingOptions("example.com//a b", WithDefaultScheme("https"), WithEncodePath(true), WithReplaceSlash(true)); got != "https://example.com/a%20b" {
		t.Fatalf("NormalizeUsingOptions: %q", got)
	}
}

func TestOpenAndContentLengthWithOptions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Token") != "secret" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Length", "4")
		_, _ = w.Write([]byte("body"))
	}))
	defer srv.Close()

	r, err := OpenWithOptions(srv.URL, WithHeader("X-Token", "secret"), WithTimeout(time.Second), WithCheckStatus(true))
	if err != nil {
		t.Fatalf("OpenWithOptions: %v", err)
	}
	data, err := io.ReadAll(r)
	_ = r.Close()
	if err != nil || string(data) != "body" {
		t.Fatalf("body = %q, %v", data, err)
	}
	length, err := ContentLengthWithOptions(srv.URL, WithHeader("X-Token", "secret"), WithCheckStatus(true))
	if err != nil || length != 4 {
		t.Fatalf("ContentLengthWithOptions = %d, %v", length, err)
	}
	if _, err := OpenWithOptions(srv.URL, WithCheckStatus(true)); err == nil {
		t.Fatal("OpenWithOptions status check error = nil")
	}
}

func TestResourceProviderOptions(t *testing.T) {
	openedPath := ""
	r, err := OpenWithOptions("file:///virtual/data.txt", WithOpenFile(func(path string) (io.ReadCloser, error) {
		openedPath = path
		return io.NopCloser(strings.NewReader("virtual-body")), nil
	}))
	if err != nil {
		t.Fatalf("OpenWithOptions custom open file: %v", err)
	}
	data, err := io.ReadAll(r)
	_ = r.Close()
	if err != nil || string(data) != "virtual-body" || openedPath != "/virtual/data.txt" {
		t.Fatalf("custom open file data=%q path=%q err=%v", data, openedPath, err)
	}

	statSource := t.TempDir() + "/stat.txt"
	if err := os.WriteFile(statSource, []byte("1234567"), 0o600); err != nil {
		t.Fatal(err)
	}
	statPath := ""
	length, err := ContentLengthWithOptions("/virtual/stat.txt", WithStat(func(path string) (os.FileInfo, error) {
		statPath = path
		return os.Stat(statSource)
	}))
	if err != nil || length != 7 || statPath != "/virtual/stat.txt" {
		t.Fatalf("custom stat length=%d path=%q err=%v", length, statPath, err)
	}
}

func TestResourceRequestFactoryOption(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Factory") != "yes" {
			http.Error(w, "factory header missing", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte("factory"))
	}))
	defer srv.Close()

	var gotMethod, gotURL string
	factory := func(ctx context.Context, method, raw string) (*http.Request, error) {
		gotMethod, gotURL = method, raw
		req, err := http.NewRequestWithContext(ctx, method, raw, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("X-Factory", "yes")
		return req, nil
	}
	r, err := OpenWithOptions(srv.URL, WithRequestFactory(factory), WithCheckStatus(true))
	if err != nil {
		t.Fatalf("OpenWithOptions with request factory: %v", err)
	}
	data, err := io.ReadAll(r)
	_ = r.Close()
	if err != nil || string(data) != "factory" || gotMethod != http.MethodGet || gotURL != srv.URL {
		t.Fatalf("factory data=%q method=%q url=%q err=%v", data, gotMethod, gotURL, err)
	}
}

func TestQueryHelpers(t *testing.T) {
	queryPart := URLEncode("a b&c=d")
	if queryPart != "a+b%26c%3Dd" {
		t.Fatalf("URLEncode: %q", queryPart)
	}
	decoded, err := URLDecode(queryPart)
	if err != nil || decoded != "a b&c=d" {
		t.Fatalf("URLDecode: %v %q", err, decoded)
	}
	encoded := BuildQuery(map[string]any{"a": "1", "b": "x y", "": "skip"})
	if !strings.Contains(encoded, "a=1") || !strings.Contains(encoded, "b=x+y") || strings.Contains(encoded, "skip") {
		t.Fatalf("BuildQuery: %q", encoded)
	}
	if got := EncodeParams("https://example.com/?q=a b"); got != "https://example.com/?q=a+b" {
		t.Fatalf("EncodeParams: %q", got)
	}
	if got := DecodeQueryFirst("a=1&a=2&b=x+y"); got["a"] == "" || got["b"] != "x y" {
		t.Fatalf("DecodeQueryFirst: %#v", got)
	}
	if got := AppendQuery("https://example.com/path?x=1", map[string]any{"y": 2}); !strings.Contains(got, "x=1") || !strings.Contains(got, "y=2") {
		t.Fatalf("AppendQuery: %q", got)
	}
}

func TestURLChecksAndDataURI(t *testing.T) {
	if !IsHTTP("Http://example.com") || !IsHTTPS("HTTPS://example.com") {
		t.Fatal("scheme prefix checks failed")
	}
	if !IsWebURL("https://example.com/a") || IsWebURL("ftp://example.com/a") {
		t.Fatal("web URL checks failed")
	}
	if !IsAbsoluteURL("ftp://example.com/a") || IsAbsoluteURL("/relative") {
		t.Fatal("absolute URL checks failed")
	}
	if got := DataURIBase64("image/png", "AAAA"); got != "data:image/png;base64,AAAA" {
		t.Fatalf("DataURIBase64: %q", got)
	}
}

func TestHostDecodedPathAndJar(t *testing.T) {
	u, _ := url.Parse("https://example.com/a%20b?q=1")
	if got := Host(u).String(); got != "https://example.com" {
		t.Fatalf("Host: %q", got)
	}
	if got := DecodedPath(u); got != "/a b" {
		t.Fatalf("DecodedPath: %q", got)
	}
	jar, _ := url.Parse("file:///tmp/a.jar")
	if !IsFileURL(jar) || !IsJarFileURL(jar) {
		t.Fatal("jar file checks failed")
	}
}

func TestEncodeAndURLBuilder(t *testing.T) {
	if got := EncodePath("/a b/c+d"); got != "/a%20b/c+d" {
		t.Fatalf("EncodePath = %q", got)
	}
	if got := EncodePathSegment("a/b"); got != "a%2Fb" {
		t.Fatalf("EncodePathSegment = %q", got)
	}
	if got := EncodeQuery("a b+c"); got != "a+b%2Bc" {
		t.Fatalf("EncodeQuery = %q", got)
	}
	if got := EncodeQueryWithOptions("a b", WithQueryEscapeFunc(func(s string) string { return "query:" + s })); got != "query:a b" {
		t.Fatalf("EncodeQueryWithOptions = %q", got)
	}
	if got := FormURLEncodeWithOptions("a b", WithQueryEscapeFunc(func(s string) string { return "form:" + s })); got != "form:a b" {
		t.Fatalf("FormURLEncodeWithOptions = %q", got)
	}
	if got := EncodePathSegmentWithOptions("a/b", WithPathEscapeFunc(func(s string) string { return "path:" + s })); got != "path:a/b" {
		t.Fatalf("EncodePathSegmentWithOptions = %q", got)
	}
	if got := EncodeWithOptions("a b", WithQueryEscapeFunc(func(s string) string { return "encode:" + s })); got != "encode:a b" {
		t.Fatalf("EncodeWithOptions = %q", got)
	}
	if got, _ := DecodePlus("a+b%2Bc", false); got != "a+b+c" {
		t.Fatalf("DecodePlus = %q", got)
	}
	if got, _ := DecodeWithOptions("a+b%2Bc", WithPlusAsSpace(false)); got != "a+b+c" {
		t.Fatalf("DecodeWithOptions = %q", got)
	}
	built := NewHTTPURLBuilder("example.com").AddPathSegment("a b").AddQuery("q", "go net").SetFragment("top 1").Build()
	if built != "http://example.com/a%20b?q=go+net#top%201" {
		t.Fatalf("URLBuilder = %q", built)
	}
}
