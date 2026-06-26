package vzip_test

import (
	archivezip "archive/zip"
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/imajinyun/knifer-go/vzip"
)

func TestFacadeZipFilesUsingOptions(t *testing.T) {
	tmp, src := newZipArchiveSource(t)
	archive := filepath.Join(tmp, "filtered.zip")
	if err := vzip.ZipFilesUsingOptions(archive, []string{src}, vzip.WithSourceDir(true), vzip.WithFileFilter(zipArchiveTextFilter)); err != nil {
		t.Fatalf("ZipFilesUsingOptions: %v", err)
	}
	data, err := vzip.GetBytes(archive, "src/keep.txt")
	if err != nil || string(data) != "keep" {
		t.Fatalf("GetBytes keep = %q, %v", data, err)
	}
	if _, err := vzip.GetBytes(archive, "src/skip.log"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("skip.log err = %v, want not exist", err)
	}
}

func TestFacadeZipEntriesOptionsPreserveCompressionAndOverwrite(t *testing.T) {
	tmp := t.TempDir()
	archive := filepath.Join(tmp, "store.zip")
	if err := vzip.ZipEntriesWithOptions(archive, []vzip.EntryData{{Name: "stored.txt", Data: []byte("stored")}},
		vzip.WithCompressionMethod(archivezip.Store),
		vzip.WithOverwrite(false),
	); err != nil {
		t.Fatalf("ZipEntriesWithOptions: %v", err)
	}

	r, err := archivezip.OpenReader(archive)
	if err != nil {
		t.Fatalf("OpenReader: %v", err)
	}
	defer func() { _ = r.Close() }()
	if len(r.File) != 1 || r.File[0].Method != archivezip.Store {
		t.Fatalf("stored entry method = %#v", r.File)
	}

	if err := vzip.ZipEntriesWithOptions(archive, []vzip.EntryData{{Name: "again.txt", Data: []byte("again")}}, vzip.WithOverwrite(false)); err == nil {
		t.Fatal("ZipEntriesWithOptions overwrite=false error = nil")
	}
}

func TestFacadeUnzipReaderOptionsPreserveModeAndOpenFile(t *testing.T) {
	var buf bytes.Buffer
	zw := archivezip.NewWriter(&buf)
	header := &archivezip.FileHeader{Name: "script.sh", Method: archivezip.Store}
	header.SetMode(0o755)
	w, err := zw.CreateHeader(header)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("#!/bin/sh\n")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	zr, err := archivezip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}

	var gotPath string
	var gotFlag int
	var gotPerm os.FileMode
	var out bytes.Buffer
	if err := vzip.UnzipReaderToWithOptions(zr, t.TempDir(),
		vzip.WithPreserveMode(true),
		vzip.WithOpenFile(func(path string, flag int, perm os.FileMode) (io.WriteCloser, error) {
			gotPath = path
			gotFlag = flag
			gotPerm = perm
			return zipNopWriteCloser{Writer: &out}, nil
		}),
	); err != nil {
		t.Fatalf("UnzipReaderToWithOptions: %v", err)
	}
	if filepath.Base(gotPath) != "script.sh" || gotFlag&os.O_EXCL != 0 || gotPerm.Perm() != 0o755 || out.String() != "#!/bin/sh\n" {
		t.Fatalf("open file args = path %q flag %o perm %o out %q", gotPath, gotFlag, gotPerm.Perm(), out.String())
	}
}

type zipNopWriteCloser struct{ io.Writer }

func (w zipNopWriteCloser) Close() error { return nil }
