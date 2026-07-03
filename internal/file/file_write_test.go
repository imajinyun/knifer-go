package file

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestFileWriteRead(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "a.txt")
	if err := FileWriteString(path, "你好"); err != nil {
		t.Fatalf("write: %v", err)
	}
	got, err := FileReadString(path)
	if err != nil || got != "你好" {
		t.Fatalf("read: %v %q", err, got)
	}
	if err := FileAppendString(path, "X"); err != nil {
		t.Fatalf("append: %v", err)
	}
	got, _ = FileReadString(path)
	if !strings.HasSuffix(got, "X") {
		t.Fatalf("append result: %q", got)
	}
}

func TestWriteOptions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.txt")
	if err := FileWriteString(path, "first", WithFilePerm(0o600)); err != nil {
		t.Fatalf("write with options: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("perm = %o, want 600", got)
	}
	if err := FileWriteString(path, "second", WithOverwrite(false)); !errors.Is(err, os.ErrExist) {
		t.Fatalf("overwrite false error = %v, want exists", err)
	}
	if err := FileWriteString(path, "second", WithOverwrite(false)); !errors.Is(err, knifer.ErrCodeInternal) {
		t.Fatalf("overwrite false code = %v, want internal", err)
	}
	missingParent := filepath.Join(dir, "missing", "b.txt")
	assertFileCode(t, FileWriteString(missingParent, "x", WithCreateParents(false)), knifer.ErrCodeNotFound)
}

func TestFileWriteProviderOptions(t *testing.T) {
	var mkdirPath string
	var mkdirPerm fs.FileMode
	var openPath string
	var openFlag int
	var openPerm fs.FileMode
	buf := &bytes.Buffer{}
	closer := &bufferWriteCloser{Buffer: buf}
	err := FileWriteBytes("parent/out.txt", []byte("payload"),
		WithDirPerm(0o700),
		WithFilePerm(0o600),
		WithMkdirAll(func(path string, perm fs.FileMode) error {
			mkdirPath = path
			mkdirPerm = perm
			return nil
		}),
		WithOpenFile(func(path string, flag int, perm fs.FileMode) (io.WriteCloser, error) {
			openPath = path
			openFlag = flag
			openPerm = perm
			return closer, nil
		}),
	)
	if err != nil {
		t.Fatalf("FileWriteBytes() error = %v", err)
	}
	if mkdirPath != "parent" || mkdirPerm != 0o700 {
		t.Fatalf("mkdir = %q/%o, want parent/700", mkdirPath, mkdirPerm)
	}
	if openPath != "parent/out.txt" || openPerm != 0o600 || openFlag&os.O_TRUNC == 0 {
		t.Fatalf("open = %q/%o/%#x", openPath, openPerm, openFlag)
	}
	if got := buf.String(); got != "payload" || !closer.closed {
		t.Fatalf("written = %q closed=%v, want payload/true", got, closer.closed)
	}
	if err := FileAppendString("append.txt", "!", WithCreateParents(false), WithOpenFile(func(string, int, fs.FileMode) (io.WriteCloser, error) {
		return &bufferWriteCloser{Buffer: buf}, nil
	})); err != nil {
		t.Fatalf("FileAppendString() with provider error = %v", err)
	}
}

func TestFileAppendStringReturnsCloseError(t *testing.T) {
	closeErr := errors.New("close failed")
	err := FileAppendString("append.txt", "payload",
		WithCreateParents(false),
		WithOpenFile(func(string, int, fs.FileMode) (io.WriteCloser, error) {
			return closeErrorWriteCloser{Writer: io.Discard, err: closeErr}, nil
		}),
	)
	assertFileCode(t, err, knifer.ErrCodeInternal)
	if !errors.Is(err, closeErr) {
		t.Fatalf("FileAppendString close error = %v, want close cause", err)
	}
}
