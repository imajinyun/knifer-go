package vfile

import (
	"strings"
	"testing"
)

func TestFileFacade(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/a/b.txt"
	if err := WriteFileString(path, "line1\nline2"); err != nil {
		t.Fatal(err)
	}
	if !Exists(path) || !IsFile(path) || !IsDirectory(dir) {
		t.Fatal("file predicates failed")
	}
	if got, err := ReadFileString(path); err != nil || got != "line1\nline2" {
		t.Fatalf("ReadFileString = %q, %v", got, err)
	}
	if got, err := ReadFileBytes(path); err != nil || string(got) != "line1\nline2" {
		t.Fatalf("ReadFileBytes = %q, %v", got, err)
	}
	if lines, err := ReadFileLines(path); err != nil || len(lines) != 2 {
		t.Fatalf("ReadFileLines = %v, %v", lines, err)
	}
	if _, err := Copy(&strings.Builder{}, ReaderFromString("x")); err != nil {
		t.Fatal(err)
	}
	if MainName(path) != "b" || Extension(path) != "txt" || Size(path) <= 0 {
		t.Fatal("file name/size helpers failed")
	}
	copyPath := dir + "/copy.txt"
	if err := CopyFile(path, copyPath); err != nil || !Exists(copyPath) {
		t.Fatalf("CopyFile failed: %v", err)
	}
	if err := AppendFileString(copyPath, "!"); err != nil {
		t.Fatal(err)
	}
	if err := Touch(dir + "/touch.txt"); err != nil {
		t.Fatal(err)
	}
	if err := Del(dir + "/a"); err != nil || Exists(path) {
		t.Fatalf("Del failed: %v", err)
	}
	CloseQuietly(nil)
}
