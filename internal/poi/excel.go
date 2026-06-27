package poi

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unicode/utf8"

	knifer "github.com/imajinyun/knifer-go"
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
	// ErrInvalidSheetName indicates a worksheet name that Excel cannot represent.
	ErrInvalidSheetName error = &sentinel{code: knifer.ErrCodeInvalidInput, msg: "poi: sheet name is invalid"}
)

const maxSheetNameRunes = 31

var invalidSheetNameChars = strings.NewReplacer(
	":", "",
	"\\", "",
	"/", "",
	"?", "",
	"*", "",
	"[", "",
	"]", "",
)

// CellType identifies the workbook cell value type reported by Excelize.
type CellType = excelize.CellType

// Cell contains a worksheet cell value with its type and 1-based position.
type Cell struct {
	Value string
	Type  CellType
	Axis  string
	Row   int
	Col   int
}

type readConfig struct {
	sheet       string
	startRow    int
	startCol    int
	maxRows     int
	maxCols     int
	openOptions []excelize.Options
	openFile    OpenFileFunc
	openReader  OpenReaderFunc
}

// ReadOption customizes Excel read helpers.
type ReadOption func(*readConfig)

// OpenFileFunc opens an Excel workbook from a file path.
type OpenFileFunc func(string, ...excelize.Options) (*excelize.File, error)

// OpenReaderFunc opens an Excel workbook from a reader.
type OpenReaderFunc func(io.Reader, ...excelize.Options) (*excelize.File, error)

// NewFileFunc creates a new Excel workbook.
type NewFileFunc func() *excelize.File

// SaveAsFunc saves an Excel workbook to path.
type SaveAsFunc func(*excelize.File, string, ...excelize.Options) error

type writeConfig struct {
	sheet         string
	startRow      int
	startCol      int
	dirPerm       fs.FileMode
	filePerm      fs.FileMode
	overwrite     bool
	createParents bool
	saveOptions   []excelize.Options
	mkdirAll      func(string, fs.FileMode) error
	stat          func(string) (os.FileInfo, error)
	chmod         func(string, fs.FileMode) error
	newFile       NewFileFunc
	saveAs        SaveAsFunc
}

// WriteOption customizes Excel write helpers.
type WriteOption func(*writeConfig)

func defaultOpenFile(path string, opts ...excelize.Options) (*excelize.File, error) {
	return excelize.OpenFile(path, opts...)
}

func defaultOpenReader(r io.Reader, opts ...excelize.Options) (*excelize.File, error) {
	return excelize.OpenReader(r, opts...)
}

func defaultNewFile() *excelize.File { return excelize.NewFile() }

func defaultSaveAs(f *excelize.File, path string, opts ...excelize.Options) error {
	return f.SaveAs(path, opts...)
}

func invalidWorkbookError() error {
	return knifer.NewError(knifer.ErrCodeInvalidInput, "poi: workbook is nil")
}

func defaultReadConfig() readConfig {
	return readConfig{startRow: 1, startCol: 1, openFile: defaultOpenFile, openReader: defaultOpenReader}
}

func defaultWriteConfig() writeConfig {
	return writeConfig{sheet: DefaultSheetName, startRow: 1, startCol: 1, dirPerm: 0o750, filePerm: 0o644, overwrite: true, createParents: true, mkdirAll: os.MkdirAll, stat: os.Stat, chmod: os.Chmod, newFile: defaultNewFile, saveAs: defaultSaveAs}
}

// WithReadSheet selects the worksheet read by read helpers.
func WithReadSheet(sheet string) ReadOption { return func(c *readConfig) { c.sheet = sheet } }

// WithReadStartCell sets the 1-based start row and column used by row-reading helpers.
func WithReadStartCell(row, col int) ReadOption {
	return func(c *readConfig) {
		if row > 0 {
			c.startRow = row
		}
		if col > 0 {
			c.startCol = col
		}
	}
}

// WithReadLimit limits the number of rows and columns returned by row-reading helpers.
func WithReadLimit(maxRows, maxCols int) ReadOption {
	return func(c *readConfig) {
		if maxRows > 0 {
			c.maxRows = maxRows
		}
		if maxCols > 0 {
			c.maxCols = maxCols
		}
	}
}

// WithOpenOptions sets excelize options used when opening workbooks.
func WithOpenOptions(opts ...excelize.Options) ReadOption {
	return func(c *readConfig) { c.openOptions = slices.Clone(opts) }
}

// WithOpenFileFunc sets the workbook opener used by path-based read helpers.
func WithOpenFileFunc(openFile OpenFileFunc) ReadOption {
	return func(c *readConfig) {
		if openFile != nil {
			c.openFile = openFile
		}
	}
}

// WithOpenReaderFunc sets the workbook opener used by reader-based read helpers.
func WithOpenReaderFunc(openReader OpenReaderFunc) ReadOption {
	return func(c *readConfig) {
		if openReader != nil {
			c.openReader = openReader
		}
	}
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
	return func(c *writeConfig) { c.saveOptions = slices.Clone(opts) }
}

// WithMkdirAll sets the directory creator used when saving workbooks.
func WithMkdirAll(mkdirAll func(string, fs.FileMode) error) WriteOption {
	return func(c *writeConfig) { c.mkdirAll = mkdirAll }
}

// WithStat sets the stat provider used when checking workbook overwrite behavior.
func WithStat(stat func(string) (os.FileInfo, error)) WriteOption {
	return func(c *writeConfig) { c.stat = stat }
}

// WithChmod sets the chmod provider used after saving workbooks.
func WithChmod(chmod func(string, fs.FileMode) error) WriteOption {
	return func(c *writeConfig) { c.chmod = chmod }
}

// WithNewFileFunc sets the workbook factory used by write helpers.
func WithNewFileFunc(newFile NewFileFunc) WriteOption {
	return func(c *writeConfig) {
		if newFile != nil {
			c.newFile = newFile
		}
	}
}

// WithSaveAsFunc sets the workbook saver used by write helpers.
func WithSaveAsFunc(saveAs SaveAsFunc) WriteOption {
	return func(c *writeConfig) {
		if saveAs != nil {
			c.saveAs = saveAs
		}
	}
}

// IsValidSheetName reports whether sheet can be used as an Excel worksheet name.
func IsValidSheetName(sheet string) bool { return ValidateSheetName(sheet) == nil }

// ValidateSheetName validates Excel worksheet naming constraints.
func ValidateSheetName(sheet string) error {
	if sheet == "" {
		return ErrEmptySheetName
	}
	if utf8.RuneCountInString(sheet) > maxSheetNameRunes || invalidSheetNameChars.Replace(sheet) != sheet {
		return knifer.WrapError(knifer.ErrCodeInvalidInput, "poi: invalid sheet name", ErrInvalidSheetName)
	}
	return nil
}

func applyReadOptions(opts []ReadOption) readConfig {
	cfg := defaultReadConfig()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.openFile == nil {
		cfg.openFile = defaultOpenFile
	}
	if cfg.openReader == nil {
		cfg.openReader = defaultOpenReader
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
	if cfg.mkdirAll == nil {
		cfg.mkdirAll = os.MkdirAll
	}
	if cfg.stat == nil {
		cfg.stat = os.Stat
	}
	if cfg.chmod == nil {
		cfg.chmod = os.Chmod
	}
	if cfg.newFile == nil {
		cfg.newFile = defaultNewFile
	}
	if cfg.saveAs == nil {
		cfg.saveAs = defaultSaveAs
	}
	return cfg
}

// SheetNames returns all worksheet names in path.
func SheetNames(path string) ([]string, error) {
	return SheetNamesWithOptions(path)
}

// SheetNamesWithOptions returns all worksheet names in path with custom open options.
func SheetNamesWithOptions(path string, opts ...ReadOption) ([]string, error) {
	cfg := applyReadOptions(opts)
	if cfg.sheet != "" {
		if err := ValidateSheetName(cfg.sheet); err != nil {
			return nil, err
		}
	}
	f, err := cfg.openFile(path, cfg.openOptions...)
	if err != nil {
		return nil, err
	}
	if f == nil {
		return nil, invalidWorkbookError()
	}
	defer func() { _ = f.Close() }()
	return f.GetSheetList(), nil
}

// ReadRows reads rows from the first worksheet in path.
func ReadRows(path string, opts ...ReadOption) ([][]string, error) {
	cfg := applyReadOptions(opts)
	if cfg.sheet != "" {
		if err := ValidateSheetName(cfg.sheet); err != nil {
			return nil, err
		}
	}
	f, err := cfg.openFile(path, cfg.openOptions...)
	if err != nil {
		return nil, err
	}
	if f == nil {
		return nil, invalidWorkbookError()
	}
	defer func() { _ = f.Close() }()
	return readRowsWithConfig(f, cfg)
}

// ReadSheetRows reads rows from sheet in path.
func ReadSheetRows(path, sheet string) ([][]string, error) {
	return ReadSheetRowsWithOptions(path, sheet)
}

// ReadSheetRowsWithOptions reads rows from sheet in path with custom open options.
func ReadSheetRowsWithOptions(path, sheet string, opts ...ReadOption) ([][]string, error) {
	if err := ValidateSheetName(sheet); err != nil {
		return nil, err
	}
	cfg := applyReadOptions(opts)
	f, err := cfg.openFile(path, cfg.openOptions...)
	if err != nil {
		return nil, err
	}
	if f == nil {
		return nil, invalidWorkbookError()
	}
	defer func() { _ = f.Close() }()
	return readSheetRowsWithConfig(f, sheet, cfg)
}

// ReadRowsFromReader reads rows from the first worksheet in r.
func ReadRowsFromReader(r io.Reader, opts ...ReadOption) ([][]string, error) {
	cfg := applyReadOptions(opts)
	if cfg.sheet != "" {
		if err := ValidateSheetName(cfg.sheet); err != nil {
			return nil, err
		}
	}
	f, err := cfg.openReader(r, cfg.openOptions...)
	if err != nil {
		return nil, err
	}
	if f == nil {
		return nil, invalidWorkbookError()
	}
	defer func() { _ = f.Close() }()
	return readRowsWithConfig(f, cfg)
}

// ReadCells reads typed cell metadata from the first worksheet in path.
func ReadCells(path string, opts ...ReadOption) ([][]Cell, error) {
	cfg := applyReadOptions(opts)
	if cfg.sheet != "" {
		if err := ValidateSheetName(cfg.sheet); err != nil {
			return nil, err
		}
	}
	f, err := cfg.openFile(path, cfg.openOptions...)
	if err != nil {
		return nil, err
	}
	if f == nil {
		return nil, invalidWorkbookError()
	}
	defer func() { _ = f.Close() }()
	return readCellsWithConfig(f, cfg)
}

// ReadSheetCellsWithOptions reads typed cell metadata from sheet in path.
func ReadSheetCellsWithOptions(path, sheet string, opts ...ReadOption) ([][]Cell, error) {
	if err := ValidateSheetName(sheet); err != nil {
		return nil, err
	}
	cfg := applyReadOptions(opts)
	f, err := cfg.openFile(path, cfg.openOptions...)
	if err != nil {
		return nil, err
	}
	if f == nil {
		return nil, invalidWorkbookError()
	}
	defer func() { _ = f.Close() }()
	return readSheetCellsWithConfig(f, sheet, cfg)
}

// ReadCellsFromReader reads typed cell metadata from the first worksheet in r.
func ReadCellsFromReader(r io.Reader, opts ...ReadOption) ([][]Cell, error) {
	cfg := applyReadOptions(opts)
	if cfg.sheet != "" {
		if err := ValidateSheetName(cfg.sheet); err != nil {
			return nil, err
		}
	}
	f, err := cfg.openReader(r, cfg.openOptions...)
	if err != nil {
		return nil, err
	}
	if f == nil {
		return nil, invalidWorkbookError()
	}
	defer func() { _ = f.Close() }()
	return readCellsWithConfig(f, cfg)
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

// WriteAnyRows writes typed cell values into path using the default worksheet name.
func WriteAnyRows(path string, rows [][]any, opts ...WriteOption) error {
	return writeAnyRows(path, rows, applyWriteOptions(opts))
}

// WriteSheetAnyRows writes typed cell values into path using sheet.
func WriteSheetAnyRows(path, sheet string, rows [][]any, opts ...WriteOption) error {
	allOpts := append([]WriteOption{WithWriteSheet(sheet)}, opts...)
	return writeAnyRows(path, rows, applyWriteOptions(allOpts))
}

func writeRows(path string, rows [][]string, cfg writeConfig) error {
	sheet := cfg.sheet
	if err := ValidateSheetName(sheet); err != nil {
		return err
	}
	f := cfg.newFile()
	if f == nil {
		return invalidWorkbookError()
	}
	defer func() { _ = f.Close() }()
	if err := replaceDefaultSheet(f, sheet); err != nil {
		return err
	}
	if err := setRows(f, sheet, rows, cfg.startRow, cfg.startCol); err != nil {
		return err
	}
	if cfg.createParents {
		if err := ensureParentDir(path, cfg); err != nil {
			return err
		}
	}
	return saveWorkbook(f, path, cfg)
}

func writeAnyRows(path string, rows [][]any, cfg writeConfig) error {
	sheet := cfg.sheet
	if err := ValidateSheetName(sheet); err != nil {
		return err
	}
	f := cfg.newFile()
	if f == nil {
		return invalidWorkbookError()
	}
	defer func() { _ = f.Close() }()
	if err := replaceDefaultSheet(f, sheet); err != nil {
		return err
	}
	if err := setAnyRows(f, sheet, rows, cfg.startRow, cfg.startCol); err != nil {
		return err
	}
	if cfg.createParents {
		if err := ensureParentDir(path, cfg); err != nil {
			return err
		}
	}
	return saveWorkbook(f, path, cfg)
}

// WriteSheets writes multiple worksheets into path.
func WriteSheets(path string, sheets map[string][][]string, opts ...WriteOption) error {
	cfg := applyWriteOptions(opts)
	f := cfg.newFile()
	if f == nil {
		return invalidWorkbookError()
	}
	defer func() { _ = f.Close() }()

	if len(sheets) == 0 {
		if cfg.createParents {
			if err := ensureParentDir(path, cfg); err != nil {
				return err
			}
		}
		return saveWorkbook(f, path, cfg)
	}

	sheetNames := make([]string, 0, len(sheets))
	for sheet := range sheets {
		if err := ValidateSheetName(sheet); err != nil {
			return err
		}
		sheetNames = append(sheetNames, sheet)
	}
	slices.Sort(sheetNames)

	first := true
	for _, sheet := range sheetNames {
		rows := sheets[sheet]
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
		if err := ensureParentDir(path, cfg); err != nil {
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
	if err := ValidateSheetName(sheet); err != nil {
		return nil, err
	}
	f := cfg.newFile()
	if f == nil {
		return nil, invalidWorkbookError()
	}
	defer func() { _ = f.Close() }()
	if err := replaceDefaultSheet(f, sheet); err != nil {
		return nil, err
	}
	if err := setRows(f, sheet, rows, cfg.startRow, cfg.startCol); err != nil {
		return nil, err
	}
	return f.WriteToBuffer()
}

// WriteAnyRowsToBuffer writes typed cell values into an in-memory XLSX workbook.
func WriteAnyRowsToBuffer(sheet string, rows [][]any, opts ...WriteOption) (*bytes.Buffer, error) {
	allOpts := append([]WriteOption{WithWriteSheet(sheet)}, opts...)
	cfg := applyWriteOptions(allOpts)
	sheet = cfg.sheet
	if err := ValidateSheetName(sheet); err != nil {
		return nil, err
	}
	f := cfg.newFile()
	if f == nil {
		return nil, invalidWorkbookError()
	}
	defer func() { _ = f.Close() }()
	if err := replaceDefaultSheet(f, sheet); err != nil {
		return nil, err
	}
	if err := setAnyRows(f, sheet, rows, cfg.startRow, cfg.startCol); err != nil {
		return nil, err
	}
	return f.WriteToBuffer()
}

func readRowsWithConfig(f *excelize.File, cfg readConfig) ([][]string, error) {
	if f == nil {
		return nil, invalidWorkbookError()
	}
	if cfg.sheet != "" {
		return readSheetRowsWithConfig(f, cfg.sheet, cfg)
	}
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, ErrNoSheet
	}
	return readSheetRowsWithConfig(f, sheets[0], cfg)
}

func readSheetRowsWithConfig(f *excelize.File, sheet string, cfg readConfig) ([][]string, error) {
	if f == nil {
		return nil, invalidWorkbookError()
	}
	if err := ValidateSheetName(sheet); err != nil {
		return nil, err
	}
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, err
	}
	return limitRows(rows, cfg), nil
}

func readCellsWithConfig(f *excelize.File, cfg readConfig) ([][]Cell, error) {
	if f == nil {
		return nil, invalidWorkbookError()
	}
	if cfg.sheet != "" {
		return readSheetCellsWithConfig(f, cfg.sheet, cfg)
	}
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, ErrNoSheet
	}
	return readSheetCellsWithConfig(f, sheets[0], cfg)
}

func readSheetCellsWithConfig(f *excelize.File, sheet string, cfg readConfig) ([][]Cell, error) {
	if f == nil {
		return nil, invalidWorkbookError()
	}
	if err := ValidateSheetName(sheet); err != nil {
		return nil, err
	}
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, err
	}
	values := limitRows(rows, cfg)
	return buildCells(f, sheet, values, cfg)
}

func limitRows(rows [][]string, cfg readConfig) [][]string {
	if cfg.startRow <= 1 && cfg.startCol <= 1 && cfg.maxRows <= 0 && cfg.maxCols <= 0 {
		return rows
	}

	startRow := cfg.startRow
	if startRow <= 0 {
		startRow = 1
	}
	startCol := cfg.startCol
	if startCol <= 0 {
		startCol = 1
	}
	if startRow > len(rows) {
		return [][]string{}
	}

	out := rows[startRow-1:]
	if cfg.maxRows > 0 && cfg.maxRows < len(out) {
		out = out[:cfg.maxRows]
	}

	startColIndex := startCol - 1
	limited := make([][]string, 0, len(out))
	for _, row := range out {
		if startColIndex >= len(row) {
			limited = append(limited, []string{})
			continue
		}
		cellValues := row[startColIndex:]
		if cfg.maxCols > 0 && cfg.maxCols < len(cellValues) {
			cellValues = cellValues[:cfg.maxCols]
		}
		limited = append(limited, slices.Clone(cellValues))
	}
	return limited
}

func readStartRow(cfg readConfig) int {
	if cfg.startRow > 0 {
		return cfg.startRow
	}
	return 1
}

func readStartCol(cfg readConfig) int {
	if cfg.startCol > 0 {
		return cfg.startCol
	}
	return 1
}

func buildCells(f *excelize.File, sheet string, rows [][]string, cfg readConfig) ([][]Cell, error) {
	startRow := readStartRow(cfg)
	startCol := readStartCol(cfg)
	out := make([][]Cell, 0, len(rows))
	for rowIndex, row := range rows {
		cells := make([]Cell, 0, len(row))
		for colIndex, value := range row {
			rowNumber := startRow + rowIndex
			colNumber := startCol + colIndex
			axis, err := excelize.CoordinatesToCellName(colNumber, rowNumber)
			if err != nil {
				return nil, fmt.Errorf("poi: cell coordinates row=%d col=%d: %w", rowNumber, colNumber, err)
			}
			cellType, err := f.GetCellType(sheet, axis)
			if err != nil {
				return nil, err
			}
			cellType = normalizeCellType(value, cellType)
			cells = append(cells, Cell{
				Value: value,
				Type:  cellType,
				Axis:  axis,
				Row:   rowNumber,
				Col:   colNumber,
			})
		}
		out = append(out, cells)
	}
	return out, nil
}

func normalizeCellType(value string, cellType CellType) CellType {
	if cellType == excelize.CellTypeUnset && value != "" {
		return excelize.CellTypeNumber
	}
	return cellType
}

func replaceDefaultSheet(f *excelize.File, sheet string) error {
	if f == nil {
		return invalidWorkbookError()
	}
	if sheet == DefaultSheetName {
		return nil
	}
	if err := f.SetSheetName(DefaultSheetName, sheet); err != nil {
		return err
	}
	return nil
}

func setRows(f *excelize.File, sheet string, rows [][]string, startRow, startCol int) error {
	if f == nil {
		return invalidWorkbookError()
	}
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

func setAnyRows(f *excelize.File, sheet string, rows [][]any, startRow, startCol int) error {
	if f == nil {
		return invalidWorkbookError()
	}
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
			if err := f.SetCellValue(sheet, cell, value); err != nil {
				return err
			}
		}
	}
	return nil
}

func ensureParentDir(path string, cfg writeConfig) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return cfg.mkdirAll(dir, cfg.dirPerm)
}

func saveWorkbook(f *excelize.File, path string, cfg writeConfig) error {
	if f == nil {
		return invalidWorkbookError()
	}
	if !cfg.overwrite {
		if _, err := cfg.stat(path); err == nil {
			return knifer.WrapError(knifer.ErrCodeInvalidInput, "poi: file already exists", fs.ErrExist)
		} else if !os.IsNotExist(err) {
			return err
		}
	}
	if err := cfg.saveAs(f, path, cfg.saveOptions...); err != nil {
		return err
	}
	if cfg.filePerm != 0 {
		return cfg.chmod(path, cfg.filePerm)
	}
	return nil
}
