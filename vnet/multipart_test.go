package vnet_test

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/imajinyun/go-knifer/vnet"
)

func TestVNetUploadSaveOptionsFacade(t *testing.T) {
	req := multipartRequest(t, "avatar", "a.txt", "hello")
	form, err := vnet.ParseMultipartForm(req, vnet.NewUploadSetting())
	if err != nil {
		t.Fatalf("ParseMultipartForm: %v", err)
	}
	file := form.GetFile("avatar")
	if file == nil {
		t.Fatal("uploaded file is nil")
	}
	if vnet.UploadFileName(file) != "a.txt" || vnet.UploadFileSize(file) != int64(len("hello")) || vnet.UploadFileContentType(file) == "" {
		t.Fatalf("upload metadata = name:%q size:%d type:%q", vnet.UploadFileName(file), vnet.UploadFileSize(file), vnet.UploadFileContentType(file))
	}

	dir := t.TempDir()
	dest := filepath.Join(dir, "nested", "a.txt")
	if err := vnet.SaveUploadedFile(file, dest, vnet.WithUploadFilePerm(0o600), vnet.WithUploadDirPerm(0o700)); err != nil {
		t.Fatalf("SaveUploadedFile: %v", err)
	}
	info, err := os.Stat(dest)
	if err != nil {
		t.Fatalf("stat saved file: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("saved file perm = %v", info.Mode().Perm())
	}
	if err := vnet.SaveUploadedFile(file, dest, vnet.WithUploadOverwrite(false)); err == nil {
		t.Fatal("SaveUploadedFile should reject overwrite when disabled")
	}
	missingParent := filepath.Join(dir, "missing", "b.txt")
	if err := vnet.SaveUploadedFile(file, missingParent, vnet.WithUploadCreateParents(false)); err == nil {
		t.Fatal("SaveUploadedFile should reject missing parent when parent creation is disabled")
	}

	var mkdirPath string
	var mkdirPerm fs.FileMode
	var openPath string
	var openFlag int
	var openPerm fs.FileMode
	var written bytes.Buffer
	err = vnet.SaveUploadedFile(file, "/virtual/upload/a.txt",
		vnet.WithUploadMkdirAll(func(path string, perm fs.FileMode) error {
			mkdirPath, mkdirPerm = path, perm
			return nil
		}),
		vnet.WithUploadOpenFile(func(path string, flag int, perm fs.FileMode) (io.WriteCloser, error) {
			openPath, openFlag, openPerm = path, flag, perm
			return nopWriteCloser{Writer: &written}, nil
		}),
		vnet.WithUploadDirPerm(0o700), vnet.WithUploadFilePerm(0o600),
	)
	if err != nil {
		t.Fatalf("SaveUploadedFile provider: %v", err)
	}
	if mkdirPath != "/virtual/upload" || mkdirPerm != 0o700 || openPath != "/virtual/upload/a.txt" || openPerm != 0o600 || openFlag&os.O_CREATE == 0 || written.String() != "hello" {
		t.Fatalf("providers mkdir=%q/%v open=%q flag=%#x perm=%v content=%q", mkdirPath, mkdirPerm, openPath, openFlag, openPerm, written.String())
	}
}

func TestVNetUploadOpenSourceFacade(t *testing.T) {
	req := multipartRequest(t, "avatar", "a.txt", "ignored")
	form, err := vnet.ParseMultipartForm(req, vnet.NewUploadSetting())
	if err != nil {
		t.Fatalf("ParseMultipartForm: %v", err)
	}
	file := form.GetFile("avatar")

	var opened *multipart.FileHeader
	var written bytes.Buffer
	err = vnet.SaveUploadedFile(file, "/virtual/upload/a.txt",
		vnet.WithUploadOpenSource(func(f *multipart.FileHeader) (multipart.File, error) {
			opened = f
			return uploadReadCloser{Reader: bytes.NewReader([]byte("from-source-provider"))}, nil
		}),
		vnet.WithUploadMkdirAll(func(string, fs.FileMode) error { return nil }),
		vnet.WithUploadOpenFile(func(string, int, fs.FileMode) (io.WriteCloser, error) {
			return nopWriteCloser{Writer: &written}, nil
		}),
	)
	if err != nil {
		t.Fatalf("SaveUploadedFile with source provider: %v", err)
	}
	if opened != file || written.String() != "from-source-provider" {
		t.Fatalf("source provider opened=%v content=%q", opened == file, written.String())
	}

	wantErr := errors.New("open source")
	err = vnet.SaveUploadedFile(file, "/virtual/upload/a.txt",
		vnet.WithUploadOpenSource(func(*multipart.FileHeader) (multipart.File, error) { return nil, wantErr }),
		vnet.WithUploadMkdirAll(func(string, fs.FileMode) error { return nil }),
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("SaveUploadedFile open source err = %v, want %v", err, wantErr)
	}
}

type nopWriteCloser struct{ io.Writer }

func (w nopWriteCloser) Close() error { return nil }

type uploadReadCloser struct{ io.Reader }

func (r uploadReadCloser) ReadAt(p []byte, off int64) (int, error) {
	if seeker, ok := r.Reader.(io.ReaderAt); ok {
		return seeker.ReadAt(p, off)
	}
	return 0, errors.New("ReadAt unsupported")
}

func (r uploadReadCloser) Seek(offset int64, whence int) (int64, error) {
	if seeker, ok := r.Reader.(io.Seeker); ok {
		return seeker.Seek(offset, whence)
	}
	return 0, errors.New("Seek unsupported")
}

func (r uploadReadCloser) Close() error { return nil }

func multipartRequest(t *testing.T, field, filename, content string) *http.Request {
	t.Helper()
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	part, err := w.CreateFormFile(field, filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write([]byte(content)); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, "/upload", body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}
