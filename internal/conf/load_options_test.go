package conf

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestLoadWithOptionsPassesParseOptions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.yaml")
	if err := os.WriteFile(path, []byte("ignored"), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := LoadWithOptions(path, LoadOptions{ParseOptions: []ParseOption{WithYAMLUnmarshalFunc(func([]byte, any) error {
		return errors.New("custom yaml error")
	})}})
	if err == nil {
		t.Fatalf("LoadWithOptions = %#v, nil error", c)
	}
	assertConfCode(t, err, knifer.ErrCodeInvalidInput)
}
