package vpoi_test

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/imajinyun/knifer-go/vpoi"
	"github.com/xuri/excelize/v2"
)

func TestExcelFacadeSheetAndCellOptions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "book.xlsx")
	rows := [][]string{{"name", "score"}, {"go", "100"}}
	if err := vpoi.WriteRows(
		path, rows,
		vpoi.WithWriteSheet("Scores"),
		vpoi.WithStartCell(2, 2),
		vpoi.WithSaveOptions(excelize.Options{}),
	); err != nil {
		t.Fatalf("WriteRows with sheet/cell options: %v", err)
	}
	sheets, err := vpoi.SheetNames(path)
	if err != nil {
		t.Fatalf("SheetNames: %v", err)
	}
	if !reflect.DeepEqual(sheets, []string{"Scores"}) {
		t.Fatalf("SheetNames = %#v, want [Scores]", sheets)
	}
	f, err := excelize.OpenFile(path)
	if err != nil {
		t.Fatalf("open workbook: %v", err)
	}
	defer func() { _ = f.Close() }()
	if got, err := f.GetCellValue("Scores", "B2"); err != nil || got != "name" {
		t.Fatalf("Scores!B2 = %q, %v; want name, nil", got, err)
	}
	got, err := vpoi.ReadRows(path, vpoi.WithReadSheet("Scores"), vpoi.WithOpenOptions(excelize.Options{}))
	if err != nil {
		t.Fatalf("ReadRows with sheet/open options: %v", err)
	}
	if len(got) == 0 {
		t.Fatalf("ReadRows with sheet/open options returned no rows")
	}
}
