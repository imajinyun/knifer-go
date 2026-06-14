package zip

import (
	"path/filepath"
	"testing"
)

func TestArchiveMaxBytesOption(t *testing.T) {
	tmp := t.TempDir()
	archive := filepath.Join(tmp, "entries.zip")
	entries := []EntryData{{Name: "a.txt", Data: []byte("abcd")}}
	if err := ZipEntriesWithOptions(archive, entries); err != nil {
		t.Fatalf("ZipEntriesWithOptions: %v", err)
	}
	if _, err := GetBytesWithOptions(archive, "a.txt", WithMaxBytes(3)); err == nil {
		t.Fatal("GetBytesWithOptions over limit error = nil")
	}
}
