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

func TestReadChunks(t *testing.T) {
	var chunks []string
	if err := ReadChunks(strings.NewReader("xy"), func(chunk []byte) error {
		chunks = append(chunks, string(chunk))
		return nil
	}); err != nil {
		t.Fatalf("ReadChunks: %v", err)
	}
	if strings.Join(chunks, "") != "xy" {
		t.Fatalf("ReadChunks chunks = %#v", chunks)
	}

	chunks = nil
	err := ReadChunksWithOptions(strings.NewReader("abcdef"), func(chunk []byte) error {
		chunks = append(chunks, string(chunk))
		return nil
	}, WithBufferSize(2), WithMaxBytes(6))
	if err != nil {
		t.Fatalf("ReadChunksWithOptions: %v", err)
	}
	if strings.Join(chunks, "|") != "ab|cd|ef" {
		t.Fatalf("ReadChunksWithOptions chunks = %#v", chunks)
	}

	if err := ReadChunksWithOptions(strings.NewReader("abcd"), func([]byte) error { return nil }, WithBufferSize(2), WithMaxBytes(3)); !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("ReadChunksWithOptions over limit error = %v", err)
	}
	if err := ReadChunksWithOptions(strings.NewReader("x"), nil); !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("ReadChunksWithOptions nil handler error = %v", err)
	}
	if err := ReadChunksWithOptions(nil, func([]byte) error { return nil }); !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("ReadChunksWithOptions nil reader error = %v", err)
	}
}

func TestFileReadChunksWithProvider(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "chunks.txt")
	if err := FileWriteString(path, "xyz"); err != nil {
		t.Fatalf("FileWriteString: %v", err)
	}
	if got, err := FileReadBytes(path); err != nil || string(got) != "xyz" {
		t.Fatalf("FileReadBytes = %q, %v", got, err)
	}
	var fileChunks []string
	if err := FileReadChunks(path, func(chunk []byte) error {
		fileChunks = append(fileChunks, string(chunk))
		return nil
	}); err != nil {
		t.Fatalf("FileReadChunks: %v", err)
	}
	if strings.Join(fileChunks, "") != "xyz" {
		t.Fatalf("FileReadChunks chunks = %#v", fileChunks)
	}

	opened := ""
	open := func(path string) (io.ReadCloser, error) {
		opened = path
		return io.NopCloser(strings.NewReader("abcde")), nil
	}
	var chunks []string
	err := FileReadChunksWithOptions("virtual-chunks.txt", func(chunk []byte) error {
		chunks = append(chunks, string(chunk))
		return nil
	}, WithOpen(open), WithBufferSize(3))
	if err != nil {
		t.Fatalf("FileReadChunksWithOptions: %v", err)
	}
	if opened != "virtual-chunks.txt" || strings.Join(chunks, "|") != "abc|de" {
		t.Fatalf("FileReadChunksWithOptions opened=%q chunks=%#v", opened, chunks)
	}
	if err := FileReadChunksWithOptions("virtual-chunks.txt", nil, WithOpen(open)); !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("FileReadChunksWithOptions nil handler error = %v", err)
	}
	if err := FileReadChunksWithOptions("missing.txt", func([]byte) error { return nil }, WithOpen(func(string) (io.ReadCloser, error) {
		return nil, os.ErrNotExist
	})); !errors.Is(err, knifer.ErrCodeNotFound) {
		t.Fatalf("FileReadChunksWithOptions missing error = %v", err)
	}
}
