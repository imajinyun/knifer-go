package vfile

import (
	"strings"
	"testing"
)

func TestFileTypeFacade(t *testing.T) {
	ft, err := DetectFileType(strings.NewReader("%PDF-1.7"))
	if err != nil {
		t.Fatalf("DetectFileType error = %v", err)
	}
	if !IsDocument(ft) || ft.Extension != ".pdf" {
		t.Fatalf("DetectFileType = %#v", ft)
	}
	if DetectFileTypeBytes([]byte("unknown")) != UnknownFileType {
		t.Fatal("unknown file type mismatch")
	}
}
