package vfile

import (
	"io"
	"io/fs"
	"os"
	"strings"
	"testing"
)

func TestFacadeAdditionalStatDeleteAndCopyEdges(t *testing.T) {
	if IsFileWithOptions("dir", WithStat(func(string) (fs.FileInfo, error) {
		return fakeDirFacadeFileInfo{name: "dir"}, nil
	})) {
		t.Fatal("IsFileWithOptions directory = true")
	}
	if !IsDirectoryWithOptions("dir", WithStat(func(string) (fs.FileInfo, error) {
		return fakeDirFacadeFileInfo{name: "dir"}, nil
	})) {
		t.Fatal("IsDirectoryWithOptions directory = false")
	}
	if got := SizeWithOptions("x", WithStat(func(string) (fs.FileInfo, error) {
		return fakeSizedFacadeFileInfo{name: "x", size: 42}, nil
	})); got != 42 {
		t.Fatalf("SizeWithOptions = %d", got)
	}
	if err := DelWithOptions("missing", WithStat(func(string) (fs.FileInfo, error) {
		return nil, os.ErrNotExist
	})); err != nil {
		t.Fatalf("DelWithOptions missing = %v", err)
	}
	dir := t.TempDir()
	src := dir + "/src.txt"
	if err := os.WriteFile(src, []byte("copy"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := CopyFileWithOptions(src, dir+"/virtual-dst",
		WithOpenFile(func(path string, flag int, perm fs.FileMode) (io.WriteCloser, error) {
			if !strings.HasSuffix(path, "/virtual-dst") || flag&os.O_CREATE == 0 || perm == 0 {
				t.Fatalf("copy openFile path=%q flag=%#x perm=%v", path, flag, perm)
			}
			return nopFacadeWriteCloser{Writer: io.Discard}, nil
		}),
	); err != nil {
		t.Fatalf("CopyFileWithOptions providers: %v", err)
	}
}
