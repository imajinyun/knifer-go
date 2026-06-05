package vpoi

import (
	"bytes"
	"io"
	"io/fs"

	poiimpl "github.com/imajinyun/go-knifer/internal/poi"
	"github.com/xuri/excelize/v2"
)

// ReadOption customizes Excel read helpers.
type ReadOption = poiimpl.ReadOption

// WriteOption customizes Excel write helpers.
type WriteOption = poiimpl.WriteOption

// WithReadSheet selects the worksheet read by read helpers.
func WithReadSheet(sheet string) ReadOption { return poiimpl.WithReadSheet(sheet) }

// WithOpenOptions sets excelize options used when opening workbooks.
func WithOpenOptions(opts ...excelize.Options) ReadOption { return poiimpl.WithOpenOptions(opts...) }

// WithWriteSheet selects the worksheet written by write helpers.
func WithWriteSheet(sheet string) WriteOption { return poiimpl.WithWriteSheet(sheet) }

// WithStartCell sets the 1-based start row and column used by row-writing helpers.
func WithStartCell(row, col int) WriteOption { return poiimpl.WithStartCell(row, col) }

// WithDirPerm sets the parent-directory permission used when saving workbooks.
func WithDirPerm(perm fs.FileMode) WriteOption { return poiimpl.WithDirPerm(perm) }

// WithFilePerm sets the file permission after saving workbooks.
func WithFilePerm(perm fs.FileMode) WriteOption { return poiimpl.WithFilePerm(perm) }

// WithOverwrite controls whether an existing workbook may be replaced.
func WithOverwrite(overwrite bool) WriteOption { return poiimpl.WithOverwrite(overwrite) }

// WithCreateParents controls whether write helpers create parent directories.
func WithCreateParents(create bool) WriteOption { return poiimpl.WithCreateParents(create) }

// WithSaveOptions sets excelize options used when saving workbooks.
func WithSaveOptions(opts ...excelize.Options) WriteOption { return poiimpl.WithSaveOptions(opts...) }

const (
	// DefaultSheetName is the default worksheet name used for read/write helpers.
	DefaultSheetName = poiimpl.DefaultSheetName
)

var (
	// ErrNoSheet indicates that a workbook does not contain any worksheet.
	ErrNoSheet = poiimpl.ErrNoSheet
	// ErrEmptySheetName indicates an empty worksheet name.
	ErrEmptySheetName = poiimpl.ErrEmptySheetName
)

// SheetNames returns all worksheet names in path.
func SheetNames(path string) ([]string, error) { return poiimpl.SheetNames(path) }

// ReadRows reads rows from the first worksheet in path.
func ReadRows(path string, opts ...ReadOption) ([][]string, error) {
	return poiimpl.ReadRows(path, opts...)
}

// ReadSheetRows reads rows from sheet in path.
func ReadSheetRows(path, sheet string) ([][]string, error) {
	return poiimpl.ReadSheetRows(path, sheet)
}

// ReadRowsFromReader reads rows from the first worksheet in r.
func ReadRowsFromReader(r io.Reader, opts ...ReadOption) ([][]string, error) {
	return poiimpl.ReadRowsFromReader(r, opts...)
}

// WriteRows writes rows into path using the default worksheet name.
func WriteRows(path string, rows [][]string, opts ...WriteOption) error {
	return poiimpl.WriteRows(path, rows, opts...)
}

// WriteSheetRows writes rows into path using sheet.
func WriteSheetRows(path, sheet string, rows [][]string, opts ...WriteOption) error {
	return poiimpl.WriteSheetRows(path, sheet, rows, opts...)
}

// WriteSheets writes multiple worksheets into path.
func WriteSheets(path string, sheets map[string][][]string, opts ...WriteOption) error {
	return poiimpl.WriteSheets(path, sheets, opts...)
}

// WriteRowsToBuffer writes rows into an in-memory XLSX workbook.
func WriteRowsToBuffer(sheet string, rows [][]string, opts ...WriteOption) (*bytes.Buffer, error) {
	return poiimpl.WriteRowsToBuffer(sheet, rows, opts...)
}
