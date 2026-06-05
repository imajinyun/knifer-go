package poi

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	knifer "github.com/imajinyun/go-knifer"
	"github.com/xuri/excelize/v2"
)

const (
	// DefaultSheetName is the default worksheet name used for read/write helpers.
	DefaultSheetName = "Sheet1"
)

type sentinel struct {
	code knifer.ErrCode
	msg  string
}

func (e *sentinel) Error() string { return e.msg }

func (e *sentinel) ErrorCode() knifer.ErrCode { return e.code }

func (e *sentinel) Is(target error) bool {
	if e == target {
		return true
	}
	code, ok := target.(knifer.ErrCode)
	return ok && e.code == code
}

var (
	// ErrNoSheet indicates that a workbook does not contain any worksheet.
	ErrNoSheet error = &sentinel{code: knifer.ErrCodeNotFound, msg: "poi: workbook has no sheet"}
	// ErrEmptySheetName indicates an empty worksheet name.
	ErrEmptySheetName error = &sentinel{code: knifer.ErrCodeInvalidInput, msg: "poi: sheet name is empty"}
)

type readConfig struct {
	sheet       string
	openOptions []excelize.Options
}

// ReadOption customizes Excel read helpers.
type ReadOption func(*readConfig)

type writeConfig struct {
	sheet         string
	startRow      int
	startCol      int
	dirPerm       fs.FileMode
	filePerm      fs.FileMode
	overwrite     bool
	createParents bool
	saveOptions   []excelize.Options
}

// WriteOption customizes Excel write helpers.
type WriteOption func(*writeConfig)

func defaultReadConfig() readConfig { return readConfig{} }

func defaultWriteConfig() writeConfig {
	return writeConfig{sheet: DefaultSheetName, startRow: 1, startCol: 1, dirPerm: 0o750, filePerm: 0o644, overwrite: true, createParents: true}
}

// WithReadSheet selects the worksheet read by read helpers.
func WithReadSheet(sheet string) ReadOption { return func(c *readConfig) { c.sheet = sheet } }

// WithOpenOptions sets excelize options used when opening workbooks.
func WithOpenOptions(opts ...excelize.Options) ReadOption {
	return func(c *readConfig) { c.openOptions = append([]excelize.Options(nil), opts...) }
}

// WithWriteSheet selects the worksheet written by write helpers.
func WithWriteSheet(sheet string) WriteOption { return func(c *writeConfig) { c.sheet = sheet } }

// WithStartCell sets the 1-based start row and column used by row-writing helpers.
func WithStartCell(row, col int) WriteOption {
	return func(c *writeConfig) {
		if row > 0 {
			c.startRow = row
		}
		if col > 0 {
			c.startCol = col
		}
	}
}

// WithDirPerm sets the parent-directory permission used when saving workbooks.
func WithDirPerm(perm fs.FileMode) WriteOption { return func(c *writeConfig) { c.dirPerm = perm } }

// WithFilePerm sets the file permission after saving workbooks.
func WithFilePerm(perm fs.FileMode) WriteOption { return func(c *writeConfig) { c.filePerm = perm } }

// WithOverwrite controls whether an existing workbook may be replaced.
func WithOverwrite(overwrite bool) WriteOption {
	return func(c *writeConfig) { c.overwrite = overwrite }
}

// WithCreateParents controls whether write helpers create parent directories.
func WithCreateParents(create bool) WriteOption {
	return func(c *writeConfig) { c.createParents = create }
}

// WithSaveOptions sets excelize options used when saving workbooks.
func WithSaveOptions(opts ...excelize.Options) WriteOption {
	return func(c *writeConfig) { c.saveOptions = append([]excelize.Options(nil), opts...) }
}

func applyReadOptions(opts []ReadOption) readConfig {
	cfg := defaultReadConfig()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return cfg
}

func applyWriteOptions(opts []WriteOption) writeConfig {
	cfg := defaultWriteConfig()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return cfg
}

// SheetNames returns all worksheet names in path.
func SheetNames(path string) ([]string, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	return f.GetSheetList(), nil
}

// ReadRows reads rows from the first worksheet in path.
func ReadRows(path string, opts ...ReadOption) ([][]string, error) {
	cfg := applyReadOptions(opts)
	f, err := excelize.OpenFile(path, cfg.openOptions...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	return readRowsWithConfig(f, cfg)
}

// ReadSheetRows reads rows from sheet in path.
func ReadSheetRows(path, sheet string) ([][]string, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	return readSheetRows(f, sheet)
}

// ReadRowsFromReader reads rows from the first worksheet in r.
func ReadRowsFromReader(r io.Reader, opts ...ReadOption) ([][]string, error) {
	cfg := applyReadOptions(opts)
	f, err := excelize.OpenReader(r, cfg.openOptions...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	return readRowsWithConfig(f, cfg)
}

// WriteRows writes rows into path using the default worksheet name.
func WriteRows(path string, rows [][]string, opts ...WriteOption) error {
	return writeRows(path, rows, applyWriteOptions(opts))
}

// WriteSheetRows writes rows into path using sheet.
func WriteSheetRows(path, sheet string, rows [][]string, opts ...WriteOption) error {
	allOpts := append([]WriteOption{WithWriteSheet(sheet)}, opts...)
	return writeRows(path, rows, applyWriteOptions(allOpts))
}

func writeRows(path string, rows [][]string, cfg writeConfig) error {
	sheet := cfg.sheet
	if sheet == "" {
		return ErrEmptySheetName
	}
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()
	if err := replaceDefaultSheet(f, sheet); err != nil {
		return err
	}
	if err := setRows(f, sheet, rows, cfg.startRow, cfg.startCol); err != nil {
		return err
	}
	if cfg.createParents {
		if err := ensureParentDir(path, cfg.dirPerm); err != nil {
			return err
		}
	}
	return saveWorkbook(f, path, cfg)
}

// WriteSheets writes multiple worksheets into path.
func WriteSheets(path string, sheets map[string][][]string, opts ...WriteOption) error {
	cfg := applyWriteOptions(opts)
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	if len(sheets) == 0 {
		if cfg.createParents {
			if err := ensureParentDir(path, cfg.dirPerm); err != nil {
				return err
			}
		}
		return saveWorkbook(f, path, cfg)
	}

	first := true
	for sheet, rows := range sheets {
		if sheet == "" {
			return ErrEmptySheetName
		}
		if first {
			if err := replaceDefaultSheet(f, sheet); err != nil {
				return err
			}
			first = false
		} else if _, err := f.NewSheet(sheet); err != nil {
			return err
		}
		if err := setRows(f, sheet, rows, cfg.startRow, cfg.startCol); err != nil {
			return err
		}
	}
	if cfg.createParents {
		if err := ensureParentDir(path, cfg.dirPerm); err != nil {
			return err
		}
	}
	return saveWorkbook(f, path, cfg)
}

// WriteRowsToBuffer writes rows into an in-memory XLSX workbook.
func WriteRowsToBuffer(sheet string, rows [][]string, opts ...WriteOption) (*bytes.Buffer, error) {
	allOpts := append([]WriteOption{WithWriteSheet(sheet)}, opts...)
	cfg := applyWriteOptions(allOpts)
	sheet = cfg.sheet
	if sheet == "" {
		return nil, ErrEmptySheetName
	}
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()
	if err := replaceDefaultSheet(f, sheet); err != nil {
		return nil, err
	}
	if err := setRows(f, sheet, rows, cfg.startRow, cfg.startCol); err != nil {
		return nil, err
	}
	return f.WriteToBuffer()
}

func readFirstSheetRows(f *excelize.File) ([][]string, error) {
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, ErrNoSheet
	}
	return readSheetRows(f, sheets[0])
}

func readRowsWithConfig(f *excelize.File, cfg readConfig) ([][]string, error) {
	if cfg.sheet != "" {
		return readSheetRows(f, cfg.sheet)
	}
	return readFirstSheetRows(f)
}

func readSheetRows(f *excelize.File, sheet string) ([][]string, error) {
	if sheet == "" {
		return nil, ErrEmptySheetName
	}
	return f.GetRows(sheet)
}

func replaceDefaultSheet(f *excelize.File, sheet string) error {
	if sheet == DefaultSheetName {
		return nil
	}
	if err := f.SetSheetName(DefaultSheetName, sheet); err != nil {
		return err
	}
	return nil
}

func setRows(f *excelize.File, sheet string, rows [][]string, startRow, startCol int) error {
	if startRow <= 0 {
		startRow = 1
	}
	if startCol <= 0 {
		startCol = 1
	}
	for rowIndex, row := range rows {
		for colIndex, value := range row {
			cell, err := excelize.CoordinatesToCellName(startCol+colIndex, startRow+rowIndex)
			if err != nil {
				return fmt.Errorf("poi: cell coordinates row=%d col=%d: %w", rowIndex+1, colIndex+1, err)
			}
			if err := f.SetCellStr(sheet, cell, value); err != nil {
				return err
			}
		}
	}
	return nil
}

func ensureParentDir(path string, perm fs.FileMode) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, perm)
}

func saveWorkbook(f *excelize.File, path string, cfg writeConfig) error {
	if !cfg.overwrite {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("poi: file already exists: %s", path)
		} else if !os.IsNotExist(err) {
			return err
		}
	}
	if err := f.SaveAs(path, cfg.saveOptions...); err != nil {
		return err
	}
	if cfg.filePerm != 0 {
		return os.Chmod(path, cfg.filePerm)
	}
	return nil
}
