package poi

import (
	"bytes"
	"errors"
	"path/filepath"
	"reflect"
	"testing"

	knifer "github.com/imajinyun/go-knifer"
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

func TestRowsBufferRoundTrip(t *testing.T) {
	rows := [][]string{{"a", "b"}, {"1", "2"}}
	buf, err := WriteRowsToBuffer("Data", rows)
	if err != nil {
		t.Fatalf("WriteRowsToBuffer: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("buffer is empty")
	}

	got, err := ReadRowsFromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("ReadRowsFromReader: %v", err)
	}
	if !reflect.DeepEqual(got, rows) {
		t.Fatalf("ReadRowsFromReader = %#v, want %#v", got, rows)
	}
}

func TestReadWriteOptions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "book.xlsx")
	rows := [][]string{{"name", "score"}, {"go", "100"}}
	if err := WriteRows(path, rows, WithWriteSheet("Data"), WithStartCell(2, 3)); err != nil {
		t.Fatalf("WriteRows with options: %v", err)
	}
	got, err := ReadRows(path, WithReadSheet("Data"))
	if err != nil {
		t.Fatalf("ReadRows with sheet option: %v", err)
	}
	want := [][]string{nil, {"", "", "name", "score"}, {"", "", "go", "100"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ReadRows with options = %#v, want %#v", got, want)
	}
}

func TestWriteRowsToBufferOptions(t *testing.T) {
	rows := [][]string{{"x"}}
	buf, err := WriteRowsToBuffer("Data", rows, WithStartCell(2, 2))
	if err != nil {
		t.Fatalf("WriteRowsToBuffer with options: %v", err)
	}
	got, err := ReadRowsFromReader(bytes.NewReader(buf.Bytes()), WithReadSheet("Data"))
	if err != nil {
		t.Fatalf("ReadRowsFromReader: %v", err)
	}
	want := [][]string{nil, {"", "x"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("rows = %#v, want %#v", got, want)
	}
}

func TestEmptySheetName(t *testing.T) {
	if err := WriteSheetRows(filepath.Join(t.TempDir(), "book.xlsx"), "", nil); !errors.Is(err, ErrEmptySheetName) {
		t.Fatalf("WriteSheetRows empty sheet error = %v", err)
	}
	if err := WriteSheetRows(filepath.Join(t.TempDir(), "book.xlsx"), "", nil); !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("WriteSheetRows empty sheet error = %v, want ErrCodeInvalidInput", err)
	}
	if _, err := WriteRowsToBuffer("", nil); !errors.Is(err, ErrEmptySheetName) {
		t.Fatalf("WriteRowsToBuffer empty sheet error = %v", err)
	}
}

func TestNoSheetMatchesErrCode(t *testing.T) {
	if !errors.Is(ErrNoSheet, knifer.ErrCodeNotFound) {
		t.Fatalf("ErrNoSheet should match ErrCodeNotFound")
	}
}
