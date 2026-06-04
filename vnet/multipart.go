package vnet

import (
	"io/fs"
	"mime/multipart"
	"net/http"

	netimpl "github.com/imajinyun/go-knifer/internal/net"
)

func NewUploadSetting() UploadSetting { return netimpl.NewUploadSetting() }

func ParseMultipartForm(r *http.Request, setting UploadSetting) (*MultipartFormData, error) {
	return netimpl.ParseMultipartForm(r, setting)
}

func WithUploadFilePerm(perm fs.FileMode) UploadSaveOption { return netimpl.WithUploadFilePerm(perm) }

func WithUploadDirPerm(perm fs.FileMode) UploadSaveOption { return netimpl.WithUploadDirPerm(perm) }

func WithUploadOverwrite(overwrite bool) UploadSaveOption {
	return netimpl.WithUploadOverwrite(overwrite)
}

func WithUploadCreateParents(create bool) UploadSaveOption {
	return netimpl.WithUploadCreateParents(create)
}

func SaveUploadedFile(file *multipart.FileHeader, destPath string, opts ...UploadSaveOption) error {
	return netimpl.SaveUploadedFile(file, destPath, opts...)
}

func UploadFileName(file *multipart.FileHeader) string { return netimpl.UploadFileName(file) }

func UploadFileSize(file *multipart.FileHeader) int64 { return netimpl.UploadFileSize(file) }

func UploadFileContentType(file *multipart.FileHeader) string {
	return netimpl.UploadFileContentType(file)
}
