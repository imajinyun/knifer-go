package zip

import (
	archivezip "archive/zip"
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestUnzipRejectsPathTraversal(t *testing.T) {
	for _, name := range []string{"../evil.txt", "/evil.txt"} {
		t.Run(name, func(t *testing.T) {
			tmp := t.TempDir()
			archive := filepath.Join(tmp, "bad.zip")
			buf := zipArchiveBuffer(t, name, []byte("bad"))
			if err := os.WriteFile(archive, buf.Bytes(), 0o644); err != nil {
				t.Fatal(err)
			}
			assertZipCode(t, UnzipTo(archive, filepath.Join(tmp, "dest")), knifer.ErrCodeInvalidInput)
		})
	}
}

func TestUnzipRejectsSymlinkEscape(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink behavior differs on windows")
	}

	tmp := t.TempDir()
	dest := filepath.Join(tmp, "dest")
	outside := filepath.Join(tmp, "outside")
	if err := os.MkdirAll(filepath.Join(dest, "link"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(dest, "link")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(dest, "link")); err != nil {
		t.Fatalf("create symlink: %v", err)
	}

	var buf bytes.Buffer
	zw := archivezip.NewWriter(&buf)
	w, err := zw.Create("link/evil.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("bad")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	r, err := archivezip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}

	assertZipCode(t, UnzipReaderTo(r, dest), knifer.ErrCodeInvalidInput)
	if _, err := os.Stat(filepath.Join(outside, "evil.txt")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("symlink escape wrote outside file, stat err=%v", err)
	}
}

func TestUnzipSymlinkArchiveEntryDoesNotCreateSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink behavior differs on windows")
	}

	tmp := t.TempDir()
	dest := filepath.Join(tmp, "dest")
	outside := filepath.Join(tmp, "outside.txt")
	var buf bytes.Buffer
	zw := archivezip.NewWriter(&buf)
	header := &archivezip.FileHeader{Name: "link"}
	header.SetMode(os.ModeSymlink | 0o777)
	w, err := zw.CreateHeader(header)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("../outside.txt")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	r, err := archivezip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}

	if err := UnzipReaderTo(r, dest); err != nil {
		t.Fatalf("UnzipReaderTo: %v", err)
	}
	info, err := os.Lstat(filepath.Join(dest, "link"))
	if err != nil {
		t.Fatalf("Lstat extracted link entry: %v", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Fatal("symlink archive entry created filesystem symlink")
	}
	if got, err := os.ReadFile(filepath.Join(dest, "link")); err != nil || string(got) != "../outside.txt" {
		t.Fatalf("extracted link entry = %q, %v", got, err)
	}
	if _, err := os.Stat(outside); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("symlink archive entry wrote outside file, stat err=%v", err)
	}
}

func zipArchiveBuffer(t *testing.T, name string, data []byte) bytes.Buffer {
	t.Helper()

	var buf bytes.Buffer
	zw := archivezip.NewWriter(&buf)
	w, err := zw.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf
}
