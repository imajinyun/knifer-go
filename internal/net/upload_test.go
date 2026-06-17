package net

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"os"
	"reflect"
	"testing"
)

func TestMultipartFileExts(t *testing.T) {
	req := multipartAvatarRequest(t, "a.txt")
	setting := NewUploadSetting()
	setting.FileExts = []string{".jpg"}
	setting.AllowFileExts = true
	if _, err := ParseMultipartForm(req, setting); err == nil {
		t.Fatal("ParseMultipartForm should reject extension outside allow list")
	}

	req = multipartAvatarRequest(t, "a.txt")
	setting.FileExts = []string{"txt"}
	setting.AllowFileExts = true
	if _, err := ParseMultipartForm(req, setting); err != nil {
		t.Fatalf("ParseMultipartForm should accept allowed extension: %v", err)
	}

	req = multipartAvatarRequest(t, "a.exe")
	setting.FileExts = []string{".exe"}
	setting.AllowFileExts = false
	if _, err := ParseMultipartForm(req, setting); err == nil {
		t.Fatal("ParseMultipartForm should reject extension in deny list")
	}
}

func TestSaveUploadedFileProviderOptions(t *testing.T) {
	req := multipartAvatarRequest(t, "a.txt")
	form, err := ParseMultipartForm(req, NewUploadSetting())
	if err != nil {
		t.Fatalf("ParseMultipartForm: %v", err)
	}
	file := form.GetFile("avatar")
	if file == nil {
		t.Fatal("uploaded file is nil")
	}

	var mkdirPath string
	var mkdirPerm fs.FileMode
	var openPath string
	var openFlag int
	var openPerm fs.FileMode
	var written bytes.Buffer
	err = SaveUploadedFile(file, "/virtual/upload/a.txt",
		WithUploadMkdirAll(func(path string, perm fs.FileMode) error {
			mkdirPath, mkdirPerm = path, perm
			return nil
		}),
		WithUploadOpenFile(func(path string, flag int, perm fs.FileMode) (io.WriteCloser, error) {
			openPath, openFlag, openPerm = path, flag, perm
			return nopWriteCloser{Writer: &written}, nil
		}),
		WithUploadDirPerm(0o700), WithUploadFilePerm(0o600),
	)
	if err != nil {
		t.Fatalf("SaveUploadedFile provider: %v", err)
	}
	if mkdirPath != "/virtual/upload" || mkdirPerm != 0o700 || openPath != "/virtual/upload/a.txt" || openPerm != 0o600 || openFlag&os.O_CREATE == 0 || written.String() != "hello" {
		t.Fatalf("providers mkdir=%q/%v open=%q flag=%#x perm=%v content=%q", mkdirPath, mkdirPerm, openPath, openFlag, openPerm, written.String())
	}
}

func TestMultipartFormAccessors(t *testing.T) {
	req := multipartAvatarRequest(t, "avatar.txt")
	form, err := ParseMultipartForm(req, NewUploadSetting())
	if err != nil {
		t.Fatalf("ParseMultipartForm: %v", err)
	}
	form.Form.Value["name"] = []string{"alice", "bob"}
	form.Form.Value["empty"] = nil

	if !form.IsLoaded() || (*MultipartFormData)(nil).IsLoaded() {
		t.Fatalf("IsLoaded mismatch")
	}
	if got := form.GetParam("name"); got != "alice" {
		t.Fatalf("GetParam = %q", got)
	}
	if got := form.GetParam("empty"); got != "" {
		t.Fatalf("GetParam(empty) = %q", got)
	}
	if got := form.GetArrayParam("name"); !reflect.DeepEqual(got, []string{"alice", "bob"}) {
		t.Fatalf("GetArrayParam = %#v", got)
	}
	if got := form.GetListParam("missing"); got != nil {
		t.Fatalf("GetListParam(missing) = %#v", got)
	}
	if names := form.GetParamNames(); len(names) != 2 {
		t.Fatalf("GetParamNames len = %d", len(names))
	}
	gotParamMap := form.GetParamMap()
	if _, ok := gotParamMap["empty"]; gotParamMap["name"] != "alice" || ok {
		t.Fatalf("GetParamMap = %#v", gotParamMap)
	}
	if got := form.GetParamListMap(); !reflect.DeepEqual(got["name"], []string{"alice", "bob"}) {
		t.Fatalf("GetParamListMap = %#v", got)
	}
	if files := form.GetFiles("avatar"); len(files) != 1 || files[0].Filename != "avatar.txt" {
		t.Fatalf("GetFiles = %#v", files)
	}
	if names := form.GetFileParamNames(); len(names) != 1 || names[0] != "avatar" {
		t.Fatalf("GetFileParamNames = %#v", names)
	}
	if got := form.GetFileMap(); got["avatar"] == nil || got["avatar"].Filename != "avatar.txt" {
		t.Fatalf("GetFileMap = %#v", got)
	}
	if got := form.GetFileListValueMap(); len(got["avatar"]) != 1 {
		t.Fatalf("GetFileListValueMap = %#v", got)
	}
	file := form.GetFile("avatar")
	if UploadFileName(file) != "avatar.txt" || UploadFileSize(file) != 5 || UploadFileContentType(file) == "" {
		t.Fatalf("file metadata name=%q size=%d type=%q", UploadFileName(file), UploadFileSize(file), UploadFileContentType(file))
	}
	if UploadFileName(nil) != "" || UploadFileSize(nil) != 0 || UploadFileContentType(nil) != "" {
		t.Fatalf("nil file metadata should be empty")
	}
}

func TestMultipartFormNilAccessors(t *testing.T) {
	var form *MultipartFormData
	if form.GetParam("x") != "" || form.GetFile("x") != nil {
		t.Fatalf("nil form scalar accessors should be empty")
	}
	if form.GetParamNames() != nil || form.GetListParam("x") != nil || form.GetFileList("x") != nil || form.GetFileParamNames() != nil {
		t.Fatalf("nil form list accessors should be nil")
	}
	if len(form.GetParamMap()) != 0 || len(form.GetParamListMap()) != 0 || len(form.GetFileMap()) != 0 || len(form.GetFileListValueMap()) != 0 {
		t.Fatalf("nil form maps should be empty")
	}
}

func TestSaveUploadedFileOptionBoundaries(t *testing.T) {
	req := multipartAvatarRequest(t, "a.txt")
	form, err := ParseMultipartForm(req, NewUploadSetting())
	if err != nil {
		t.Fatalf("ParseMultipartForm: %v", err)
	}
	file := form.GetFile("avatar")

	if err := SaveUploadedFile(nil, "/ignored"); err != nil {
		t.Fatalf("SaveUploadedFile(nil) = %v", err)
	}

	var mkdirCalled bool
	var openFlag int
	var written bytes.Buffer
	err = SaveUploadedFile(file, "leaf.txt",
		WithUploadCreateParents(false),
		WithUploadOverwrite(false),
		WithUploadMkdirAll(func(string, fs.FileMode) error {
			mkdirCalled = true
			return nil
		}),
		WithUploadOpenFile(func(_ string, flag int, _ fs.FileMode) (io.WriteCloser, error) {
			openFlag = flag
			return nopWriteCloser{Writer: &written}, nil
		}),
	)
	if err != nil {
		t.Fatalf("SaveUploadedFile no-parent/no-overwrite: %v", err)
	}
	if mkdirCalled || openFlag&os.O_EXCL == 0 || written.String() != "hello" {
		t.Fatalf("mkdirCalled=%v flag=%#x written=%q", mkdirCalled, openFlag, written.String())
	}

	wantErr := errors.New("open source")
	err = SaveUploadedFile(file, "leaf.txt", WithUploadOpenSource(func(*multipart.FileHeader) (multipart.File, error) { return nil, wantErr }))
	if !errors.Is(err, wantErr) {
		t.Fatalf("SaveUploadedFile open source error = %v", err)
	}
}

type nopWriteCloser struct{ io.Writer }

func (w nopWriteCloser) Close() error { return nil }

func multipartAvatarRequest(t *testing.T, filename string) *http.Request {
	t.Helper()
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	part, err := w.CreateFormFile("avatar", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write([]byte("hello")); err != nil {
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
