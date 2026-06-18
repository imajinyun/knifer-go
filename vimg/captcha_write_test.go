package vimg_test

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/imajinyun/go-knifer/vimg"
)

func TestFacadeCaptchaWriteToFileOptions(t *testing.T) {
	c := vimg.NewLineCaptchaWithOptions(100, 40, vimg.WithGenerator(fixedGenerator{code: "ABCD"}))
	c.CreateCode()
	path := filepath.Join(t.TempDir(), "nested", "imgx.png")
	if err := c.WriteToFileWithOptions(path, vimg.WithFilePerm(0o600), vimg.WithDirPerm(0o700)); err != nil {
		t.Fatalf("WriteToFileWithOptions: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat captcha file: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("captcha file perm = %o, want 600", got)
	}
	if err := c.WriteToFileWithOptions(path, vimg.WithOverwrite(false)); err == nil {
		t.Fatal("WriteToFileWithOptions should reject overwrite=false for existing file")
	}
}

func TestFacadeCaptchaWriteProviderOptions(t *testing.T) {
	c := vimg.NewLineCaptchaWithOptions(100, 40, vimg.WithGenerator(fixedGenerator{code: "ABCD"}))
	c.CreateCode()

	var mkdirPath string
	var mkdirPerm fs.FileMode
	var openPath string
	var openFlag int
	var openPerm fs.FileMode
	var written bytes.Buffer
	if err := c.WriteToFileWithOptions("/virtual/imgx.png",
		vimg.WithMkdirAll(func(path string, perm fs.FileMode) error {
			mkdirPath, mkdirPerm = path, perm
			return nil
		}),
		vimg.WithOpenFile(func(path string, flag int, perm fs.FileMode) (io.WriteCloser, error) {
			openPath, openFlag, openPerm = path, flag, perm
			return nopWriteCloser{Writer: &written}, nil
		}),
		vimg.WithDirPerm(0o700), vimg.WithFilePerm(0o600),
	); err != nil {
		t.Fatalf("WriteToFileWithOptions provider: %v", err)
	}
	if mkdirPath != "/virtual" || mkdirPerm != 0o700 || openPath != "/virtual/imgx.png" || openPerm != 0o600 || openFlag&os.O_CREATE == 0 || written.Len() == 0 {
		t.Fatalf("providers mkdir=%q/%v open=%q flag=%#x perm=%v bytes=%d", mkdirPath, mkdirPerm, openPath, openFlag, openPerm, written.Len())
	}
}

func TestFacadeCaptchaWriteWithCreateParents(t *testing.T) {
	c := vimg.NewLineCaptchaWithOptions(100, 40, vimg.WithGenerator(fixedGenerator{code: "ABCD"}))
	c.CreateCode()

	opt := vimg.WithCreateParents(true)
	if opt == nil {
		t.Fatal("WithCreateParents returned nil")
	}

	var written bytes.Buffer
	if err := c.WriteToFileWithOptions("/virtual/nested/imgx.png",
		vimg.WithCreateParents(true),
		vimg.WithMkdirAll(func(path string, perm fs.FileMode) error {
			return nil
		}),
		vimg.WithOpenFile(func(path string, flag int, perm fs.FileMode) (io.WriteCloser, error) {
			return nopWriteCloser{Writer: &written}, nil
		}),
	); err != nil {
		t.Fatalf("WriteToFileWithOptions with createParents: %v", err)
	}
	if written.Len() == 0 {
		t.Fatal("expected written content with createParents")
	}
}
