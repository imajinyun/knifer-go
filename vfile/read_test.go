package vfile

import "testing"

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
