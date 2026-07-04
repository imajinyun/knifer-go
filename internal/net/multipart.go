package net

import (
	"fmt"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// UploadSetting configures multipart form parsing and file saving.
type UploadSetting struct {
	MaxFileSize     int64
	MemoryThreshold int64
	TmpUploadPath   string
	FileExts        []string
	AllowFileExts   bool
}

type uploadSaveConfig struct {
	filePerm      fs.FileMode
	dirPerm       fs.FileMode
	overwrite     bool
	createParents bool
	mkdirAll      func(string, fs.FileMode) error
	openSource    OpenUploadedFileFunc
	openFile      func(string, int, fs.FileMode) (io.WriteCloser, error)
}

// UploadSaveOption customizes uploaded-file saving.
type UploadSaveOption func(*uploadSaveConfig)

// OpenUploadedFileFunc opens an uploaded multipart file for reading.
type OpenUploadedFileFunc func(*multipart.FileHeader) (multipart.File, error)

func defaultUploadSaveConfig() uploadSaveConfig {
	return uploadSaveConfig{filePerm: 0o644, dirPerm: 0o750, overwrite: true, createParents: true, mkdirAll: os.MkdirAll, openSource: defaultOpenUploadedFile, openFile: defaultOpenUploadFile}
}

func defaultOpenUploadedFile(file *multipart.FileHeader) (multipart.File, error) { return file.Open() }

func defaultOpenUploadFile(path string, flag int, perm fs.FileMode) (io.WriteCloser, error) {
	// #nosec G304 -- upload helpers intentionally write to the caller-provided destination path.
	return os.OpenFile(path, flag, perm)
}

// WithUploadFilePerm sets the file permission used when creating the destination file.
func WithUploadFilePerm(perm fs.FileMode) UploadSaveOption {
	return func(c *uploadSaveConfig) { c.filePerm = perm }
}

// WithUploadDirPerm sets the directory permission used when creating parent directories.
func WithUploadDirPerm(perm fs.FileMode) UploadSaveOption {
	return func(c *uploadSaveConfig) { c.dirPerm = perm }
}

// WithUploadOverwrite controls whether an existing destination file may be replaced.
func WithUploadOverwrite(overwrite bool) UploadSaveOption {
	return func(c *uploadSaveConfig) { c.overwrite = overwrite }
}

// WithUploadCreateParents controls whether parent directories are created automatically.
func WithUploadCreateParents(create bool) UploadSaveOption {
	return func(c *uploadSaveConfig) { c.createParents = create }
}

// WithUploadMkdirAll sets the directory creator used when saving uploaded files.
func WithUploadMkdirAll(mkdirAll func(string, fs.FileMode) error) UploadSaveOption {
	return func(c *uploadSaveConfig) {
		if mkdirAll != nil {
			c.mkdirAll = mkdirAll
		}
	}
}

// WithUploadOpenSource sets the source opener used when reading uploaded files.
func WithUploadOpenSource(openSource OpenUploadedFileFunc) UploadSaveOption {
	return func(c *uploadSaveConfig) {
		if openSource != nil {
			c.openSource = openSource
		}
	}
}

// WithUploadOpenFile sets the file opener used when saving uploaded files.
func WithUploadOpenFile(openFile func(string, int, fs.FileMode) (io.WriteCloser, error)) UploadSaveOption {
	return func(c *uploadSaveConfig) {
		if openFile != nil {
			c.openFile = openFile
		}
	}
}

func applyUploadSaveOptions(opts []UploadSaveOption) uploadSaveConfig {
	cfg := defaultUploadSaveConfig()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.mkdirAll == nil {
		cfg.mkdirAll = os.MkdirAll
	}
	if cfg.openSource == nil {
		cfg.openSource = defaultOpenUploadedFile
	}
	if cfg.openFile == nil {
		cfg.openFile = defaultOpenUploadFile
	}
	return cfg
}

// NewUploadSetting returns a default upload setting.
func NewUploadSetting() UploadSetting {
	return UploadSetting{MaxFileSize: 32 << 20, MemoryThreshold: 32 << 20, AllowFileExts: true}
}

// MultipartFormData wraps a parsed multipart form.
type MultipartFormData struct {
	Form   *multipart.Form
	loaded bool
}

// ParseMultipartForm parses multipart/form-data from an HTTP request.
func ParseMultipartForm(r *http.Request, setting UploadSetting) (*MultipartFormData, error) {
	if setting.MaxFileSize <= 0 {
		setting.MaxFileSize = 32 << 20
	}
	if setting.MemoryThreshold <= 0 {
		setting.MemoryThreshold = 32 << 20
	}
	r.Body = http.MaxBytesReader(nil, r.Body, setting.MaxFileSize)        //nolint:bodyclose // request body lifecycle is owned by caller.
	if err := r.ParseMultipartForm(setting.MemoryThreshold); err != nil { //nolint:gosec // request body is bounded by MaxBytesReader above.
		return nil, err
	}
	if err := validateMultipartFileExts(r.MultipartForm, setting); err != nil {
		return nil, err
	}
	return &MultipartFormData{Form: r.MultipartForm, loaded: true}, nil
}

func validateMultipartFileExts(form *multipart.Form, setting UploadSetting) error {
	if form == nil || len(setting.FileExts) == 0 {
		return nil
	}
	exts := make(map[string]struct{}, len(setting.FileExts))
	for _, ext := range setting.FileExts {
		ext = normalizeUploadExt(ext)
		if ext != "" {
			exts[ext] = struct{}{}
		}
	}
	if len(exts) == 0 {
		return nil
	}
	for _, files := range form.File {
		for _, file := range files {
			ext := normalizeUploadExt(filepath.Ext(file.Filename))
			_, matched := exts[ext]
			if setting.AllowFileExts && !matched {
				return fmt.Errorf("uploaded file %q extension %q is not allowed", file.Filename, ext)
			}
			if !setting.AllowFileExts && matched {
				return fmt.Errorf("uploaded file %q extension %q is denied", file.Filename, ext)
			}
		}
	}
	return nil
}

func normalizeUploadExt(ext string) string {
	ext = strings.ToLower(strings.TrimSpace(ext))
	if ext == "" {
		return ""
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	return ext
}

// IsLoaded reports whether the form has been parsed.
func (m *MultipartFormData) IsLoaded() bool { return m != nil && m.loaded }

// GetParam returns the first value for name.
func (m *MultipartFormData) GetParam(name string) string {
	values := m.GetListParam(name)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

// GetParamNames returns parameter names.
func (m *MultipartFormData) GetParamNames() []string {
	if m == nil || m.Form == nil {
		return nil
	}
	out := make([]string, 0, len(m.Form.Value))
	for k := range m.Form.Value {
		out = append(out, k)
	}
	return out
}

// GetArrayParam returns all values for name.
func (m *MultipartFormData) GetArrayParam(name string) []string { return m.GetListParam(name) }

// GetListParam returns all values for name.
func (m *MultipartFormData) GetListParam(name string) []string {
	if m == nil || m.Form == nil {
		return nil
	}
	return m.Form.Value[name]
}

// GetParamMap returns first parameter values.
func (m *MultipartFormData) GetParamMap() map[string]string {
	out := map[string]string{}
	if m == nil || m.Form == nil {
		return out
	}
	for k, values := range m.Form.Value {
		if len(values) > 0 {
			out[k] = values[0]
		}
	}
	return out
}

// GetParamListMap returns all parameter values.
func (m *MultipartFormData) GetParamListMap() map[string][]string {
	if m == nil || m.Form == nil {
		return map[string][]string{}
	}
	return m.Form.Value
}

// GetFile returns the first file for name.
func (m *MultipartFormData) GetFile(name string) *multipart.FileHeader {
	files := m.GetFileList(name)
	if len(files) == 0 {
		return nil
	}
	return files[0]
}

// GetFiles returns all files for name.
func (m *MultipartFormData) GetFiles(name string) []*multipart.FileHeader { return m.GetFileList(name) }

// GetFileList returns all files for name.
func (m *MultipartFormData) GetFileList(name string) []*multipart.FileHeader {
	if m == nil || m.Form == nil {
		return nil
	}
	return m.Form.File[name]
}

// GetFileParamNames returns file parameter names.
func (m *MultipartFormData) GetFileParamNames() []string {
	if m == nil || m.Form == nil {
		return nil
	}
	out := make([]string, 0, len(m.Form.File))
	for k := range m.Form.File {
		out = append(out, k)
	}
	return out
}

// GetFileMap returns first file values.
func (m *MultipartFormData) GetFileMap() map[string]*multipart.FileHeader {
	out := map[string]*multipart.FileHeader{}
	if m == nil || m.Form == nil {
		return out
	}
	for k, files := range m.Form.File {
		if len(files) > 0 {
			out[k] = files[0]
		}
	}
	return out
}

// GetFileListValueMap returns all file values.
func (m *MultipartFormData) GetFileListValueMap() map[string][]*multipart.FileHeader {
	if m == nil || m.Form == nil {
		return map[string][]*multipart.FileHeader{}
	}
	return m.Form.File
}

// SaveUploadedFile saves file to destPath.
func SaveUploadedFile(file *multipart.FileHeader, destPath string, opts ...UploadSaveOption) (err error) {
	if file == nil {
		return nil
	}
	cfg := applyUploadSaveOptions(opts)
	if cfg.createParents {
		if err := cfg.mkdirAll(filepath.Dir(destPath), cfg.dirPerm); err != nil {
			return err
		}
	}
	src, err := cfg.openSource(file)
	if err != nil {
		return err
	}
	defer func() { _ = src.Close() }()
	flag := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if !cfg.overwrite {
		flag |= os.O_EXCL
	}
	dst, err := cfg.openFile(destPath, flag, cfg.filePerm) // #nosec G304 -- caller controls destination path.
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := dst.Close(); err == nil {
			err = closeErr
		}
	}()
	_, err = io.Copy(dst, src)
	return err
}

// UploadFileName returns the uploaded file name.
func UploadFileName(file *multipart.FileHeader) string {
	if file == nil {
		return ""
	}
	return file.Filename
}

// UploadFileSize returns the uploaded file size.
func UploadFileSize(file *multipart.FileHeader) int64 {
	if file == nil {
		return 0
	}
	return file.Size
}

// UploadFileContentType returns the uploaded file content type header.
func UploadFileContentType(file *multipart.FileHeader) string {
	if file == nil {
		return ""
	}
	return file.Header.Get("Content-Type")
}
