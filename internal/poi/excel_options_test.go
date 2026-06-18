package poi

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/xuri/excelize/v2"
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

func TestWithOpenFileFunc(t *testing.T) {
	cfg := defaultReadConfig()
	fn := func(path string, opts ...excelize.Options) (*excelize.File, error) { return nil, nil }
	WithOpenFileFunc(fn)(&cfg)
	if cfg.openFile == nil {
		t.Fatal("WithOpenFileFunc did not set openFile")
	}
	// nil should not replace
	WithOpenFileFunc(nil)(&cfg)
	if cfg.openFile == nil {
		t.Fatal("nil WithOpenFileFunc should not clear openFile")
	}
}

func TestWithOpenReaderFunc(t *testing.T) {
	cfg := defaultReadConfig()
	fn := func(r io.Reader, opts ...excelize.Options) (*excelize.File, error) { return nil, nil }
	WithOpenReaderFunc(fn)(&cfg)
	if cfg.openReader == nil {
		t.Fatal("WithOpenReaderFunc did not set openReader")
	}
	// nil should not replace
	WithOpenReaderFunc(nil)(&cfg)
	if cfg.openReader == nil {
		t.Fatal("nil WithOpenReaderFunc should not clear openReader")
	}
}

func TestWithCreateParents(t *testing.T) {
	cfg := writeConfig{}
	WithCreateParents(false)(&cfg)
	if cfg.createParents {
		t.Fatal("WithCreateParents(false) did not set createParents")
	}
}

func TestWithSaveOptions(t *testing.T) {
	cfg := writeConfig{}
	WithSaveOptions(excelize.Options{RawCellValue: true})(&cfg)
	if len(cfg.saveOptions) == 0 || !cfg.saveOptions[0].RawCellValue {
		t.Fatal("WithSaveOptions did not set saveOptions")
	}
}

func TestWithNewFileFunc(t *testing.T) {
	cfg := writeConfig{}
	fn := func() *excelize.File { return nil }
	WithNewFileFunc(fn)(&cfg)
	if cfg.newFile == nil {
		t.Fatal("WithNewFileFunc did not set newFile")
	}
	// nil should not replace
	WithNewFileFunc(nil)(&cfg)
	if cfg.newFile == nil {
		t.Fatal("nil WithNewFileFunc should not clear newFile")
	}
}

func TestWithSaveAsFunc(t *testing.T) {
	cfg := writeConfig{}
	fn := func(f *excelize.File, path string, opts ...excelize.Options) error { return nil }
	WithSaveAsFunc(fn)(&cfg)
	if cfg.saveAs == nil {
		t.Fatal("WithSaveAsFunc did not set saveAs")
	}
}
