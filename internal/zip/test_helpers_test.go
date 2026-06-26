package zip

import (
	"bytes"
	"errors"
	"os"
	"testing"
	"time"

	knifer "github.com/imajinyun/knifer-go"
)

type zipBufferWriteCloser struct {
	*bytes.Buffer
	closed bool
}

func (w *zipBufferWriteCloser) Close() error {
	w.closed = true
	return nil
}

type zipFakeFileInfo struct {
	name string
	size int64
	dir  bool
}

func (f zipFakeFileInfo) Name() string { return f.name }
func (f zipFakeFileInfo) Size() int64  { return f.size }
func (f zipFakeFileInfo) Mode() os.FileMode {
	if f.dir {
		return os.ModeDir | 0o755
	}
	return 0o644
}
func (f zipFakeFileInfo) ModTime() time.Time { return time.Unix(0, 0) }
func (f zipFakeFileInfo) IsDir() bool        { return f.dir }
func (f zipFakeFileInfo) Sys() any           { return nil }

func assertZipCode(t *testing.T, err error, code knifer.ErrCode) {
	t.Helper()
	if err == nil {
		t.Fatalf("err = nil, want %s", code)
	}
	if !errors.Is(err, code) {
		t.Fatalf("errors.Is(%v, %s) = false", err, code)
	}
	got, ok := knifer.CodeOf(err)
	if !ok || got != code {
		t.Fatalf("CodeOf(%v) = %q, %v; want %q, true", err, got, ok, code)
	}
}
