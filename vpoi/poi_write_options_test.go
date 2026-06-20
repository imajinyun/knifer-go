package vpoi_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	knifer "github.com/imajinyun/go-knifer"
	"github.com/imajinyun/go-knifer/vpoi"
)

func TestExcelFacadeWriteOptions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "book.xlsx")
	rows := [][]string{{"name", "score"}, {"go", "100"}}
	if err := vpoi.WriteRows(path, rows, vpoi.WithCreateParents(false)); err == nil {
		t.Fatal("WriteRows should fail when parent directory is missing and WithCreateParents(false) is set")
	}
	if err := vpoi.WriteRows(path, rows, vpoi.WithCreateParents(true), vpoi.WithFilePerm(0o600), vpoi.WithDirPerm(0o700)); err != nil {
		t.Fatalf("WriteRows with options: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat workbook: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("workbook file perm = %o, want 600", got)
	}
	if err := vpoi.WriteRows(path, rows, vpoi.WithOverwrite(false)); err == nil {
		t.Fatal("WriteRows should reject overwrite=false for existing workbook")
	} else {
		if !errors.Is(err, fs.ErrExist) {
			t.Fatalf("WriteRows overwrite error = %v, want fs.ErrExist", err)
		}
		if !errors.Is(err, knifer.ErrCodeInvalidInput) {
			t.Fatalf("WriteRows overwrite error = %v, want ErrCodeInvalidInput", err)
		}
	}

	providerPath := filepath.Join(t.TempDir(), "nested", "provider.xlsx")
	var mkdirPath string
	var mkdirPerm fs.FileMode
	var chmodPath string
	var chmodPerm fs.FileMode
	if err := vpoi.WriteRows(providerPath, rows,
		vpoi.WithMkdirAll(func(path string, perm fs.FileMode) error {
			mkdirPath, mkdirPerm = path, perm
			return os.MkdirAll(path, perm)
		}),
		vpoi.WithChmod(func(path string, perm fs.FileMode) error {
			chmodPath, chmodPerm = path, perm
			return nil
		}),
		vpoi.WithDirPerm(0o700), vpoi.WithFilePerm(0o600),
	); err != nil {
		t.Fatalf("WriteRows provider options: %v", err)
	}
	if mkdirPath != filepath.Dir(providerPath) || mkdirPerm != 0o700 || chmodPath != providerPath || chmodPerm != 0o600 {
		t.Fatalf("providers mkdir=%q/%v chmod=%q/%v", mkdirPath, mkdirPerm, chmodPath, chmodPerm)
	}

	statErr := errors.New("stat denied")
	err = vpoi.WriteRows(providerPath, rows, vpoi.WithOverwrite(false), vpoi.WithStat(func(string) (os.FileInfo, error) {
		return nil, statErr
	}))
	if !errors.Is(err, statErr) {
		t.Fatalf("WriteRows should return custom stat error, got %v", err)
	}
}
