package vfile

import (
	"io"
	"io/fs"
	"os"
	"strings"
	"testing"
)

func TestFacadeWriteAndDirOptions(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/private/a.txt"
	if err := WriteFileString(path, "first", WithCreateParents(false)); err == nil {
		t.Fatal("WriteFileString() should fail when parent directory is missing and WithCreateParents(false) is set")
	}
	if err := Mkdir(dir+"/private", WithMkdirPerm(0o700)); err != nil {
		t.Fatalf("Mkdir() with options error = %v", err)
	}
	dirInfo, err := os.Stat(dir + "/private")
	if err != nil {
		t.Fatalf("stat private dir: %v", err)
	}
	if got := dirInfo.Mode().Perm(); got != 0o700 {
		t.Fatalf("private dir perm = %o, want 700", got)
	}
	if err := WriteFileString(path, "first", WithFilePerm(0o600)); err != nil {
		t.Fatalf("WriteFileString() with options error = %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat file: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("file perm = %o, want 600", got)
	}
	if err := WriteFileBytes(path, []byte("second"), WithOverwrite(false)); err == nil {
		t.Fatal("WriteFileBytes() should reject overwrite=false for existing file")
	}
	if err := AppendFileString(path, "!"); err != nil {
		t.Fatalf("AppendFileString() error = %v", err)
	}
	copyPath := dir + "/copy.txt"
	if err := CopyFile(path, copyPath, WithFilePerm(0o600)); err != nil {
		t.Fatalf("CopyFile() with options error = %v", err)
	}
	if err := CopyFile(path, copyPath, WithOverwrite(false)); err == nil {
		t.Fatal("CopyFile() should reject overwrite=false for existing destination")
	}
	if err := Touch(dir+"/touch.txt", WithFilePerm(0o600)); err != nil {
		t.Fatalf("Touch() with options error = %v", err)
	}
}

func TestFacadeWriteProviderOptions(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/nested/out.txt"
	mkdirCalled := false
	openCalled := false
	if err := WriteFileBytes(path, []byte("provider"),
		WithDirPerm(0o700),
		WithMkdirAll(func(p string, perm fs.FileMode) error {
			mkdirCalled = strings.HasSuffix(p, "/nested") && perm == 0o700
			return os.MkdirAll(p, perm)
		}),
		WithOpenFile(func(p string, flag int, perm fs.FileMode) (io.WriteCloser, error) {
			openCalled = p == path && flag&os.O_CREATE != 0 && perm == 0o644
			return os.OpenFile(p, flag, perm)
		}),
	); err != nil {
		t.Fatalf("WriteFileBytes provider options: %v", err)
	}
	if !mkdirCalled || !openCalled {
		t.Fatalf("provider called mkdir=%v open=%v", mkdirCalled, openCalled)
	}
	if got, err := ReadFileString(path); err != nil || got != "provider" {
		t.Fatalf("ReadFileString provider file = %q, %v", got, err)
	}
}
