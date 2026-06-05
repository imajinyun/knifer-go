package file

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	knifer "github.com/imajinyun/go-knifer"
)

// Tests cover the utility toolkit-core IoUtilTest, FileUtilTest, and FileNameUtilTest.

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

func TestFileWriteRead(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "a.txt")
	if err := FileWriteString(path, "你好"); err != nil {
		t.Fatalf("write: %v", err)
	}
	if !FileExists(path) || !IsFile(path) {
		t.Fatalf("FileExists/IsFile failed")
	}
	if FileSize(path) <= 0 {
		t.Fatalf("FileSize failed")
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

func TestMkdirTouchDel(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "x", "y")
	if err := Mkdir(sub); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if !IsDirectory(sub) {
		t.Fatalf("IsDirectory failed")
	}
	f := filepath.Join(sub, "a.txt")
	if err := Touch(f); err != nil {
		t.Fatalf("touch: %v", err)
	}
	if !IsFile(f) {
		t.Fatalf("touch did not create file")
	}
	if err := Del(f); err != nil {
		t.Fatalf("del: %v", err)
	}
	if FileExists(f) {
		t.Fatalf("Del did not remove")
	}
}

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

func TestMkdirAndCopyOptions(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "private")
	if err := Mkdir(sub, WithMkdirPerm(0o700)); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	info, err := os.Stat(sub)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o700 {
		t.Fatalf("dir perm = %o, want 700", got)
	}
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

func assertFileCode(t *testing.T, err error, code knifer.ErrCode) {
	t.Helper()
	if err == nil {
		t.Fatalf("err = nil, want %s", code)
	}
	if !errors.Is(err, code) {
		t.Fatalf("errors.Is(%v, %s) = false", err, code)
	}
	got, ok := knifer.CodeOf(err)
	if !ok || got != code {
		t.Fatalf("CodeOf(%v) = %q, %v; want %q, true", err, got, ok, code)
	}
}

func TestMainNameAndExtension(t *testing.T) {
	if MainName("/x/y/foo.txt") != "foo" {
		t.Fatalf("MainName failed")
	}
	if Extension("/x/y/foo.txt") != "txt" {
		t.Fatalf("Extension failed")
	}
	if MainName("foo") != "foo" || Extension("foo") != "" {
		t.Fatalf("no-ext failed")
	}
}
