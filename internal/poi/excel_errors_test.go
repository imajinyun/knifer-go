package poi

import (
	"bytes"
	"errors"
	"path/filepath"
	"reflect"
	"testing"

	knifer "github.com/imajinyun/go-knifer"
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
