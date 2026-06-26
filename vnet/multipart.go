package vnet

import (
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"

	netimpl "github.com/imajinyun/knifer-go/internal/net"
)

// NewUploadSetting returns a default upload setting.
func NewUploadSetting() UploadSetting { return netimpl.NewUploadSetting() }

// ParseMultipartForm parses multipart/form-data from an HTTP request.
func ParseMultipartForm(r *http.Request, setting UploadSetting) (*MultipartFormData, error) {
	return netimpl.ParseMultipartForm(r, setting)
}

// WithUploadFilePerm sets the file permission used when creating the destination file.
func WithUploadFilePerm(perm fs.FileMode) UploadSaveOption { return netimpl.WithUploadFilePerm(perm) }

// WithUploadDirPerm sets the directory permission used when creating parent directories.
func WithUploadDirPerm(perm fs.FileMode) UploadSaveOption { return netimpl.WithUploadDirPerm(perm) }

// WithUploadOverwrite controls whether an existing destination file may be replaced.
func WithUploadOverwrite(overwrite bool) UploadSaveOption {
	return netimpl.WithUploadOverwrite(overwrite)
}

// WithUploadCreateParents controls whether parent directories are created automatically.
func WithUploadCreateParents(create bool) UploadSaveOption {
	return netimpl.WithUploadCreateParents(create)
}

// WithUploadMkdirAll sets the directory creator used when saving uploaded files.
func WithUploadMkdirAll(mkdirAll func(string, fs.FileMode) error) UploadSaveOption {
	return netimpl.WithUploadMkdirAll(mkdirAll)
}

// WithUploadOpenSource sets the source opener used when reading uploaded files.
func WithUploadOpenSource(openSource OpenUploadedFileFunc) UploadSaveOption {
	return netimpl.WithUploadOpenSource(openSource)
}

// WithUploadOpenFile sets the file opener used when saving uploaded files.
func WithUploadOpenFile(openFile func(string, int, fs.FileMode) (io.WriteCloser, error)) UploadSaveOption {
	return netimpl.WithUploadOpenFile(openFile)
}

// SaveUploadedFile saves file to destPath.
func SaveUploadedFile(file *multipart.FileHeader, destPath string, opts ...UploadSaveOption) error {
	return netimpl.SaveUploadedFile(file, destPath, opts...)
}

// UploadFileName returns the uploaded file name.
func UploadFileName(file *multipart.FileHeader) string { return netimpl.UploadFileName(file) }

// UploadFileSize returns the uploaded file size.
func UploadFileSize(file *multipart.FileHeader) int64 { return netimpl.UploadFileSize(file) }

// UploadFileContentType returns the uploaded file content type header.
func UploadFileContentType(file *multipart.FileHeader) string {
	return netimpl.UploadFileContentType(file)
}
