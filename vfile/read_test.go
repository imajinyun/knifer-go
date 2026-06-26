package vfile

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestFacadeReadOptions(t *testing.T) {
	if got, err := ReadStringWithOptions(ReaderFromString("abc"), WithMaxBytes(3)); err != nil || got != "abc" {
		t.Fatalf("ReadStringWithOptions() = %q, %v; want abc, nil", got, err)
	}
	if _, err := ReadStringWithOptions(ReaderFromString("abcd"), WithMaxBytes(3)); err == nil {
		t.Fatal("ReadStringWithOptions() over limit error = nil")
	}
	lines, err := ReadLinesWithOptions(ReaderFromString("abc"), WithInitialLineBuffer(1), WithMaxLineBytes(4))
	if err != nil {
		t.Fatalf("ReadLinesWithOptions() error = %v", err)
	}
	if len(lines) != 1 || lines[0] != "abc" {
		t.Fatalf("ReadLinesWithOptions() = %v, want [abc]", lines)
	}
}

func TestFacadeAdditionalReadWrappers(t *testing.T) {
	if got, err := ReadAll(ReaderFromString("all")); err != nil || string(got) != "all" {
		t.Fatalf("ReadAll = %q, %v", got, err)
	}
	if got, err := ReadAllWithOptions(ReaderFromString("abcd"), WithMaxBytes(1), WithUnlimitedRead()); err != nil || string(got) != "abcd" {
		t.Fatalf("ReadAllWithOptions unlimited = %q, %v", got, err)
	}
	if got, err := ReadLines(ReaderFromString("a\nb\n")); err != nil || len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("ReadLines = %v, %v", got, err)
	}
}

func TestFacadeChunkReadAndLimitedCopy(t *testing.T) {
	var chunks []string
	if err := ReadChunks(ReaderFromString("abcd"), func(chunk []byte) error {
		chunks = append(chunks, string(chunk))
		return nil
	}); err != nil {
		t.Fatalf("ReadChunks: %v", err)
	}
	if strings.Join(chunks, "") != "abcd" {
		t.Fatalf("ReadChunks chunks = %#v", chunks)
	}

	chunks = nil
	if err := ReadChunksWithOptions(ReaderFromString("abcdef"), func(chunk []byte) error {
		chunks = append(chunks, string(chunk))
		return nil
	}, WithBufferSize(2)); err != nil {
		t.Fatalf("ReadChunksWithOptions: %v", err)
	}
	if strings.Join(chunks, "|") != "ab|cd|ef" {
		t.Fatalf("ReadChunksWithOptions chunks = %#v", chunks)
	}

	var dst bytes.Buffer
	n, err := CopyWithOptions(&dst, ReaderFromString("abcdef"), WithBufferSize(2), WithMaxBytes(3))
	if !errors.Is(err, knifer.ErrCodeInvalidInput) || n != 3 || dst.String() != "abc" {
		t.Fatalf("CopyWithOptions limited n=%d dst=%q err=%v", n, dst.String(), err)
	}
}

func TestFacadeFileChunkWrappers(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/chunks.txt"
	if err := WriteFileString(path, "abcde"); err != nil {
		t.Fatalf("WriteFileString: %v", err)
	}
	var chunks []string
	if err := ReadFileChunks(path, func(chunk []byte) error {
		chunks = append(chunks, string(chunk))
		return nil
	}); err != nil {
		t.Fatalf("ReadFileChunks: %v", err)
	}
	if strings.Join(chunks, "") != "abcde" {
		t.Fatalf("ReadFileChunks chunks = %#v", chunks)
	}

	chunks = nil
	if err := ReadFileChunksWithOptions(path, func(chunk []byte) error {
		chunks = append(chunks, string(chunk))
		return nil
	}, WithBufferSize(2)); err != nil {
		t.Fatalf("ReadFileChunksWithOptions: %v", err)
	}
	if strings.Join(chunks, "|") != "ab|cd|e" {
		t.Fatalf("ReadFileChunksWithOptions chunks = %#v", chunks)
	}
}
