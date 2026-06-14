package zip

import (
	archivezip "archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestArchiveProviderOptionsForReadAndExtract(t *testing.T) {
	tmp := t.TempDir()
	archive := filepath.Join(tmp, "entries.zip")
	if err := ZipEntries(archive, EntryData{Name: "a.txt", Data: []byte("a")}); err != nil {
		t.Fatalf("ZipEntries() error = %v", err)
	}
	opened := ""
	openZip := func(path string) (*archivezip.ReadCloser, error) {
		opened = path
		return archivezip.OpenReader(path)
	}
	data, err := GetBytesWithOptions(archive, "a.txt", WithOpenZipReader(openZip))
	if err != nil || string(data) != "a" || opened != archive {
		t.Fatalf("GetBytesWithOptions() = %q, %v, opened=%q", data, err, opened)
	}
	names, err := ListFileNamesWithOptions(archive, "", WithOpenZipReader(openZip))
	if err != nil || !reflect.DeepEqual(names, []string{"a.txt"}) {
		t.Fatalf("ListFileNamesWithOptions() = %v, %v", names, err)
	}
	seen := false
	if err := ReadWithOptions(archive, func(f *archivezip.File) error {
		seen = f.Name == "a.txt"
		return nil
	}, WithOpenZipReader(openZip)); err != nil || !seen {
		t.Fatalf("ReadWithOptions() = %v, seen=%v", err, seen)
	}

	r, err := archivezip.OpenReader(archive)
	if err != nil {
		t.Fatalf("OpenReader() error = %v", err)
	}
	defer func() { _ = r.Close() }()
	var extracted bytes.Buffer
	var mkdirs []string
	if err := UnzipReaderToWithOptions(&r.Reader, "dest",
		WithMkdirAll(func(path string, perm os.FileMode) error {
			mkdirs = append(mkdirs, path)
			return nil
		}),
		WithEvalSymlinks(func(path string) (string, error) {
			return path, nil
		}),
		WithOpenFile(func(path string, flag int, perm os.FileMode) (io.WriteCloser, error) {
			return &zipBufferWriteCloser{Buffer: &extracted}, nil
		}),
	); err != nil {
		t.Fatalf("UnzipReaderToWithOptions() error = %v", err)
	}
	if extracted.String() != "a" || len(mkdirs) == 0 {
		t.Fatalf("extracted=%q mkdirs=%v", extracted.String(), mkdirs)
	}
}
