package zip

import (
	"os"
	"path/filepath"
	"testing"
)

func TestArchivePermOptions(t *testing.T) {
	tmp := t.TempDir()
	archive := filepath.Join(tmp, "entries.zip")
	entries := []EntryData{{Name: "a.txt", Data: []byte("abcd")}}
	if err := ZipEntriesWithOptions(archive, entries, WithFilePerm(0o600)); err != nil {
		t.Fatalf("ZipEntriesWithOptions: %v", err)
	}
	info, err := os.Stat(archive)
	if err != nil {
		t.Fatalf("stat archive: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("archive perm = %o, want 600", got)
	}

	dest := filepath.Join(tmp, "dest")
	if err := UnzipToWithOptions(archive, dest, WithDirPerm(0o700), WithFilePerm(0o600), WithPreserveMode(false)); err != nil {
		t.Fatalf("UnzipToWithOptions: %v", err)
	}
	fileInfo, err := os.Stat(filepath.Join(dest, "a.txt"))
	if err != nil {
		t.Fatalf("stat extracted: %v", err)
	}
	if got := fileInfo.Mode().Perm(); got != 0o600 {
		t.Fatalf("extracted perm = %o, want 600", got)
	}
}
