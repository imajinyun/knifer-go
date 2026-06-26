package vzip_test

import (
	archivezip "archive/zip"
	"bytes"
	"io"
	"testing"

	"github.com/imajinyun/knifer-go/vzip"
)

func TestFacadeZipToWriterUsingOptions(t *testing.T) {
	_, src := newZipArchiveSource(t)
	var buf bytes.Buffer
	if err := vzip.ZipToWriterUsingOptions(&buf, []string{src}, vzip.WithFileFilter(zipArchiveTextFilter)); err != nil {
		t.Fatalf("ZipToWriterUsingOptions: %v", err)
	}
	bufReader, err := archivezip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	if len(bufReader.File) != 1 || bufReader.File[0].Name != "keep.txt" {
		t.Fatalf("writer archive entries = %#v", bufReader.File)
	}
	entry, err := vzip.GetStream(bufReader.File[0])
	if err != nil {
		t.Fatalf("GetStream: %v", err)
	}
	defer func() { _ = entry.Close() }()
	if _, err := io.ReadAll(entry); err != nil {
		t.Fatalf("read entry: %v", err)
	}
}

func TestFacadeZipToWriter(t *testing.T) {
	_, src := newZipArchiveSource(t)
	var buf bytes.Buffer
	if err := vzip.ZipToWriter(&buf, false, zipArchiveTextFilter, src); err != nil {
		t.Fatalf("ZipToWriter: %v", err)
	}
	zr, err := archivezip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil || len(zr.File) != 1 || zr.File[0].Name != "keep.txt" {
		t.Fatalf("ZipToWriter archive = %#v, %v", zr, err)
	}
}
