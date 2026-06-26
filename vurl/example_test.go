package vurl_test

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"time"

	"github.com/imajinyun/knifer-go/vurl"
)

func ExampleAppendQuery() {
	fmt.Println(vurl.AppendQuery("https://example.com/search?lang=en", map[string]any{"q": "go url"}))
	// Output: https://example.com/search?lang=en&q=go+url
}

func ExampleBuildQuery() {
	fmt.Println(vurl.BuildQuery(map[string]any{"page": 2, "q": "go"}))
	// Output: page=2&q=go
}

func ExampleComplete() {
	full, err := vurl.Complete("https://example.com/docs/", "api/tools.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(full)
	// Output: https://example.com/docs/api/tools.json
}

func ExampleContentLength() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "5")
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte("hello"))
		}
	}))
	defer server.Close()

	n, err := vurl.ContentLength(server.URL)
	fmt.Println(n, err == nil)
	// Output: 5 true
}

func ExampleContentLengthSafe() {
	n, err := vurl.ContentLengthSafe("file:///tmp/secret.txt")
	fmt.Println(n, err != nil)
	// Output: -1 true
}

func ExampleContentLengthSafeWithOptions() {
	n, err := vurl.ContentLengthSafeWithOptions(
		"https://example.com/data",
		vurl.WithHTTPClient(exampleURLClient("payload", http.StatusOK)),
		vurl.WithLookupIP(examplePublicLookupIP),
	)
	fmt.Println(n, err == nil)
	// Output: 7 true
}

func ExampleContentLengthWithOptions() {
	n, err := vurl.ContentLengthWithOptions("memory.txt", vurl.WithStat(exampleURLStat(12)))
	fmt.Println(n, err == nil)
	// Output: 12 true
}

func ExampleDataURI() {
	fmt.Println(vurl.DataURI("text/plain", "utf-8", "base64", "aGVsbG8="))
	// Output: data:text/plain;charset=utf-8;base64,aGVsbG8=
}

func ExampleDataURIBase64() {
	fmt.Println(vurl.DataURIBase64("text/plain", "aGVsbG8="))
	// Output: data:text/plain;base64,aGVsbG8=
}

func ExampleDecode() {
	decoded, err := vurl.Decode("a+b%26c")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(decoded)
	// Output: a b&c
}

func ExampleDecodeForPath() {
	decoded, err := vurl.DecodeForPath("a+b%2Fc")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(decoded)
	// Output: a+b/c
}

func ExampleDecodePlus() {
	query, _ := vurl.DecodePlus("a+b", true)
	path, _ := vurl.DecodePlus("a+b", false)
	fmt.Println(query, path)
	// Output: a b a+b
}

func ExampleDecodeQuery() {
	values := vurl.DecodeQuery("tag=go&tag=url&q=a+b")
	fmt.Println(values["tag"][0], values["tag"][1], values["q"][0])
	// Output: go url a b
}

func ExampleDecodeQueryFirst() {
	values := vurl.DecodeQueryFirst("tag=go&tag=url&q=a+b")
	fmt.Println(values["tag"], values["q"])
	// Output: go a b
}

func ExampleDecodeWithOptions() {
	decoded, err := vurl.DecodeWithOptions("a+b", vurl.WithPlusAsSpace(false))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(decoded)
	// Output: a+b
}

func ExampleDecodedPath() {
	u, _ := url.Parse("https://example.com/a%20b/c")
	fmt.Println(vurl.DecodedPath(u))
	// Output: /a b/c
}

func ExampleEncode() {
	fmt.Println(vurl.Encode("a b&c"))
	// Output: a+b%26c
}

func ExampleEncodeAll() {
	fmt.Println(vurl.EncodeAll("a b/c?d"))
	// Output: a%20b%2Fc%3Fd
}

func ExampleEncodeBlank() {
	fmt.Println(vurl.EncodeBlank("https://example.com/a b\tc"))
	// Output: https://example.com/a%20b%20c
}

func ExampleEncodeFragment() {
	fmt.Println(vurl.EncodeFragment("section 1?x=go"))
	// Output: section%201?x=go
}

func ExampleEncodeParams() {
	fmt.Println(vurl.EncodeParams("https://example.com/search?q=go url&lang=en"))
	// Output: https://example.com/search?lang=en&q=go+url
}

func ExampleEncodePath() {
	fmt.Println(vurl.EncodePath("/a b/c+d"))
	// Output: /a%20b/c+d
}

func ExampleEncodePathSegment() {
	fmt.Println(vurl.EncodePathSegment("a/b c"))
	// Output: a%2Fb%20c
}

func ExampleEncodePathSegmentWithOptions() {
	out := vurl.EncodePathSegmentWithOptions("a/b", vurl.WithPathEscapeFunc(strings.ToUpper))
	fmt.Println(out)
	// Output: A/B
}

func ExampleEncodeQuery() {
	fmt.Println(vurl.EncodeQuery("a b&c"))
	// Output: a+b%26c
}

func ExampleEncodeQueryMap() {
	fmt.Println(vurl.EncodeQueryMap(map[string]any{"q": "go"}))
	// Output: q=go
}

func ExampleEncodeQueryWithOptions() {
	out := vurl.EncodeQueryWithOptions("go url", vurl.WithQueryEscapeFunc(func(s string) string {
		return "query:" + s
	}))
	fmt.Println(out)
	// Output: query:go url
}

func ExampleEncodeWithOptions() {
	out := vurl.EncodeWithOptions("go url", vurl.WithQueryEscapeFunc(func(s string) string {
		return strings.ReplaceAll(s, " ", "_")
	}))
	fmt.Println(out)
	// Output: go_url
}

func ExampleFileURL() {
	u, err := vurl.FileURL("/tmp/knifer-go.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(u.Scheme, u.Path)
	// Output: file /tmp/knifer-go.txt
}

func ExampleFileURLs() {
	urls, err := vurl.FileURLs("/tmp/a.txt", "/tmp/b.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(len(urls), urls[0].Path, urls[1].Path)
	// Output: 2 /tmp/a.txt /tmp/b.txt
}

func ExampleFormURLEncode() {
	fmt.Println(vurl.FormURLEncode("a b&c"))
	// Output: a+b%26c
}

func ExampleFormURLEncodeWithOptions() {
	out := vurl.FormURLEncodeWithOptions("a b", vurl.WithQueryEscapeFunc(func(s string) string {
		return "form:" + s
	}))
	fmt.Println(out)
	// Output: form:a b
}

func ExampleHost() {
	u, _ := url.Parse("https://example.com/docs?q=go#top")
	fmt.Println(vurl.Host(u))
	// Output: https://example.com
}

func ExampleIsAbsoluteURL() {
	fmt.Println(vurl.IsAbsoluteURL("https://example.com/a"), vurl.IsAbsoluteURL("/a"))
	// Output: true false
}

func ExampleIsFileURL() {
	fileURL, _ := url.Parse("file:///tmp/app.jar")
	httpURL, _ := url.Parse("https://example.com")
	fmt.Println(vurl.IsFileURL(fileURL), vurl.IsFileURL(httpURL))
	// Output: true false
}

func ExampleIsHTTP() {
	fmt.Println(vurl.IsHTTP("http://example.com"), vurl.IsHTTP("https://example.com"))
	// Output: true false
}

func ExampleIsHTTPS() {
	fmt.Println(vurl.IsHTTPS("https://example.com"), vurl.IsHTTPS("http://example.com"))
	// Output: true false
}

func ExampleIsHTTPSURL() {
	fmt.Println(vurl.IsHTTPSURL("https://example.com"), vurl.IsHTTPSURL("http://example.com"))
	// Output: true false
}

func ExampleIsHTTPURL() {
	fmt.Println(vurl.IsHTTPURL("http://example.com"))
	fmt.Println(vurl.IsHTTPURL("ftp://example.com"))
	// Output:
	// true
	// false
}

func ExampleIsJarFileURL() {
	jarURL, _ := url.Parse("file:///opt/app.jar")
	txtURL, _ := url.Parse("file:///opt/readme.txt")
	fmt.Println(vurl.IsJarFileURL(jarURL), vurl.IsJarFileURL(txtURL))
	// Output: true false
}

func ExampleIsJarURL() {
	jarURL, _ := url.Parse("jar:file:///opt/app.jar!/config.yaml")
	fileURL, _ := url.Parse("file:///opt/app.jar")
	fmt.Println(vurl.IsJarURL(jarURL), vurl.IsJarURL(fileURL))
	// Output: true false
}

func ExampleIsWebURL() {
	fmt.Println(vurl.IsWebURL("https://example.com"), vurl.IsWebURL("file:///tmp/a"))
	// Output: true false
}

func ExampleNewHTTPURLBuilder() {
	b := vurl.NewHTTPURLBuilder("example.com").AddPath("docs").AddQuery("q", "go")
	fmt.Println(b.Build())
	// Output: http://example.com/docs?q=go
}

func ExampleNewURLBuilder() {
	b := vurl.NewURLBuilder().SetScheme("https").SetHost("example.com").AddPathSegment("a b")
	fmt.Println(b.Build())
	// Output: https://example.com/a%20b
}

func ExampleNormalize() {
	fmt.Println(vurl.Normalize("example.com//a b", true, true))
	// Output: http://example.com/a%20b
}

func ExampleNormalizeUsingOptions() {
	out := vurl.NormalizeUsingOptions(
		"example.com//a b",
		vurl.WithDefaultScheme("https"),
		vurl.WithEncodePath(true),
		vurl.WithReplaceSlash(true),
	)
	fmt.Println(out)
	// Output: https://example.com/a%20b
}

func ExampleNormalizeWithOptions() {
	out := vurl.NormalizeWithOptions(
		"example.com/a b",
		true,
		false,
		vurl.WithDefaultScheme("https"),
	)
	fmt.Println(out)
	// Output: https://example.com/a%20b
}

func ExampleOpen() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hello"))
	}))
	defer server.Close()

	rc, err := vurl.Open(server.URL)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() { _ = rc.Close() }()
	data, _ := io.ReadAll(rc)
	fmt.Println(string(data))
	// Output: hello
}

func ExampleOpenSafe() {
	rc, err := vurl.OpenSafe("file:///tmp/secret.txt")
	if rc != nil {
		_ = rc.Close()
	}
	fmt.Println(err != nil)
	// Output: true
}

func ExampleOpenSafeWithOptions() {
	rc, err := vurl.OpenSafeWithOptions(
		"https://example.com/data",
		vurl.WithHTTPClient(exampleURLClient("safe", http.StatusOK)),
		vurl.WithLookupIP(examplePublicLookupIP),
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() { _ = rc.Close() }()
	data, _ := io.ReadAll(rc)
	fmt.Println(string(data))
	// Output: safe
}

func ExampleOpenWithOptions() {
	rc, err := vurl.OpenWithOptions("memory.txt", vurl.WithOpenFile(exampleURLOpenFile("memory")))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() { _ = rc.Close() }()
	data, _ := io.ReadAll(rc)
	fmt.Println(string(data))
	// Output: memory
}

func ExampleParse() {
	u, err := vurl.Parse(" https://example.com/docs?q=go ")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(u.Scheme, u.Host, u.Query().Get("q"))
	// Output: https example.com go
}

func ExampleParseHTTP() {
	u, err := vurl.ParseHTTP("https://example.com/a b")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(u.String())
	// Output: https://example.com/a%20b
}

func ExampleParseURLBuilder() {
	b, err := vurl.ParseURLBuilder("https://example.com:8443/a%20b?q=go#top")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(b.Scheme(), b.Host(), b.Port(), b.PathString(), b.QueryString(), b.Fragment())
	// Output: https example.com 8443 /a%20b q=go top
}

func ExamplePath() {
	path, err := vurl.Path("https://example.com/a%20b/c")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(path)
	// Output: /a b/c
}

func ExampleSize() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "4")
	}))
	defer server.Close()

	n, err := vurl.Size(server.URL)
	fmt.Println(n, err == nil)
	// Output: 4 true
}

func ExampleSizeWithOptions() {
	n, err := vurl.SizeWithOptions("memory.txt", vurl.WithStat(exampleURLStat(9)))
	fmt.Println(n, err == nil)
	// Output: 9 true
}

func ExampleStringURI() {
	fmt.Println(vurl.StringURI("hello"))
	// Output: string:///hello
}

func ExampleToURI() {
	u, err := vurl.ToURI("https://example.com/a b", true)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(u.String())
	// Output: https://example.com/a%20b
}

func ExampleURLDecode() {
	decoded, err := vurl.URLDecode("a+b%26c")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(decoded)
	// Output: a b&c
}

func ExampleURLEncode() {
	fmt.Println(vurl.URLEncode("a b&c"))
	// Output: a+b%26c
}

func ExampleURLEncodeWithOptions() {
	out := vurl.URLEncodeWithOptions("go url", vurl.WithQueryEscapeFunc(strings.ToUpper))
	fmt.Println(out)
	// Output: GO URL
}

func ExampleWithAllowLocalFiles() {
	rc, err := vurl.OpenWithOptions("/tmp/secret.txt", vurl.WithAllowLocalFiles(false))
	if rc != nil {
		_ = rc.Close()
	}
	fmt.Println(err != nil)
	// Output: true
}

func ExampleWithAllowedHosts() {
	rc, err := vurl.OpenWithOptions("https://blocked.test", vurl.WithAllowedHosts("example.com"))
	if rc != nil {
		_ = rc.Close()
	}
	fmt.Println(err != nil)
	// Output: true
}

func ExampleWithAllowedSchemes() {
	rc, err := vurl.OpenWithOptions("ftp://example.com/file", vurl.WithAllowedSchemes("http", "https"))
	if rc != nil {
		_ = rc.Close()
	}
	fmt.Println(err != nil)
	// Output: true
}

func ExampleWithCheckStatus() {
	rc, err := vurl.OpenWithOptions(
		"https://example.com/missing",
		vurl.WithHTTPClient(exampleURLClient("missing", http.StatusNotFound)),
		vurl.WithCheckStatus(true),
	)
	if rc != nil {
		_ = rc.Close()
	}
	fmt.Println(err != nil)
	// Output: true
}

func ExampleWithContext() {
	ctx := context.WithValue(context.Background(), exampleURLContextKey{}, "trace-1")
	client := &http.Client{Transport: exampleURLRoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		fmt.Println(req.Context().Value(exampleURLContextKey{}))
		return exampleURLResponse("ok", http.StatusOK), nil
	})}
	rc, err := vurl.OpenWithOptions("https://example.com", vurl.WithContext(ctx), vurl.WithHTTPClient(client))
	if rc != nil {
		_ = rc.Close()
	}
	if err != nil {
		fmt.Println(err)
	}
	// Output: trace-1
}

func ExampleWithDefaultScheme() {
	fmt.Println(vurl.NormalizeWithOptions("example.com/docs", false, false, vurl.WithDefaultScheme("https")))
	// Output: https://example.com/docs
}

func ExampleWithEncodePath() {
	fmt.Println(vurl.NormalizeUsingOptions("example.com/a b", vurl.WithEncodePath(true)))
	// Output: http://example.com/a%20b
}

func ExampleWithHTTPClient() {
	rc, err := vurl.OpenWithOptions(
		"https://example.com/data",
		vurl.WithHTTPClient(exampleURLClient("client", http.StatusOK)),
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() { _ = rc.Close() }()
	data, _ := io.ReadAll(rc)
	fmt.Println(string(data))
	// Output: client
}

func ExampleWithHeader() {
	client := &http.Client{Transport: exampleURLRoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		fmt.Println(req.Header.Get("X-Trace"))
		return exampleURLResponse("ok", http.StatusOK), nil
	})}
	rc, err := vurl.OpenWithOptions(
		"https://example.com/data",
		vurl.WithHTTPClient(client),
		vurl.WithHeader("X-Trace", "one"),
	)
	if rc != nil {
		_ = rc.Close()
	}
	if err != nil {
		fmt.Println(err)
	}
	// Output: one
}

func ExampleWithHeaders() {
	client := &http.Client{Transport: exampleURLRoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		fmt.Println(req.Header.Get("X-Trace"), req.Header.Get("X-Mode"))
		return exampleURLResponse("ok", http.StatusOK), nil
	})}
	rc, err := vurl.OpenWithOptions(
		"https://example.com/data",
		vurl.WithHTTPClient(client),
		vurl.WithHeaders(http.Header{"X-Trace": {"one"}, "X-Mode": {"test"}}),
	)
	if rc != nil {
		_ = rc.Close()
	}
	if err != nil {
		fmt.Println(err)
	}
	// Output: one test
}

func ExampleWithLookupIP() {
	rc, err := vurl.OpenSafeWithOptions(
		"https://example.com/data",
		vurl.WithLookupIP(func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		}),
	)
	if rc != nil {
		_ = rc.Close()
	}
	fmt.Println(err != nil)
	// Output: true
}

func ExampleWithMaxBytes() {
	client := &http.Client{Transport: exampleURLRoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		resp := exampleURLResponse("hello", http.StatusOK)
		resp.ContentLength = -1
		return resp, nil
	})}
	rc, err := vurl.OpenWithOptions(
		"https://example.com/data",
		vurl.WithHTTPClient(client),
		vurl.WithMaxBytes(3),
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() { _ = rc.Close() }()
	data, readErr := io.ReadAll(rc)
	fmt.Println(string(data), readErr != nil)
	// Output: hel true
}

func ExampleWithOpenFile() {
	rc, err := vurl.OpenWithOptions("memory.txt", vurl.WithOpenFile(exampleURLOpenFile("opened")))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() { _ = rc.Close() }()
	data, _ := io.ReadAll(rc)
	fmt.Println(string(data))
	// Output: opened
}

func ExampleWithPathEscapeFunc() {
	out := vurl.EncodePathSegmentWithOptions("a b", vurl.WithPathEscapeFunc(func(s string) string {
		return "path:" + s
	}))
	fmt.Println(out)
	// Output: path:a b
}

func ExampleWithPlusAsSpace() {
	decoded, err := vurl.DecodeWithOptions("a+b", vurl.WithPlusAsSpace(false))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(decoded)
	// Output: a+b
}

func ExampleWithQueryEscapeFunc() {
	out := vurl.EncodeWithOptions("a b", vurl.WithQueryEscapeFunc(func(s string) string {
		return "query:" + s
	}))
	fmt.Println(out)
	// Output: query:a b
}

func ExampleWithRejectPrivateHosts() {
	rc, err := vurl.OpenWithOptions("http://localhost/data", vurl.WithRejectPrivateHosts(true))
	if rc != nil {
		_ = rc.Close()
	}
	fmt.Println(err != nil)
	// Output: true
}

func ExampleWithReplaceSlash() {
	fmt.Println(vurl.NormalizeUsingOptions("example.com//docs", vurl.WithReplaceSlash(true)))
	// Output: http://example.com/docs
}

func ExampleWithRequestFactory() {
	client := &http.Client{Transport: exampleURLRoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		fmt.Println(req.URL.Path)
		return exampleURLResponse("ok", http.StatusOK), nil
	})}
	factory := func(ctx context.Context, method, raw string) (*http.Request, error) {
		return http.NewRequestWithContext(ctx, method, "https://example.com/factory", nil)
	}
	rc, err := vurl.OpenWithOptions(
		"https://example.com/original",
		vurl.WithHTTPClient(client),
		vurl.WithRequestFactory(factory),
	)
	if rc != nil {
		_ = rc.Close()
	}
	if err != nil {
		fmt.Println(err)
	}
	// Output: /factory
}

func ExampleWithStat() {
	n, err := vurl.ContentLengthWithOptions("memory.txt", vurl.WithStat(exampleURLStat(6)))
	fmt.Println(n, err == nil)
	// Output: 6 true
}

func ExampleWithTimeout() {
	client := &http.Client{Transport: exampleURLRoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		_, ok := req.Context().Deadline()
		fmt.Println(ok)
		return exampleURLResponse("ok", http.StatusOK), nil
	})}
	rc, err := vurl.OpenWithOptions(
		"https://example.com/data",
		vurl.WithHTTPClient(client),
		vurl.WithTimeout(time.Second),
	)
	if rc != nil {
		_ = rc.Close()
	}
	if err != nil {
		fmt.Println(err)
	}
	// Output: true
}

type exampleURLRoundTripperFunc func(*http.Request) (*http.Response, error)

func (f exampleURLRoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type exampleURLContextKey struct{}

type exampleURLFileInfo struct {
	size int64
}

func (f exampleURLFileInfo) Name() string       { return "memory.txt" }
func (f exampleURLFileInfo) Size() int64        { return f.size }
func (f exampleURLFileInfo) Mode() fs.FileMode  { return 0o644 }
func (f exampleURLFileInfo) ModTime() time.Time { return time.Unix(0, 0) }
func (f exampleURLFileInfo) IsDir() bool        { return false }
func (f exampleURLFileInfo) Sys() any           { return nil }

func exampleURLClient(body string, status int) *http.Client {
	return &http.Client{Transport: exampleURLRoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return exampleURLResponse(body, status), nil
	})}
}

func exampleURLResponse(body string, status int) *http.Response {
	return &http.Response{
		StatusCode:    status,
		Status:        fmt.Sprintf("%d %s", status, http.StatusText(status)),
		Header:        make(http.Header),
		Body:          io.NopCloser(strings.NewReader(body)),
		ContentLength: int64(len(body)),
	}
}

func exampleURLOpenFile(content string) func(string) (io.ReadCloser, error) {
	return func(string) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(content)), nil
	}
}

func exampleURLStat(size int64) func(string) (fs.FileInfo, error) {
	return func(string) (fs.FileInfo, error) {
		return exampleURLFileInfo{size: size}, nil
	}
}

func examplePublicLookupIP(context.Context, string) ([]net.IP, error) {
	return []net.IP{net.ParseIP("93.184.216.34")}, nil
}
