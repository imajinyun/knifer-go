package vzip_test

import (
	archivezip "archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/imajinyun/knifer-go/vzip"
)

func TestFacadeZipStreamAndReaderHelpers(t *testing.T) {
	tmp := t.TempDir()
	archive := filepath.Join(tmp, "streams.zip")

	if err := vzip.ZipStreams(archive, vzip.StreamEntry{Name: "stream.txt", Reader: strings.NewReader("stream")}); err != nil {
		t.Fatalf("ZipStreams: %v", err)
	}
	if got, err := vzip.GetBytes(archive, "stream.txt"); err != nil || string(got) != "stream" {
		t.Fatalf("ZipStreams content = %q, %v", got, err)
	}

	var buf bytes.Buffer
	if err := vzip.ZipEntriesToWriter(&buf, vzip.EntryData{Name: "entry.txt", Data: []byte("entry")}); err != nil {
		t.Fatalf("ZipEntriesToWriter: %v", err)
	}
	var streamBuf bytes.Buffer
	if err := vzip.ZipStreamsToWriter(&streamBuf, vzip.StreamEntry{Name: "stream2.txt", Reader: strings.NewReader("stream2")}); err != nil {
		t.Fatalf("ZipStreamsToWriter: %v", err)
	}
	streamReader, err := archivezip.NewReader(bytes.NewReader(streamBuf.Bytes()), int64(streamBuf.Len()))
	if err != nil || len(streamReader.File) != 1 || streamReader.File[0].Name != "stream2.txt" {
		t.Fatalf("ZipStreamsToWriter archive = %#v, %v", streamReader, err)
	}

	dest := filepath.Join(tmp, "dest")
	zr, err := archivezip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	if err := vzip.UnzipReaderTo(zr, dest); err != nil {
		t.Fatalf("UnzipReaderTo: %v", err)
	}
	if got, err := os.ReadFile(filepath.Join(dest, "entry.txt")); err != nil || string(got) != "entry" {
		t.Fatalf("UnzipReaderTo output = %q, %v", got, err)
	}
	if err := vzip.UnzipReaderToLimit(zr, filepath.Join(tmp, "limited"), 1); err == nil {
		t.Fatal("UnzipReaderToLimit should reject archive larger than limit")
	}
}
