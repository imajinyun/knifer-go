package shared

import (
	"errors"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestCharsetMimeAuthAndDispositionHelpers(t *testing.T) {
	if got := BuildBasicAuth("aladdin", "opensesame"); got != "Basic YWxhZGRpbjpvcGVuc2VzYW1l" {
		t.Fatalf("BuildBasicAuth = %q", got)
	}
	if got := GetCharsetFromContentType("text/plain; Charset=UTF-16"); got != "UTF-16" {
		t.Fatalf("GetCharsetFromContentType = %q", got)
	}
	if got := GetCharsetFromContentTypeWithOptions("encoding=gbk", WithCharsetRegexp(regexp.MustCompile(`encoding=([^;]+)`))); got != "gbk" {
		t.Fatalf("GetCharsetFromContentTypeWithOptions = %q", got)
	}
	if got := GetCharsetFromContentTypeWithOptions("charset=utf-8", WithCharsetRegexp(nil)); got != "utf-8" {
		t.Fatalf("GetCharsetFromContentTypeWithOptions nil regexp = %q", got)
	}
	if got := GetCharsetFromHTML(`<html><meta charset="big5"></html>`); got != "big5" {
		t.Fatalf("GetCharsetFromHTML = %q", got)
	}
	if got := GetCharsetFromHTMLWithOptions(`<meta data-charset="shift-jis">`, WithMetaCharsetRegexp(regexp.MustCompile(`data-charset="([^"]+)"`))); got != "shift-jis" {
		t.Fatalf("GetCharsetFromHTMLWithOptions = %q", got)
	}
	if got := GetCharsetFromHTMLWithOptions(`<meta charset="utf-8">`, WithMetaCharsetRegexp(nil)); got != "utf-8" {
		t.Fatalf("GetCharsetFromHTMLWithOptions nil regexp = %q", got)
	}
	if got := GetMimeType("payload.JSON"); got != "application/json" {
		t.Fatalf("GetMimeType JSON = %q", got)
	}
	if got := GetMimeType("payload.unknown"); got != "" {
		t.Fatalf("GetMimeType unknown = %q", got)
	}
	if got := NormalizeEncoding(" GZip "); got != "gzip" {
		t.Fatalf("NormalizeEncoding = %q", got)
	}
	if got := FilenameFromContentDisposition(`attachment; filename="report.csv"; size=1`); got != "report.csv" {
		t.Fatalf("FilenameFromContentDisposition quoted = %q", got)
	}
	if got := FilenameFromContentDisposition("attachment"); got != "" {
		t.Fatalf("FilenameFromContentDisposition without filename = %q", got)
	}
}

func TestSafeDownloadedFilenameAndJoin(t *testing.T) {
	if got, err := SafeDownloadedFilename(" report.csv "); err != nil || got != "report.csv" {
		t.Fatalf("SafeDownloadedFilename valid = %q, %v", got, err)
	}
	if got, err := SafeDownloadedFilename(" "); err != nil || got != "" {
		t.Fatalf("SafeDownloadedFilename blank = %q, %v", got, err)
	}
	for _, name := range []string{"../escape.txt", "nested/file.txt", `nested\file.txt`, "/tmp/file.txt", "."} {
		t.Run(name, func(t *testing.T) {
			if _, err := SafeDownloadedFilename(name); !errors.Is(err, knifer.ErrCodeInvalidInput) {
				t.Fatalf("SafeDownloadedFilename(%q) error = %v", name, err)
			}
		})
	}

	dir := t.TempDir()
	joined, err := SafeJoinDownloadPath(dir, "file.txt")
	if err != nil {
		t.Fatalf("SafeJoinDownloadPath valid: %v", err)
	}
	if !strings.HasPrefix(joined, filepath.Clean(dir)+string(filepath.Separator)) || filepath.Base(joined) != "file.txt" {
		t.Fatalf("SafeJoinDownloadPath valid = %q", joined)
	}
	if _, err := SafeJoinDownloadPath(dir, "../escape.txt"); !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("SafeJoinDownloadPath escape error = %v", err)
	}
}
