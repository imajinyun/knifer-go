package vfile

import (
	"io"
	"io/fs"

	fileimpl "github.com/imajinyun/go-knifer/internal/file"
)

// WriteOption customizes file write helpers.
type WriteOption = fileimpl.WriteOption

// DirOption customizes directory helpers.
type DirOption = fileimpl.DirOption

// ReadOption customizes file and stream read helpers.
type ReadOption = fileimpl.ReadOption

// Error is the file module error type.
type Error = fileimpl.FileError

// WithFilePerm sets the file permission used when creating files.
func WithFilePerm(perm fs.FileMode) WriteOption { return fileimpl.WithFilePerm(perm) }

// WithDirPerm sets the parent-directory permission used when creating directories.
func WithDirPerm(perm fs.FileMode) WriteOption { return fileimpl.WithDirPerm(perm) }

// WithOverwrite controls whether an existing destination file may be replaced.
func WithOverwrite(overwrite bool) WriteOption { return fileimpl.WithOverwrite(overwrite) }

// WithCreateParents controls whether parent directories are created automatically.
func WithCreateParents(create bool) WriteOption { return fileimpl.WithCreateParents(create) }

// WithMkdirPerm sets the directory permission used by Mkdir.
func WithMkdirPerm(perm fs.FileMode) DirOption { return fileimpl.WithMkdirPerm(perm) }

// WithMaxBytes limits how many bytes a read helper may consume. Non-positive means unlimited.
func WithMaxBytes(n int64) ReadOption { return fileimpl.WithMaxBytes(n) }

// WithInitialLineBuffer sets the initial scanner buffer for line reads.
func WithInitialLineBuffer(n int) ReadOption { return fileimpl.WithInitialLineBuffer(n) }

// WithMaxLineBytes sets the maximum scanner token size for line reads.
func WithMaxLineBytes(n int) ReadOption { return fileimpl.WithMaxLineBytes(n) }

func ReadAll(r io.Reader) ([]byte, error)              { return fileimpl.ReadAll(r) }
func ReadString(r io.Reader) (string, error)           { return fileimpl.ReadString(r) }
func ReadLines(r io.Reader) ([]string, error)          { return fileimpl.ReadLines(r) }
func Copy(dst io.Writer, src io.Reader) (int64, error) { return fileimpl.IoCopy(dst, src) }
func CloseQuietly(c io.Closer)                         { fileimpl.CloseQuietly(c) }
func Exists(path string) bool                          { return fileimpl.FileExists(path) }
func IsFile(path string) bool                          { return fileimpl.IsFile(path) }
func IsDirectory(path string) bool                     { return fileimpl.IsDirectory(path) }
func ReadFileString(path string) (string, error)       { return fileimpl.FileReadString(path) }
func ReadFileBytes(path string) ([]byte, error)        { return fileimpl.FileReadBytes(path) }
func ReadFileLines(path string) ([]string, error)      { return fileimpl.FileReadLines(path) }

// ReadAllWithOptions reads data from r with per-call read options.
func ReadAllWithOptions(r io.Reader, opts ...ReadOption) ([]byte, error) {
	return fileimpl.ReadAllWithOptions(r, opts...)
}

// ReadStringWithOptions reads data from r as a string with per-call read options.
func ReadStringWithOptions(r io.Reader, opts ...ReadOption) (string, error) {
	return fileimpl.ReadStringWithOptions(r, opts...)
}

// ReadLinesWithOptions reads all lines from r with per-call line options.
func ReadLinesWithOptions(r io.Reader, opts ...ReadOption) ([]string, error) {
	return fileimpl.ReadLinesWithOptions(r, opts...)
}

// ReadFileStringWithOptions reads a file as a string with per-call read options.
func ReadFileStringWithOptions(path string, opts ...ReadOption) (string, error) {
	return fileimpl.FileReadStringWithOptions(path, opts...)
}

// ReadFileBytesWithOptions reads bytes from a file with per-call read options.
func ReadFileBytesWithOptions(path string, opts ...ReadOption) ([]byte, error) {
	return fileimpl.FileReadBytesWithOptions(path, opts...)
}

// ReadFileLinesWithOptions reads all lines from a file with per-call read options.
func ReadFileLinesWithOptions(path string, opts ...ReadOption) ([]string, error) {
	return fileimpl.FileReadLinesWithOptions(path, opts...)
}

// WriteFileString writes content to path, creating parent directories by default.
func WriteFileString(path, content string, opts ...WriteOption) error {
	return fileimpl.FileWriteString(path, content, opts...)
}

// WriteFileBytes writes data to path, creating parent directories by default.
func WriteFileBytes(path string, data []byte, opts ...WriteOption) error {
	return fileimpl.FileWriteBytes(path, data, opts...)
}

// AppendFileString appends content to path, creating parent directories by default.
func AppendFileString(path, content string, opts ...WriteOption) error {
	return fileimpl.FileAppendString(path, content, opts...)
}

// Mkdir creates dir with directory options.
func Mkdir(dir string, opts ...DirOption) error { return fileimpl.Mkdir(dir, opts...) }

// Touch creates path when missing and updates its timestamp.
func Touch(path string, opts ...WriteOption) error {
	return fileimpl.Touch(path, opts...)
}

// Del removes path recursively.
func Del(path string) error { return fileimpl.Del(path) }

// CopyFile copies src to dst, creating destination parents by default.
func CopyFile(src, dst string, opts ...WriteOption) error {
	return fileimpl.FileCopy(src, dst, opts...)
}
func MainName(path string) string         { return fileimpl.MainName(path) }
func Extension(path string) string        { return fileimpl.Extension(path) }
func Size(path string) int64              { return fileimpl.FileSize(path) }
func ReaderFromString(s string) io.Reader { return fileimpl.ReaderFromString(s) }
