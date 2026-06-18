package vpoi_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/imajinyun/go-knifer/vpoi"
)

func TestFacadePOIOptionSetters(t *testing.T) {
	// Option setters should return non-nil options
	if vpoi.WithOpenFileFunc(nil) == nil {
		t.Fatal("WithOpenFileFunc(nil) returned nil")
	}
	if vpoi.WithOpenReaderFunc(nil) == nil {
		t.Fatal("WithOpenReaderFunc(nil) returned nil")
	}
	if vpoi.WithNewFileFunc(nil) == nil {
		t.Fatal("WithNewFileFunc(nil) returned nil")
	}
	if vpoi.WithSaveAsFunc(nil) == nil {
		t.Fatal("WithSaveAsFunc(nil) returned nil")
	}
}

func TestFacadeReadSheetRows(t *testing.T) {
	rows, err := vpoi.ReadSheetRows("nonexistent.xlsx", "Sheet1")
	if err == nil {
		t.Fatal("ReadSheetRows on nonexistent file should error")
	}
	if rows != nil {
		t.Fatalf("ReadSheetRows rows = %#v, want nil", rows)
	}
}

func TestFacadeWriteSheetRows(t *testing.T) {
	path := filepath.Join(t.TempDir(), "output.xlsx")
	err := vpoi.WriteSheetRows(path, "Sheet1", [][]string{{"a", "b"}})
	if err != nil {
		t.Fatalf("WriteSheetRows error = %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("WriteSheetRows did not create file: %v", err)
	}
}

func TestFacadeWriteSheets(t *testing.T) {
	path := filepath.Join(t.TempDir(), "multi.xlsx")
	err := vpoi.WriteSheets(path, map[string][][]string{
		"Sheet1": {{"h1"}, {"v1"}},
		"Sheet2": {{"h2"}, {"v2"}},
	})
	if err != nil {
		t.Fatalf("WriteSheets error = %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("WriteSheets did not create file: %v", err)
	}
}
