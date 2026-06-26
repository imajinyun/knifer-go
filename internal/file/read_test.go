package file

import (
	"errors"
	"strings"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestReadString(t *testing.T) {
	r := ReaderFromString("hello world")
	got, err := ReadString(r)
	if err != nil || got != "hello world" {
		t.Fatalf("ReadString: %v %q", err, got)
	}
}

func TestReadLines(t *testing.T) {
	r := ReaderFromString("a\nb\nc")
	lines, err := ReadLines(r)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(lines) != 3 || lines[0] != "a" || lines[2] != "c" {
		t.Fatalf("ReadLines: %v", lines)
	}
}

func TestReadOptions(t *testing.T) {
	if got, err := ReadStringWithOptions(ReaderFromString("abc"), WithMaxBytes(3)); err != nil || got != "abc" {
		t.Fatalf("ReadStringWithOptions exact limit = %q, %v", got, err)
	}
	if _, err := ReadStringWithOptions(ReaderFromString("abcd"), WithMaxBytes(3)); err == nil {
		t.Fatal("ReadStringWithOptions over limit error = nil")
	}

	lines, err := ReadLinesWithOptions(ReaderFromString("abc"), WithMaxBytes(3), WithInitialLineBuffer(1), WithMaxLineBytes(4))
	if err != nil {
		t.Fatalf("ReadLinesWithOptions exact limit: %v", err)
	}
	if len(lines) != 1 || lines[0] != "abc" {
		t.Fatalf("ReadLinesWithOptions lines = %v", lines)
	}
	if _, err := ReadLinesWithOptions(ReaderFromString("abcd"), WithMaxBytes(3), WithMaxLineBytes(4)); err == nil {
		t.Fatal("ReadLinesWithOptions over limit error = nil")
	}
}

func TestReadOptionsDefaultLimitAndExplicitUnlimited(t *testing.T) {
	if _, err := ReadStringWithOptions(ReaderFromString(strings.Repeat("x", int(DefaultMaxBytes)+1))); !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("ReadStringWithOptions default limit error = %v, want invalid input", err)
	}
	got, err := ReadStringWithOptions(ReaderFromString("abcd"), WithMaxBytes(3), WithUnlimitedRead())
	if err != nil || got != "abcd" {
		t.Fatalf("WithUnlimitedRead() = %q, %v", got, err)
	}
}
