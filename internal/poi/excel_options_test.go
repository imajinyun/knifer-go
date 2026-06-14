package poi

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestReadWriteOptions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "book.xlsx")
	rows := [][]string{{"name", "score"}, {"go", "100"}}
	if err := WriteRows(path, rows, WithWriteSheet("Data"), WithStartCell(2, 3), WithFilePerm(0o600)); err != nil {
		t.Fatalf("WriteRows with options: %v", err)
	}
	if err := WriteRows(path, rows, WithWriteSheet("Data"), WithOverwrite(false)); err == nil {
		t.Fatalf("WriteRows should reject overwrite when disabled")
	}
	got, err := ReadRows(path, WithReadSheet("Data"))
	if err != nil {
		t.Fatalf("ReadRows with sheet option: %v", err)
	}
	want := [][]string{nil, {"", "", "name", "score"}, {"", "", "go", "100"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ReadRows with options = %#v, want %#v", got, want)
	}
}

func TestWriteProviderOptions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "book.xlsx")
	var mkdirPath string
	var mkdirPerm fs.FileMode
	var chmodPath string
	var chmodPerm fs.FileMode
	if err := WriteRows(path, [][]string{{"x"}},
		WithMkdirAll(func(path string, perm fs.FileMode) error {
			mkdirPath, mkdirPerm = path, perm
			return os.MkdirAll(path, perm)
		}),
		WithChmod(func(path string, perm fs.FileMode) error {
			chmodPath, chmodPerm = path, perm
			return nil
		}),
		WithDirPerm(0o700), WithFilePerm(0o600),
	); err != nil {
		t.Fatalf("WriteRows provider options: %v", err)
	}
	if mkdirPath != filepath.Dir(path) || mkdirPerm != 0o700 || chmodPath != path || chmodPerm != 0o600 {
		t.Fatalf("providers mkdir=%q/%v chmod=%q/%v", mkdirPath, mkdirPerm, chmodPath, chmodPerm)
	}

	statErr := errors.New("stat denied")
	err := WriteRows(path, [][]string{{"x"}}, WithOverwrite(false), WithStat(func(string) (os.FileInfo, error) {
		return nil, statErr
	}))
	if !errors.Is(err, statErr) {
		t.Fatalf("WriteRows should return custom stat error, got %v", err)
	}
}
