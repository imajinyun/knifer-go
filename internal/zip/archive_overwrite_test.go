package zip

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestArchiveOverwriteOption(t *testing.T) {
	tmp := t.TempDir()
	archive := filepath.Join(tmp, "entries.zip")
	entries := []EntryData{{Name: "a.txt", Data: []byte("abcd")}}
	if err := ZipEntriesWithOptions(archive, entries); err != nil {
		t.Fatalf("ZipEntriesWithOptions: %v", err)
	}
	if err := ZipEntriesWithOptions(archive, entries, WithOverwrite(false)); !errors.Is(err, os.ErrExist) {
		t.Fatalf("overwrite false err = %v, want exists", err)
	}

	dest := filepath.Join(tmp, "dest")
	if err := UnzipToWithOptions(archive, dest); err != nil {
		t.Fatalf("UnzipToWithOptions: %v", err)
	}
	if err := UnzipToWithOptions(archive, dest, WithOverwrite(false)); !errors.Is(err, os.ErrExist) {
		t.Fatalf("unzip overwrite false err = %v, want exists", err)
	}
}
