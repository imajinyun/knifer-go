package vzip_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/imajinyun/knifer-go/vzip"
)

func TestFacadeZipDefaultFileHelpers(t *testing.T) {
	tmp, src := newZipArchiveSource(t)
	autoArchive, err := vzip.Zip(filepath.Join(src, "keep.txt"))
	if err != nil {
		t.Fatalf("Zip: %v", err)
	}
	if got, err := vzip.GetBytes(autoArchive, "keep.txt"); err != nil || string(got) != "keep" {
		t.Fatalf("Zip content = %q, %v", got, err)
	}

	toArchive := filepath.Join(tmp, "to.zip")
	if err := vzip.ZipTo(src, toArchive, true); err != nil {
		t.Fatalf("ZipTo: %v", err)
	}
	if got, err := vzip.GetBytes(toArchive, "src/keep.txt"); err != nil || string(got) != "keep" {
		t.Fatalf("ZipTo content = %q, %v", got, err)
	}

	filesArchive := filepath.Join(tmp, "files.zip")
	if err := vzip.ZipFiles(filesArchive, false, filepath.Join(src, "keep.txt")); err != nil {
		t.Fatalf("ZipFiles: %v", err)
	}
	if got, err := vzip.GetBytes(filesArchive, "keep.txt"); err != nil || string(got) != "keep" {
		t.Fatalf("ZipFiles content = %q, %v", got, err)
	}

	filterArchive := filepath.Join(tmp, "filter.zip")
	if err := vzip.ZipFilesFilter(filterArchive, false, zipArchiveTextFilter, src); err != nil {
		t.Fatalf("ZipFilesFilter: %v", err)
	}
	if _, err := vzip.GetBytes(filterArchive, "skip.log"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("ZipFilesFilter skip err = %v, want not exist", err)
	}
}
