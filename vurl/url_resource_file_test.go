package vurl_test

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/imajinyun/knifer-go/vurl"
)

func TestFacadeTrustedFileResourceWrappers(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "trusted.txt")
	if err := os.WriteFile(file, []byte("trusted"), 0o600); err != nil {
		t.Fatal(err)
	}
	rc, err := vurl.Open(file)
	if err != nil {
		t.Fatalf("Open file: %v", err)
	}
	data, err := io.ReadAll(rc)
	_ = rc.Close()
	if err != nil || string(data) != "trusted" {
		t.Fatalf("Open file data = %q, %v", data, err)
	}
	if length, err := vurl.ContentLength(file); err != nil || length != int64(len("trusted")) {
		t.Fatalf("ContentLength = %d, %v", length, err)
	}
	if size, err := vurl.Size(file); err != nil || size != int64(len("trusted")) {
		t.Fatalf("Size = %d, %v", size, err)
	}
	if _, err := vurl.OpenWithOptions(file, vurl.WithAllowedSchemes("http")); err == nil {
		t.Fatal("OpenWithOptions disallowed scheme error = nil")
	}
	if _, err := vurl.OpenWithOptions(file, vurl.WithAllowLocalFiles(false)); err == nil {
		t.Fatal("OpenWithOptions local file disabled error = nil")
	}
	if _, err := vurl.ContentLengthWithOptions(file, vurl.WithAllowLocalFiles(false)); err == nil {
		t.Fatal("ContentLengthWithOptions local file disabled error = nil")
	}
}
