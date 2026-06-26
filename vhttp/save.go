package vhttp

import (
	"io"
	"io/fs"
	"os"

	httpx "github.com/imajinyun/knifer-go/internal/httpx/http"
)

// WithSaveFilePerm sets the file permission used when creating the destination file.
func WithSaveFilePerm(perm fs.FileMode) SaveOption { return httpx.WithSaveFilePerm(perm) }

// WithSaveDirPerm sets the directory permission used when creating parent directories.
func WithSaveDirPerm(perm fs.FileMode) SaveOption { return httpx.WithSaveDirPerm(perm) }

// WithSaveOverwrite controls whether an existing destination file may be replaced.
func WithSaveOverwrite(overwrite bool) SaveOption { return httpx.WithSaveOverwrite(overwrite) }

// WithSaveCreateParents controls whether parent directories are created automatically.
func WithSaveCreateParents(create bool) SaveOption { return httpx.WithSaveCreateParents(create) }

// WithSaveDefaultFilename sets the fallback file name used when dest is a directory.
func WithSaveDefaultFilename(name string) SaveOption { return httpx.WithSaveDefaultFilename(name) }

// WithSaveStat sets the stat provider used to resolve directory destinations.
func WithSaveStat(stat func(string) (os.FileInfo, error)) SaveOption { return httpx.WithSaveStat(stat) }

// WithSaveMkdirAll sets the directory creator used when saving responses.
func WithSaveMkdirAll(mkdirAll func(string, fs.FileMode) error) SaveOption {
	return httpx.WithSaveMkdirAll(mkdirAll)
}

// WithSaveOpenFile sets the file opener used when saving responses.
func WithSaveOpenFile(openFile func(string, int, fs.FileMode) (io.WriteCloser, error)) SaveOption {
	return httpx.WithSaveOpenFile(openFile)
}

// Download downloads rawURL into w.
func Download(rawURL string, w io.Writer) (int64, error) { return DownloadWithOptions(rawURL, w) }

// DownloadWithOptions downloads rawURL into w with per-request options.
func DownloadWithOptions(rawURL string, w io.Writer, opts ...RequestOption) (int64, error) {
	return httpx.DownloadWithOptions(rawURL, w, opts...)
}

// DownloadSafe downloads rawURL into w with SSRF-oriented safety checks enabled.
func DownloadSafe(rawURL string, w io.Writer, opts ...RequestOption) (int64, error) {
	return httpx.DownloadSafe(rawURL, w, opts...)
}

// DownloadFile downloads rawURL to dest.
func DownloadFile(rawURL, dest string, opts ...SaveOption) (int64, error) {
	return DownloadFileWithOptions(rawURL, dest, nil, opts...)
}

// DownloadFileWithOptions downloads rawURL to dest with per-request and per-save options.
func DownloadFileWithOptions(rawURL, dest string, requestOpts []RequestOption, saveOpts ...SaveOption) (int64, error) {
	return httpx.DownloadFileWithOptions(rawURL, dest, requestOpts, saveOpts...)
}

// DownloadFileSafe downloads rawURL to dest with SSRF-oriented safety checks enabled.
func DownloadFileSafe(rawURL, dest string, opts ...SaveOption) (int64, error) {
	return DownloadFileSafeWithOptions(rawURL, dest, nil, opts...)
}

// DownloadFileSafeWithOptions downloads rawURL to dest with SSRF-oriented safety checks enabled.
func DownloadFileSafeWithOptions(rawURL, dest string, requestOpts []RequestOption, saveOpts ...SaveOption) (int64, error) {
	return httpx.DownloadFileSafeWithOptions(rawURL, dest, requestOpts, saveOpts...)
}
