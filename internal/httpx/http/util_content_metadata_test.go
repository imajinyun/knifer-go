package http

import (
	"regexp"
	"testing"
)

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
