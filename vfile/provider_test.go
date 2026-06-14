package vfile

import (
	"io"
	"io/fs"
	"strings"
	"testing"
)

func TestFacadeProviderOptions(t *testing.T) {
	if got, err := ReadFileStringWithOptions("virtual.txt", WithOpen(func(path string) (io.ReadCloser, error) {
		if path != "virtual.txt" {
			t.Fatalf("read path = %q, want virtual.txt", path)
		}
		return io.NopCloser(strings.NewReader("virtual")), nil
	})); err != nil || got != "virtual" {
		t.Fatalf("ReadFileStringWithOptions() = %q, %v", got, err)
	}
	if !ExistsWithOptions("x", WithStat(func(path string) (fs.FileInfo, error) {
		return fakeFacadeFileInfo{name: path}, nil
	})) {
		t.Fatal("ExistsWithOptions() = false, want true")
	}
	removed := false
	if err := DelWithOptions("x",
		WithStat(func(string) (fs.FileInfo, error) { return fakeFacadeFileInfo{name: "x"}, nil }),
		WithRemoveAll(func(string) error { removed = true; return nil }),
	); err != nil || !removed {
		t.Fatalf("DelWithOptions() = %v, removed=%v", err, removed)
	}
}
