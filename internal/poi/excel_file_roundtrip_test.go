package poi

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestWriteAndReadSheetRows(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "book.xlsx")
	rows := [][]string{
		{"name", "score"},
		{"go", "100"},
		{"tool", "98"},
	}

	if err := WriteSheetRows(path, "Scores", rows); err != nil {
		t.Fatalf("WriteSheetRows: %v", err)
	}

	sheets, err := SheetNames(path)
	if err != nil {
		t.Fatalf("SheetNames: %v", err)
	}
	if !reflect.DeepEqual(sheets, []string{"Scores"}) {
		t.Fatalf("SheetNames = %#v", sheets)
	}
	sheets, err = SheetNamesWithOptions(path, WithOpenOptions())
	if err != nil {
		t.Fatalf("SheetNamesWithOptions: %v", err)
	}
	if !reflect.DeepEqual(sheets, []string{"Scores"}) {
		t.Fatalf("SheetNamesWithOptions = %#v", sheets)
	}

	got, err := ReadRows(path)
	if err != nil {
		t.Fatalf("ReadRows: %v", err)
	}
	if !reflect.DeepEqual(got, rows) {
		t.Fatalf("ReadRows = %#v, want %#v", got, rows)
	}

	got, err = ReadSheetRows(path, "Scores")
	if err != nil {
		t.Fatalf("ReadSheetRows: %v", err)
	}
	if !reflect.DeepEqual(got, rows) {
		t.Fatalf("ReadSheetRows = %#v, want %#v", got, rows)
	}
	got, err = ReadSheetRowsWithOptions(path, "Scores", WithOpenOptions())
	if err != nil {
		t.Fatalf("ReadSheetRowsWithOptions: %v", err)
	}
	if !reflect.DeepEqual(got, rows) {
		t.Fatalf("ReadSheetRowsWithOptions = %#v, want %#v", got, rows)
	}
}

func TestWriteSheets(t *testing.T) {
	path := filepath.Join(t.TempDir(), "book.xlsx")
	sheets := map[string][][]string{
		"Users":  {{"id", "name"}, {"1", "alice"}},
		"Orders": {{"id", "amount"}, {"A1", "9.9"}},
	}

	if err := WriteSheets(path, sheets); err != nil {
		t.Fatalf("WriteSheets: %v", err)
	}

	for name, want := range sheets {
		got, err := ReadSheetRows(path, name)
		if err != nil {
			t.Fatalf("ReadSheetRows(%s): %v", name, err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("ReadSheetRows(%s) = %#v, want %#v", name, got, want)
		}
	}
}
