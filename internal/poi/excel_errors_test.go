package poi

import (
	"bytes"
	"errors"
	"io"
	"path/filepath"
	"reflect"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
	"github.com/xuri/excelize/v2"
)

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

func TestInvalidSheetNameContract(t *testing.T) {
	invalidNames := []string{
		"bad/name",
		"bad:name",
		"bad*name",
		"bad?name",
		"bad[name]",
		"12345678901234567890123456789012",
	}
	for _, name := range invalidNames {
		t.Run(name, func(t *testing.T) {
			if IsValidSheetName(name) {
				t.Fatalf("IsValidSheetName(%q) = true, want false", name)
			}
			if err := ValidateSheetName(name); !errors.Is(err, ErrInvalidSheetName) {
				t.Fatalf("ValidateSheetName(%q) error = %v, want ErrInvalidSheetName", name, err)
			}
			if err := ValidateSheetName(name); !errors.Is(err, knifer.ErrCodeInvalidInput) {
				t.Fatalf("ValidateSheetName(%q) error = %v, want ErrCodeInvalidInput", name, err)
			}
		})
	}

	if err := ValidateSheetName(""); !errors.Is(err, ErrEmptySheetName) {
		t.Fatalf("ValidateSheetName empty error = %v, want ErrEmptySheetName", err)
	}
	if err := WriteSheetRows(filepath.Join(t.TempDir(), "book.xlsx"), "bad/name", nil); !errors.Is(err, ErrInvalidSheetName) {
		t.Fatalf("WriteSheetRows invalid sheet error = %v, want ErrInvalidSheetName", err)
	}
	if _, err := WriteRowsToBuffer("bad/name", nil); !errors.Is(err, ErrInvalidSheetName) {
		t.Fatalf("WriteRowsToBuffer invalid sheet error = %v, want ErrInvalidSheetName", err)
	}
}

func TestReadRowsRejectsInvalidSheetNameBeforeReader(t *testing.T) {
	rows, err := ReadRowsFromReader(bytes.NewReader(nil), WithReadSheet("bad/name"))
	if !errors.Is(err, ErrInvalidSheetName) {
		t.Fatalf("ReadRowsFromReader invalid sheet error = %v, want ErrInvalidSheetName", err)
	}
	if rows != nil {
		t.Fatalf("ReadRowsFromReader invalid sheet rows = %#v, want nil", rows)
	}
}

func TestWriteSheetsDeterministicOrder(t *testing.T) {
	path := filepath.Join(t.TempDir(), "book.xlsx")
	sheets := map[string][][]string{
		"Users":  {{"id", "name"}},
		"Orders": {{"id", "total"}},
		"Audit":  {{"event"}},
	}
	if err := WriteSheets(path, sheets); err != nil {
		t.Fatalf("WriteSheets: %v", err)
	}
	got, err := SheetNames(path)
	if err != nil {
		t.Fatalf("SheetNames: %v", err)
	}
	want := []string{"Audit", "Orders", "Users"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SheetNames = %#v, want %#v", got, want)
	}
}

func TestNoSheetMatchesErrCode(t *testing.T) {
	if !errors.Is(ErrNoSheet, knifer.ErrCodeNotFound) {
		t.Fatalf("ErrNoSheet should match ErrCodeNotFound")
	}
}

func TestSentinelErrorCode(t *testing.T) {
	if got := ErrNoSheet.(*sentinel).ErrorCode(); got != knifer.ErrCodeNotFound {
		t.Fatalf("ErrNoSheet.ErrorCode = %v, want %v", got, knifer.ErrCodeNotFound)
	}
	if got := ErrEmptySheetName.(*sentinel).ErrorCode(); got != knifer.ErrCodeInvalidInput {
		t.Fatalf("ErrEmptySheetName.ErrorCode = %v, want %v", got, knifer.ErrCodeInvalidInput)
	}
}

func TestSentinelError(t *testing.T) {
	if got := ErrNoSheet.Error(); got == "" {
		t.Fatal("ErrNoSheet.Error() should not be empty")
	}
	if got := ErrEmptySheetName.Error(); got == "" {
		t.Fatal("ErrEmptySheetName.Error() should not be empty")
	}
	if got := ErrInvalidSheetName.Error(); got == "" {
		t.Fatal("ErrInvalidSheetName.Error() should not be empty")
	}
}

func TestNilWorkbookProviderBoundaries(t *testing.T) {
	openFileNil := WithOpenFileFunc(func(string, ...excelize.Options) (*excelize.File, error) { return nil, nil })
	openReaderNil := WithOpenReaderFunc(func(io.Reader, ...excelize.Options) (*excelize.File, error) { return nil, nil })
	newFileNil := WithNewFileFunc(func() *excelize.File { return nil })

	if _, err := SheetNamesWithOptions("virtual.xlsx", openFileNil); !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("SheetNamesWithOptions nil workbook err = %v", err)
	}
	if _, err := ReadRows("virtual.xlsx", openFileNil); !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("ReadRows nil workbook err = %v", err)
	}
	if _, err := ReadSheetRowsWithOptions("virtual.xlsx", "Data", openFileNil); !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("ReadSheetRowsWithOptions nil workbook err = %v", err)
	}
	if _, err := ReadRowsFromReader(bytes.NewReader(nil), openReaderNil); !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("ReadRowsFromReader nil workbook err = %v", err)
	}
	if err := WriteRows("virtual.xlsx", nil, newFileNil); !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("WriteRows nil workbook err = %v", err)
	}
	if err := WriteSheets("virtual.xlsx", nil, newFileNil); !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("WriteSheets nil workbook err = %v", err)
	}
	if _, err := WriteRowsToBuffer("Data", nil, newFileNil); !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("WriteRowsToBuffer nil workbook err = %v", err)
	}
}
