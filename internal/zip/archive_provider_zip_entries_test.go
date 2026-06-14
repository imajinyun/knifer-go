package zip

import (
	archivezip "archive/zip"
	"bytes"
	"io"
	"os"
	"testing"
)

func TestArchiveProviderOptionsForZipEntries(t *testing.T) {
	var mkdirPath string
	var mkdirPerm os.FileMode
	var openPath string
	var openFlag int
	var openPerm os.FileMode
	var buf bytes.Buffer
	closer := &zipBufferWriteCloser{Buffer: &buf}

	err := ZipEntriesWithOptions("parent/archive.zip", []EntryData{{Name: "a.txt", Data: []byte("a")}},
		WithDirPerm(0o700),
		WithFilePerm(0o600),
		WithMkdirAll(func(path string, perm os.FileMode) error {
			mkdirPath = path
			mkdirPerm = perm
			return nil
		}),
		WithOpenFile(func(path string, flag int, perm os.FileMode) (io.WriteCloser, error) {
			openPath = path
			openFlag = flag
			openPerm = perm
			return closer, nil
		}),
	)
	if err != nil {
		t.Fatalf("ZipEntriesWithOptions() error = %v", err)
	}
	if mkdirPath != "parent" || mkdirPerm != 0o700 {
		t.Fatalf("mkdir = %q/%o, want parent/700", mkdirPath, mkdirPerm)
	}
	if openPath != "parent/archive.zip" || openPerm != 0o600 || openFlag&os.O_TRUNC == 0 {
		t.Fatalf("open = %q/%o/%#x", openPath, openPerm, openFlag)
	}
	if !closer.closed {
		t.Fatal("archive output was not closed")
	}
	r, err := archivezip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("NewReader(provider output): %v", err)
	}
	if len(r.File) != 1 || r.File[0].Name != "a.txt" {
		t.Fatalf("entries = %#v, want a.txt", r.File)
	}
}
