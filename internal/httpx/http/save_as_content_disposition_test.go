package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestSaveAsViaContentDisposition(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Disposition", `attachment; filename="real.bin"`)
		_, _ = w.Write([]byte("from-cd"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	resp := Get(srv.URL).Execute()
	if _, err := resp.SaveAs(dir); err != nil {
		t.Fatalf("err: %v", err)
	}
	target := filepath.Join(dir, "real.bin")
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("not found: %v", err)
	}
	if !strings.Contains(string(data), "from-cd") {
		t.Fatalf("content: %q", string(data))
	}
}

func TestSaveAsRejectsUnsafeContentDispositionFilename(t *testing.T) {
	tests := []string{
		`attachment; filename="../outside"`,
		`attachment; filename="..\outside"`,
		`attachment; filename="/tmp/outside"`,
	}
	for _, cd := range tests {
		t.Run(cd, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Disposition", cd)
				_, _ = w.Write([]byte("unsafe"))
			}))
			defer srv.Close()

			dir := t.TempDir()
			_, err := Get(srv.URL).Execute().SaveAs(dir)
			if !errors.Is(err, knifer.ErrCodeInvalidInput) {
				t.Fatalf("SaveAs error = %v, want invalid input", err)
			}
			if _, statErr := os.Stat(filepath.Join(dir, "outside")); !errors.Is(statErr, os.ErrNotExist) {
				t.Fatalf("unsafe file should not be created, stat err = %v", statErr)
			}
		})
	}
}
