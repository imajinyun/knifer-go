package vpoi_test

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/imajinyun/knifer-go/vpoi"
	"github.com/xuri/excelize/v2"
)

func TestExcelFacadeReadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "book.xlsx")
	rows := [][]string{{"name", "score"}, {"go", "100"}}

	if err := vpoi.WriteRows(path, rows); err != nil {
		t.Fatalf("WriteRows: %v", err)
	}

	sheets, err := vpoi.SheetNames(path)
	if err != nil {
		t.Fatalf("SheetNames: %v", err)
	}
	if !reflect.DeepEqual(sheets, []string{vpoi.DefaultSheetName}) {
		t.Fatalf("SheetNames = %#v", sheets)
	}
	sheets, err = vpoi.SheetNamesWithOptions(path, vpoi.WithOpenOptions(excelize.Options{}))
	if err != nil {
		t.Fatalf("SheetNamesWithOptions: %v", err)
	}
	if !reflect.DeepEqual(sheets, []string{vpoi.DefaultSheetName}) {
		t.Fatalf("SheetNamesWithOptions = %#v", sheets)
	}

	got, err := vpoi.ReadRows(path)
	if err != nil {
		t.Fatalf("ReadRows: %v", err)
	}
	if !reflect.DeepEqual(got, rows) {
		t.Fatalf("ReadRows = %#v, want %#v", got, rows)
	}
	got, err = vpoi.ReadSheetRowsWithOptions(path, vpoi.DefaultSheetName, vpoi.WithOpenOptions(excelize.Options{}))
	if err != nil {
		t.Fatalf("ReadSheetRowsWithOptions: %v", err)
	}
	if !reflect.DeepEqual(got, rows) {
		t.Fatalf("ReadSheetRowsWithOptions = %#v, want %#v", got, rows)
	}
}

func TestExcelFacadeReadRangeOptions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "book.xlsx")
	rows := [][]string{
		{"id", "name", "score"},
		{"1", "go", "100"},
		{"2", "tool", "98"},
	}
	if err := vpoi.WriteRows(path, rows); err != nil {
		t.Fatalf("WriteRows: %v", err)
	}

	got, err := vpoi.ReadRows(path, vpoi.WithReadStartCell(2, 2), vpoi.WithReadLimit(2, 2))
	if err != nil {
		t.Fatalf("ReadRows with range options: %v", err)
	}
	want := [][]string{{"go", "100"}, {"tool", "98"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ReadRows with range options = %#v, want %#v", got, want)
	}
}
