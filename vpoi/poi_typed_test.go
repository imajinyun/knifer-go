package vpoi_test

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/imajinyun/knifer-go/vpoi"
)

func TestExcelFacadeTypedRowsAndCells(t *testing.T) {
	path := filepath.Join(t.TempDir(), "typed.xlsx")
	rows := [][]any{
		{"name", "score", "active"},
		{"go", 100, true},
	}

	if err := vpoi.WriteAnyRows(path, rows); err != nil {
		t.Fatalf("WriteAnyRows: %v", err)
	}

	cells, err := vpoi.ReadCells(path, vpoi.WithReadStartCell(2, 2), vpoi.WithReadLimit(1, 2))
	if err != nil {
		t.Fatalf("ReadCells: %v", err)
	}
	if len(cells) != 1 || len(cells[0]) != 2 {
		t.Fatalf("ReadCells shape = %#v, want one row with two cells", cells)
	}
	if cells[0][0].Value != "100" || cells[0][0].Type != vpoi.CellTypeNumber {
		t.Fatalf("score cell = %#v, want value 100 numeric type", cells[0][0])
	}
	if cells[0][1].Value != "TRUE" || cells[0][1].Type != vpoi.CellTypeBool {
		t.Fatalf("active cell = %#v, want TRUE bool type", cells[0][1])
	}
}

func TestExcelFacadeTypedRowsBuffer(t *testing.T) {
	buf, err := vpoi.WriteAnyRowsToBuffer("Typed", [][]any{{"id", "ok"}, {1, false}})
	if err != nil {
		t.Fatalf("WriteAnyRowsToBuffer: %v", err)
	}

	cells, err := vpoi.ReadCellsFromReader(bytes.NewReader(buf.Bytes()), vpoi.WithReadSheet("Typed"))
	if err != nil {
		t.Fatalf("ReadCellsFromReader: %v", err)
	}
	if cells[1][0].Type != vpoi.CellTypeNumber || cells[1][1].Type != vpoi.CellTypeBool {
		t.Fatalf("typed cells = %#v, want number and bool types", cells[1])
	}
}
