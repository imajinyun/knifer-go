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

func TestFileCopy(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "a.txt")
	dst := filepath.Join(dir, "sub", "b.txt")
	if err := FileWriteString(src, "hello"); err != nil {
		t.Fatalf("write src: %v", err)
	}
	if err := FileCopy(src, dst); err != nil {
		t.Fatalf("copy: %v", err)
	}
	got, _ := FileReadString(dst)
	if got != "hello" {
		t.Fatalf("copy content: %q", got)
	}
	missingErr := FileCopy(filepath.Join(dir, "missing.txt"), filepath.Join(dir, "unused.txt"))
	assertFileCode(t, missingErr, knifer.ErrCodeInvalidInput)
}

func TestMkdirAndCopyOptionsCopy(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")
	if err := FileWriteString(src, "src"); err != nil {
		t.Fatalf("write src: %v", err)
	}
	if err := FileWriteString(dst, "dst"); err != nil {
		t.Fatalf("write dst: %v", err)
	}
	if err := FileCopy(src, dst, WithOverwrite(false)); !errors.Is(err, os.ErrExist) {
		t.Fatalf("copy overwrite false error = %v, want exists", err)
	}
	if err := FileCopy(src, dst, WithOverwrite(false)); !errors.Is(err, knifer.ErrCodeInternal) {
		t.Fatalf("copy overwrite false code = %v, want internal", err)
	}
}

func TestFileCopyProviderOptions(t *testing.T) {
	var copied bytes.Buffer
	if err := FileCopy("src.txt", "dst/out.txt",
		WithStat(func(path string) (fs.FileInfo, error) {
			if path == "src.txt" {
				return fakeFileInfo{name: path, size: 3}, nil
			}
			return nil, os.ErrNotExist
		}),
		WithMkdirAll(func(string, fs.FileMode) error { return nil }),
		WithOpen(func(path string) (io.ReadCloser, error) {
			if path != "src.txt" {
				t.Fatalf("open path = %q, want src.txt", path)
			}
			return io.NopCloser(strings.NewReader("abc")), nil
		}),
		WithOpenFile(func(path string, flag int, perm fs.FileMode) (io.WriteCloser, error) {
			if path != "dst/out.txt" {
				t.Fatalf("open destination = %q, want dst/out.txt", path)
			}
			return &bufferWriteCloser{Buffer: &copied}, nil
		}),
	); err != nil {
		t.Fatalf("FileCopy() with providers error = %v", err)
	}
	if got := copied.String(); got != "abc" {
		t.Fatalf("copied content = %q, want abc", got)
	}
}

func TestFileCopyReturnsDestinationCloseError(t *testing.T) {
	closeErr := errors.New("close failed")
	err := FileCopy("src.txt", "dst.txt",
		WithStat(func(path string) (fs.FileInfo, error) {
			if path == "src.txt" {
				return fakeFileInfo{name: path, size: 3}, nil
			}
			return nil, os.ErrNotExist
		}),
		WithOpen(func(string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("abc")), nil
		}),
		WithOpenFile(func(string, int, fs.FileMode) (io.WriteCloser, error) {
			return closeErrorWriteCloser{Writer: io.Discard, err: closeErr}, nil
		}),
	)
	assertFileCode(t, err, knifer.ErrCodeInternal)
	if !errors.Is(err, closeErr) {
		t.Fatalf("FileCopy close error = %v, want close cause", err)
	}
}
