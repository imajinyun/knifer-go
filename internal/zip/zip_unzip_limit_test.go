package zip

import (
	archivezip "archive/zip"
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestUnzipDefaultLimitCanBeOverridden(t *testing.T) {
	tmp := t.TempDir()
	archive := filepath.Join(tmp, "limit.zip")
	entries := []EntryData{{Name: "a.txt", Data: []byte("abcd")}}
	if err := ZipEntries(archive, entries...); err != nil {
		t.Fatalf("ZipEntries: %v", err)
	}
	if err := UnzipToWithOptions(archive, filepath.Join(tmp, "limited"), WithMaxBytes(3)); err == nil {
		t.Fatal("UnzipToWithOptions should reject archives larger than max bytes")
	}
	if err := UnzipToLimit(archive, filepath.Join(tmp, "unlimited"), -1); err != nil {
		t.Fatalf("UnzipToLimit with explicit unlimited limit: %v", err)
	}
	if got := applyUnzipOptions(nil).maxBytes; got != DefaultUnzipMaxBytes {
		t.Fatalf("default unzip max bytes = %d, want %d", got, DefaultUnzipMaxBytes)
	}
	if got := applyUnzipOptions([]ArchiveOption{WithMaxBytes(-1)}).maxBytes; got != -1 {
		t.Fatalf("explicit unlimited unzip max bytes = %d, want -1", got)
	}
}

func TestUnzipEnforcesActualCopiedBytes(t *testing.T) {
	var buf bytes.Buffer
	if err := ZipEntriesToWriter(&buf, EntryData{Name: "a.txt", Data: []byte("abcd")}); err != nil {
		t.Fatalf("ZipEntriesToWriter: %v", err)
	}
	r, err := archivezip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	if len(r.File) != 1 {
		t.Fatalf("archive entries = %d, want 1", len(r.File))
	}
	// Simulate an archive whose declared uncompressed size is smaller than the
	// actual stream. Extraction must enforce the copy-time limit as a second line
	// of defense instead of trusting central-directory metadata only.
	r.File[0].UncompressedSize64 = 1
	dest := filepath.Join(t.TempDir(), "dest")
	if err := UnzipReaderToWithOptions(r, dest, WithMaxBytes(3)); err == nil {
		t.Fatal("UnzipReaderToWithOptions should reject streams exceeding the actual copy limit")
	}
	if _, err := os.Stat(filepath.Join(dest, "a.txt")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("failed extraction should remove partial file, stat err=%v", err)
	}
}
