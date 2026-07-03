package file

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"testing"
	"time"

	knifer "github.com/imajinyun/knifer-go"
)

type bufferWriteCloser struct {
	*bytes.Buffer
	closed bool
}

func (w *bufferWriteCloser) Close() error {
	w.closed = true
	return nil
}

type closeErrorWriteCloser struct {
	io.Writer
	err error
}

func (w closeErrorWriteCloser) Close() error { return w.err }

type fakeFileInfo struct {
	name string
	size int64
	dir  bool
}

func (f fakeFileInfo) Name() string { return f.name }
func (f fakeFileInfo) Size() int64  { return f.size }
func (f fakeFileInfo) Mode() fs.FileMode {
	if f.dir {
		return fs.ModeDir | 0o755
	}
	return 0o644
}
func (f fakeFileInfo) ModTime() time.Time { return time.Unix(0, 0) }
func (f fakeFileInfo) IsDir() bool        { return f.dir }
func (f fakeFileInfo) Sys() any           { return nil }

func assertFileCode(t *testing.T, err error, code knifer.ErrCode) {
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
