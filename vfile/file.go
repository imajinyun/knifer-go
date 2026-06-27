package vfile

import (
	"io"
	"io/fs"

	fileimpl "github.com/imajinyun/knifer-go/internal/file"
)

// WriteOption customizes file write helpers.
type WriteOption = fileimpl.WriteOption

// DirOption customizes directory helpers.
type DirOption = fileimpl.DirOption

// ReadOption customizes file and stream read helpers.
type ReadOption = fileimpl.ReadOption

// StatOption customizes stat-like file helpers.
type StatOption = fileimpl.StatOption

// DeleteOption customizes delete helpers.
type DeleteOption = fileimpl.DeleteOption

type (
	OpenFunc      = fileimpl.OpenFunc
	OpenFileFunc  = fileimpl.OpenFileFunc
	StatFunc      = fileimpl.StatFunc
	MkdirAllFunc  = fileimpl.MkdirAllFunc
	RemoveAllFunc = fileimpl.RemoveAllFunc
)

const DefaultMaxBytes = fileimpl.DefaultMaxBytes

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

// WithMaxBytes limits how many bytes a read helper may consume. Non-positive restores the default limit.
func WithMaxBytes(n int64) ReadOption { return fileimpl.WithMaxBytes(n) }

// WithUnlimitedRead disables the default read-size guard for callers that explicitly need it.
func WithUnlimitedRead() ReadOption { return fileimpl.WithUnlimitedRead() }

// WithBufferSize sets the buffer size used by chunk reads and limited copies.
func WithBufferSize(n int) ReadOption { return fileimpl.WithBufferSize(n) }

// WithInitialLineBuffer sets the initial scanner buffer for line reads.
func WithInitialLineBuffer(n int) ReadOption { return fileimpl.WithInitialLineBuffer(n) }

// WithMaxLineBytes sets the maximum scanner token size for line reads.
func WithMaxLineBytes(n int) ReadOption { return fileimpl.WithMaxLineBytes(n) }

// WithCharset converts file string and line reads from charset to UTF-8.
func WithCharset(charset string) ReadOption { return fileimpl.WithCharset(charset) }

// WithOpen sets the function used to open files for reading.
func WithOpen(open OpenFunc) ReadOption { return fileimpl.WithOpen(open) }

// WithOpenFile sets the function used to open files for writing.
func WithOpenFile(openFile OpenFileFunc) WriteOption { return fileimpl.WithOpenFile(openFile) }

// WithStat sets the function used to inspect filesystem paths.
func WithStat(stat StatFunc) StatOption { return fileimpl.WithStat(stat) }

// WithMkdirAll sets the function used to create directory trees.
func WithMkdirAll(mkdirAll MkdirAllFunc) DirOption { return fileimpl.WithMkdirAll(mkdirAll) }

// WithRemoveAll sets the function used to remove file trees.
func WithRemoveAll(removeAll RemoveAllFunc) DeleteOption { return fileimpl.WithRemoveAll(removeAll) }

// ReadAll reads all data from r using the default read-size guard.
func ReadAll(r io.Reader) ([]byte, error) { return ReadAllWithOptions(r) }

// ReadString reads all data from r as a string using the default read-size guard.
func ReadString(r io.Reader) (string, error) { return ReadStringWithOptions(r) }

// ReadLines reads all lines from r using the default scanner limits.
func ReadLines(r io.Reader) ([]string, error) { return ReadLinesWithOptions(r) }

// ReadChunks reads r in buffered chunks and invokes handle for each chunk.
func ReadChunks(r io.Reader, handle func([]byte) error) error {
	return ReadChunksWithOptions(r, handle)
}
func Copy(dst io.Writer, src io.Reader) (int64, error) { return fileimpl.IoCopy(dst, src) }
func CopyWithOptions(dst io.Writer, src io.Reader, opts ...ReadOption) (int64, error) {
	return fileimpl.IoCopyWithOptions(dst, src, opts...)
}
func CloseQuietly(c io.Closer) { fileimpl.CloseQuietly(c) }

// Exists reports whether a file or directory exists at path.
func Exists(path string) bool { return ExistsWithOptions(path) }

// IsFile reports whether path exists and is a regular file.
func IsFile(path string) bool { return IsFileWithOptions(path) }

// IsDirectory reports whether path exists and is a directory.
func IsDirectory(path string) bool { return IsDirectoryWithOptions(path) }

// ReadFileString reads the whole file at path as a string.
func ReadFileString(path string) (string, error) { return ReadFileStringWithOptions(path) }

// ReadFileBytes reads all bytes from the file at path.
func ReadFileBytes(path string) ([]byte, error) { return ReadFileBytesWithOptions(path) }

// ReadFileLines reads all lines from the file at path using the default scanner limits.
func ReadFileLines(path string) ([]string, error) { return ReadFileLinesWithOptions(path) }

// ReadFileChunks reads the file at path in buffered chunks and invokes handle for each chunk.
func ReadFileChunks(path string, handle func([]byte) error) error {
	return ReadFileChunksWithOptions(path, handle)
}

// ExistsWithOptions reports whether a file or directory exists using per-call stat options.
func ExistsWithOptions(path string, opts ...StatOption) bool {
	return fileimpl.FileExistsWithOptions(path, opts...)
}

// IsFileWithOptions reports whether path exists and is a regular file using per-call stat options.
func IsFileWithOptions(path string, opts ...StatOption) bool {
	return fileimpl.IsFileWithOptions(path, opts...)
}

// IsDirectoryWithOptions reports whether path exists and is a directory using per-call stat options.
func IsDirectoryWithOptions(path string, opts ...StatOption) bool {
	return fileimpl.IsDirectoryWithOptions(path, opts...)
}

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

// ReadChunksWithOptions reads r in chunks with per-call read options.
func ReadChunksWithOptions(r io.Reader, handle func([]byte) error, opts ...ReadOption) error {
	return fileimpl.ReadChunksWithOptions(r, handle, opts...)
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

// ReadFileChunksWithOptions reads a file in chunks with per-call read options.
func ReadFileChunksWithOptions(path string, handle func([]byte) error, opts ...ReadOption) error {
	return fileimpl.FileReadChunksWithOptions(path, handle, opts...)
}

// WriteFileString writes content to path, creating parent directories by default.
func WriteFileString(path, content string, opts ...WriteOption) error {
	return WriteFileStringWithOptions(path, content, opts...)
}

// WriteFileStringWithOptions writes content to path with per-call write options.
func WriteFileStringWithOptions(path, content string, opts ...WriteOption) error {
	return fileimpl.FileWriteStringWithOptions(path, content, opts...)
}

// WriteFileBytes writes data to path, creating parent directories by default.
func WriteFileBytes(path string, data []byte, opts ...WriteOption) error {
	return WriteFileBytesWithOptions(path, data, opts...)
}

// WriteFileBytesWithOptions writes data to path with per-call write options.
func WriteFileBytesWithOptions(path string, data []byte, opts ...WriteOption) error {
	return fileimpl.FileWriteBytesWithOptions(path, data, opts...)
}

// AppendFileString appends content to path, creating parent directories by default.
func AppendFileString(path, content string, opts ...WriteOption) error {
	return AppendFileStringWithOptions(path, content, opts...)
}

// AppendFileStringWithOptions appends content to path with per-call write options.
func AppendFileStringWithOptions(path, content string, opts ...WriteOption) error {
	return fileimpl.FileAppendStringWithOptions(path, content, opts...)
}

// Mkdir creates dir with directory options.
func Mkdir(dir string, opts ...DirOption) error { return MkdirWithOptions(dir, opts...) }

// MkdirWithOptions creates dir with per-call directory options.
func MkdirWithOptions(dir string, opts ...DirOption) error {
	return fileimpl.MkdirWithOptions(dir, opts...)
}

// Touch creates path when missing and updates its timestamp.
func Touch(path string, opts ...WriteOption) error {
	return TouchWithOptions(path, opts...)
}

// TouchWithOptions creates path when missing using per-call write options.
func TouchWithOptions(path string, opts ...WriteOption) error {
	return fileimpl.TouchWithOptions(path, opts...)
}

// Del removes path recursively.
func Del(path string) error { return DelWithOptions(path) }

// DelWithOptions removes path recursively using per-call delete options.
func DelWithOptions(path string, opts ...DeleteOption) error {
	return fileimpl.DelWithOptions(path, opts...)
}

// CopyFile copies src to dst, creating destination parents by default.
func CopyFile(src, dst string, opts ...WriteOption) error {
	return CopyFileWithOptions(src, dst, opts...)
}

// CopyFileWithOptions copies src to dst using per-call write options.
func CopyFileWithOptions(src, dst string, opts ...WriteOption) error {
	return fileimpl.FileCopyWithOptions(src, dst, opts...)
}
func MainName(path string) string  { return fileimpl.MainName(path) }
func Extension(path string) string { return fileimpl.Extension(path) }

// Size returns the file size in bytes, or -1 when path is missing or not a regular file.
func Size(path string) int64 { return SizeWithOptions(path) }

// SizeWithOptions returns the file size using per-call stat options.
func SizeWithOptions(path string, opts ...StatOption) int64 {
	return fileimpl.FileSizeWithOptions(path, opts...)
}

func ReaderFromString(s string) io.Reader { return fileimpl.ReaderFromString(s) }
