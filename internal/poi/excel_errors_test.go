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
