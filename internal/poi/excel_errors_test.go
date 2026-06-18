package poi

import (
	"errors"
	"path/filepath"
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
}
