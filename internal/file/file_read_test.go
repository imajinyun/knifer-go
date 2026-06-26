package file

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestFileLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "lines.txt")
	if err := FileWriteString(path, "x\ny\nz"); err != nil {
		t.Fatalf("write: %v", err)
	}
	lines, err := FileReadLines(path)
	if err != nil {
		t.Fatalf("read lines: %v", err)
	}
	if len(lines) != 3 || lines[2] != "z" {
		t.Fatalf("lines: %v", lines)
	}
	if _, err := FileReadLinesWithOptions(path, WithMaxBytes(2)); err == nil {
		t.Fatal("FileReadLinesWithOptions over limit error = nil")
	}
}

func TestFileReadProviderOptions(t *testing.T) {
	opened := ""
	open := func(path string) (io.ReadCloser, error) {
		opened = path
		return io.NopCloser(strings.NewReader("a\nb")), nil
	}
	got, err := FileReadStringWithOptions("virtual.txt", WithOpen(open), WithMaxBytes(3))
	if err != nil || got != "a\nb" {
		t.Fatalf("FileReadStringWithOptions() = %q, %v", got, err)
	}
	if opened != "virtual.txt" {
		t.Fatalf("open path = %q, want virtual.txt", opened)
	}
	lines, err := FileReadLinesWithOptions("virtual-lines.txt", WithOpen(open))
	if err != nil || len(lines) != 2 || lines[1] != "b" {
		t.Fatalf("FileReadLinesWithOptions() = %v, %v", lines, err)
	}
	if _, err := FileReadBytesWithOptions("too-large.txt", WithOpen(open), WithMaxBytes(2)); err == nil {
		t.Fatal("FileReadBytesWithOptions() over limit error = nil")
	}
	if _, err := FileReadStringWithOptions("ignored", WithOpen(func(string) (io.ReadCloser, error) {
		return nil, os.ErrNotExist
	})); !errors.Is(err, knifer.ErrCodeNotFound) {
		t.Fatalf("FileReadStringWithOptions() missing code = %v, want not found", err)
	}
}

func TestReadAll(t *testing.T) {
	if got, err := ReadAll(strings.NewReader("all")); err != nil || string(got) != "all" {
		t.Fatalf("ReadAll = %q, %v", got, err)
	}
}
