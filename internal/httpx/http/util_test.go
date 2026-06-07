package http

import (
	"regexp"
	"strings"
	"testing"
)

func TestIsHTTP(t *testing.T) {
	if !IsHTTP("Http://aaa.bbb") {
		t.Fatal("Http://")
	}
	if !IsHTTP("HTTP://aaa.bbb") {
		t.Fatal("HTTP://")
	}
	if IsHTTP("FTP://aaa.bbb") {
		t.Fatal("FTP://")
	}
}

func TestIsHTTPS(t *testing.T) {
	if !IsHTTPS("Https://aaa.bbb") {
		t.Fatal("Https://")
	}
	if !IsHTTPS("HTTPS://aaa.bbb") {
		t.Fatal("HTTPS://")
	}
	if !IsHTTPS("https://aaa.bbb") {
		t.Fatal("https://")
	}
	if IsHTTPS("ftp://aaa.bbb") {
		t.Fatal("ftp://")
	}
}

func TestDecodeParams(t *testing.T) {
	paramsStr := "uuuu=0&a=b&c=%3F%23%40!%24%25%5E%26%3Ddsssss555555"
	m := DecodeParams(paramsStr)
	if m["uuuu"][0] != "0" {
		t.Fatalf("uuuu: %v", m["uuuu"])
	}
	if m["a"][0] != "b" {
		t.Fatalf("a: %v", m["a"])
	}
	if m["c"][0] != "?#@!$%^&=dsssss555555" {
		t.Fatalf("c: %v", m["c"])
	}
}

func TestDecodeParamMap(t *testing.T) {
	m := DecodeParamMap("aa=123&f_token=NzBkMjQxNDM1MDVlMDliZTk1OTU3ZDI1OTI0NTBiOWQ=")
	if m["aa"] != "123" {
		t.Fatalf("aa: %q", m["aa"])
	}
	if m["f_token"] != "NzBkMjQxNDM1MDVlMDliZTk1OTU3ZDI1OTI0NTBiOWQ=" {
		t.Fatalf("f_token: %q", m["f_token"])
	}
}

func TestEncodeParams(t *testing.T) {
	got := EncodeParams("http://www.abc.dd?a=b&c=d")
	if !strings.Contains(got, "a=b") || !strings.Contains(got, "c=d") {
		t.Fatalf("encoded: %q", got)
	}
	if EncodeParams("https://www.example.com/") != "https://www.example.com/" {
		t.Fatal("URL without query should be returned unchanged")
	}
}

func TestToParams(t *testing.T) {
	m := map[string]any{"a": "1"}
	got := ToParams(m)
	if got != "a=1" {
		t.Fatalf("ToParams: %q", got)
	}
}

func TestURLWithFormFunc(t *testing.T) {
	got := URLWithForm("http://api.gokit.cn/login", map[string]any{"a": 1})
	if !strings.Contains(got, "?a=1") {
		t.Fatalf("URLWithForm: %q", got)
	}
	got2 := URLWithForm("http://api.gokit.cn/login?type=aaa", map[string]any{"x": "y"})
	if !strings.Contains(got2, "type=aaa") || !strings.Contains(got2, "x=y") || !strings.Contains(got2, "&") {
		t.Fatalf("URLWithForm2: %q", got2)
	}
}

func TestGetCharset(t *testing.T) {
	if got := GetCharsetFromContentType("Charset=UTF-8;fq=0.9"); got != "UTF-8" {
		t.Fatalf("charset: %q", got)
	}
	if got := GetCharsetFromHTML("<meta charset=utf-8"); got != "utf-8" {
		t.Fatalf("html charset: %q", got)
	}
	if got := GetCharsetFromHTML(`<meta charset='utf-8'`); got != "utf-8" {
		t.Fatalf("html charset2: %q", got)
	}
	if got := GetCharsetFromHTML(`<meta charset="utf-8"`); got != "utf-8" {
		t.Fatalf("html charset3: %q", got)
	}
	if got := GetCharsetFromHTML(`<meta charset = "utf-8"`); got != "utf-8" {
		t.Fatalf("html charset4: %q", got)
	}
	if got := GetCharsetFromContentTypeWithOptions("encoding=UTF-16", WithCharsetRegexp(regexp.MustCompile(`encoding=([^;]+)`))); got != "UTF-16" {
		t.Fatalf("charset with options: %q", got)
	}
	if got := GetCharsetFromHTMLWithOptions(`<html data-charset="gbk">`, WithMetaCharsetRegexp(regexp.MustCompile(`data-charset="([^"]+)"`))); got != "gbk" {
		t.Fatalf("html charset with options: %q", got)
	}
}

func TestGetMimeType(t *testing.T) {
	if got := GetMimeType("aaa.aaa"); got != "" {
		t.Fatalf("mime: %q", got)
	}
	if got := GetMimeType("a.json"); got != "application/json" {
		t.Fatalf("mime json: %q", got)
	}
	if got := GetMimeType("a.png"); got != "image/png" {
		t.Fatalf("mime png: %q", got)
	}
}

func TestBuildBasicAuth(t *testing.T) {
	if got := BuildBasicAuth("aladdin", "opensesame"); got != "Basic YWxhZGRpbjpvcGVuc2VzYW1l" {
		t.Fatalf("auth: %q", got)
	}
}

func TestCreateRequest(t *testing.T) {
	req := CreateRequest(MethodPut, "http://example.com")
	if req.method != MethodPut {
		t.Fatalf("method: %v", req.method)
	}
}

func TestCreateGetWithFollowRedirects(t *testing.T) {
	req := CreateGet("http://example.com", false)
	if req.followRedir == nil || *req.followRedir != false {
		t.Fatalf("followRedir: %v", req.followRedir)
	}
}

func TestCreateWithOptionsAppliesRequestOptions(t *testing.T) {
	getReq := CreateGetWithOptions("http://example.com", false, WithHeader("X-Create", "get"), WithUserAgent("create-get-agent"))
	if getReq.followRedir == nil || *getReq.followRedir {
		t.Fatalf("followRedir: %v", getReq.followRedir)
	}
	if got := getReq.headers.Get("X-Create"); got != "get" {
		t.Fatalf("CreateGetWithOptions header = %q, want get", got)
	}
	if got := getReq.userAgent; got != "create-get-agent" {
		t.Fatalf("CreateGetWithOptions userAgent = %q", got)
	}

	postReq := CreatePostWithOptions("http://example.com", WithHeader("X-Create", "post"))
	if postReq.method != MethodPost {
		t.Fatalf("CreatePostWithOptions method = %v, want POST", postReq.method)
	}
	if got := postReq.headers.Get("X-Create"); got != "post" {
		t.Fatalf("CreatePostWithOptions header = %q, want post", got)
	}
}
