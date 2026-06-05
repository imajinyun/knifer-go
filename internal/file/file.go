// Package file provides file and IO helpers.
package file

import (
	"bufio"
	"bytes"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// This section provides IO helpers aligned with the utility toolkit-core IoUtil.

type writeConfig struct {
	filePerm      fs.FileMode
	dirPerm       fs.FileMode
	overwrite     bool
	createParents bool
}

// WriteOption customizes file write helpers.
type WriteOption func(*writeConfig)

type dirConfig struct {
	dirPerm fs.FileMode
}

// DirOption customizes directory helpers.
type DirOption func(*dirConfig)

type readConfig struct {
	maxBytes         int64
	initialLineBytes int
	maxLineBytes     int
}

// ReadOption customizes file and stream read helpers.
type ReadOption func(*readConfig)

func defaultWriteConfig() writeConfig {
	return writeConfig{filePerm: 0o644, dirPerm: 0o755, overwrite: true, createParents: true}
}

func defaultDirConfig() dirConfig { return dirConfig{dirPerm: 0o755} }

func defaultReadConfig() readConfig {
	return readConfig{initialLineBytes: 64 * 1024, maxLineBytes: 1024 * 1024}
}

// WithFilePerm sets the file permission used when creating files.
func WithFilePerm(perm fs.FileMode) WriteOption { return func(c *writeConfig) { c.filePerm = perm } }

// WithDirPerm sets the parent-directory permission used when creating directories.
func WithDirPerm(perm fs.FileMode) WriteOption { return func(c *writeConfig) { c.dirPerm = perm } }

// WithOverwrite controls whether an existing destination file may be replaced.
func WithOverwrite(overwrite bool) WriteOption {
	return func(c *writeConfig) { c.overwrite = overwrite }
}

// WithCreateParents controls whether parent directories are created automatically.
func WithCreateParents(create bool) WriteOption {
	return func(c *writeConfig) { c.createParents = create }
}

// WithMkdirPerm sets the directory permission used by Mkdir.
func WithMkdirPerm(perm fs.FileMode) DirOption { return func(c *dirConfig) { c.dirPerm = perm } }

// WithMaxBytes limits how many bytes a read helper may consume. Non-positive means unlimited.
func WithMaxBytes(n int64) ReadOption { return func(c *readConfig) { c.maxBytes = n } }

// WithInitialLineBuffer sets the initial scanner buffer for line reads.
func WithInitialLineBuffer(n int) ReadOption { return func(c *readConfig) { c.initialLineBytes = n } }

// WithMaxLineBytes sets the maximum scanner token size for line reads.
func WithMaxLineBytes(n int) ReadOption { return func(c *readConfig) { c.maxLineBytes = n } }

func applyWriteOptions(opts []WriteOption) writeConfig {
	cfg := defaultWriteConfig()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return cfg
}

func applyDirOptions(opts []DirOption) dirConfig {
	cfg := defaultDirConfig()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return cfg
}

func applyReadOptions(opts []ReadOption) readConfig {
	cfg := defaultReadConfig()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.initialLineBytes <= 0 {
		cfg.initialLineBytes = 64 * 1024
	}
	if cfg.maxLineBytes <= 0 {
		cfg.maxLineBytes = 1024 * 1024
	}
	if cfg.maxLineBytes < cfg.initialLineBytes {
		cfg.maxLineBytes = cfg.initialLineBytes
	}
	return cfg
}

// ReadAll reads all data from r.
func ReadAll(r io.Reader) ([]byte, error) { return ReadAllWithOptions(r) }

// ReadAllWithOptions reads data from r with per-call read options.
func ReadAllWithOptions(r io.Reader, opts ...ReadOption) ([]byte, error) {
	return readAllLimit(r, applyReadOptions(opts).maxBytes)
}

// ReadString reads all data from r and returns it as a string.
func ReadString(r io.Reader) (string, error) { return ReadStringWithOptions(r) }

// ReadStringWithOptions reads data from r as a string with per-call read options.
func ReadStringWithOptions(r io.Reader, opts ...ReadOption) (string, error) {
	b, err := ReadAllWithOptions(r, opts...)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ReadLines reads all lines from r. The scanner buffer is enlarged to support lines up to 1 MiB.
func ReadLines(r io.Reader) ([]string, error) { return ReadLinesWithOptions(r) }

// ReadLinesWithOptions reads all lines from r with per-call line options.
func ReadLinesWithOptions(r io.Reader, opts ...ReadOption) ([]string, error) {
	cfg := applyReadOptions(opts)
	var limited *io.LimitedReader
	if cfg.maxBytes > 0 {
		limited = &io.LimitedReader{R: r, N: cfg.maxBytes + 1}
		r = limited
	}
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, cfg.initialLineBytes), cfg.maxLineBytes)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if limited != nil && limited.N == 0 {
		return nil, invalidInputf("read exceeds max bytes: %d", cfg.maxBytes)
	}
	return lines, nil
}

// IoCopy copies from src to dst and returns the number of bytes written.
func IoCopy(dst io.Writer, src io.Reader) (int64, error) { return io.Copy(dst, src) }

// CloseQuietly closes c and ignores the returned error.
func CloseQuietly(c io.Closer) {
	if c == nil {
		return
	}
	_ = c.Close()
}

// This section provides file and filename helpers aligned with the utility toolkit-core FileUtil and FileNameUtil.

// FileExists reports whether a file or directory exists.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsFile reports whether path exists and is a regular file.
func IsFile(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}

// IsDirectory reports whether path exists and is a directory.
func IsDirectory(path string) bool {
	st, err := os.Stat(path)
	return err == nil && st.IsDir()
}

// FileReadString reads the whole file as a string.
func FileReadString(path string) (string, error) { return FileReadStringWithOptions(path) }

// FileReadStringWithOptions reads a file as a string with per-call read options.
func FileReadStringWithOptions(path string, opts ...ReadOption) (string, error) {
	b, err := FileReadBytesWithOptions(path, opts...)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// FileReadBytes reads all bytes from a file.
func FileReadBytes(path string) ([]byte, error) { return FileReadBytesWithOptions(path) }

// FileReadBytesWithOptions reads bytes from a file with per-call read options.
func FileReadBytesWithOptions(path string, opts ...ReadOption) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, wrapFileIO("read file "+path, err)
	}
	defer CloseQuietly(f)
	b, err := ReadAllWithOptions(f, opts...)
	return b, wrapFileIO("read file "+path, err)
}

// FileReadLines reads all lines from a file.
func FileReadLines(path string) ([]string, error) { return FileReadLinesWithOptions(path) }

// FileReadLinesWithOptions reads all lines from a file with per-call read options.
func FileReadLinesWithOptions(path string, opts ...ReadOption) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, wrapFileIO("open file "+path, err)
	}
	defer CloseQuietly(f)
	lines, err := ReadLinesWithOptions(f, opts...)
	return lines, wrapFileIO("read lines from file "+path, err)
}

// FileWriteString writes content to a file, overwriting existing data and creating parent directories.
func FileWriteString(path, content string, opts ...WriteOption) error {
	return FileWriteBytes(path, []byte(content), opts...)
}

// FileWriteBytes writes bytes to a file, overwriting existing data and creating parent directories.
func FileWriteBytes(path string, data []byte, opts ...WriteOption) error {
	cfg := applyWriteOptions(opts)
	if err := ensureWriteParent(path, cfg); err != nil {
		return err
	}
	flag := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if !cfg.overwrite {
		flag |= os.O_EXCL
	}
	return writeFile(path, data, flag, cfg.filePerm)
}

// FileAppendString appends content to a file and creates parent directories when needed.
func FileAppendString(path, content string, opts ...WriteOption) error {
	cfg := applyWriteOptions(opts)
	if err := ensureWriteParent(path, cfg); err != nil {
		return err
	}
	flag := os.O_APPEND | os.O_CREATE | os.O_WRONLY
	if !cfg.overwrite {
		flag |= os.O_EXCL
	}
	f, err := os.OpenFile(path, flag, cfg.filePerm)
	if err != nil {
		return wrapFileIO("open file "+path, err)
	}
	defer CloseQuietly(f)
	_, err = f.WriteString(content)
	return wrapFileIO("append file "+path, err)
}

// Mkdir creates a directory tree. Empty and current-directory paths are treated as no-ops.
func Mkdir(dir string, opts ...DirOption) error {
	if dir == "" || dir == "." {
		return nil
	}
	cfg := applyDirOptions(opts)
	return wrapFileIO("create directory "+dir, os.MkdirAll(dir, cfg.dirPerm))
}

// Touch creates an empty file when it does not exist.
func Touch(path string, opts ...WriteOption) error {
	if FileExists(path) {
		return nil
	}
	cfg := applyWriteOptions(opts)
	if err := ensureWriteParent(path, cfg); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, cfg.filePerm)
	if err != nil {
		return wrapFileIO("touch file "+path, err)
	}
	return wrapFileIO("close file "+path, f.Close())
}

// Del removes a file or directory recursively. Missing paths are treated as success.
func Del(path string) error {
	if !FileExists(path) {
		return nil
	}
	return wrapFileIO("delete "+path, os.RemoveAll(path))
}

// FileCopy copies a file and overwrites the destination when it already exists.
func FileCopy(src, dst string, opts ...WriteOption) error {
	if !IsFile(src) {
		return invalidInputf("source is not a file: %s", src)
	}
	cfg := applyWriteOptions(opts)
	if err := ensureWriteParent(dst, cfg); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return wrapFileIO("open source file "+src, err)
	}
	defer CloseQuietly(in)
	flag := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if !cfg.overwrite {
		flag |= os.O_EXCL
	}
	out, err := os.OpenFile(dst, flag, cfg.filePerm)
	if err != nil {
		return wrapFileIO("open destination file "+dst, err)
	}
	defer CloseQuietly(out)
	_, err = io.Copy(out, in)
	return wrapFileIO("copy file "+src+" to "+dst, err)
}

func ensureWriteParent(path string, cfg writeConfig) error {
	if !cfg.createParents {
		return nil
	}
	return Mkdir(filepath.Dir(path), WithMkdirPerm(cfg.dirPerm))
}

func writeFile(path string, data []byte, flag int, perm fs.FileMode) error {
	f, err := os.OpenFile(path, flag, perm)
	if err != nil {
		return wrapFileIO("open file "+path, err)
	}
	defer CloseQuietly(f)
	n, err := f.Write(data)
	if err == nil && n < len(data) {
		err = io.ErrShortWrite
	}
	if err1 := f.Close(); err == nil {
		err = err1
	}
	if err != nil {
		return wrapFileIO("write file "+path, err)
	}
	return nil
}

func readAllLimit(r io.Reader, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 {
		return io.ReadAll(r)
	}
	limited := &io.LimitedReader{R: r, N: maxBytes + 1}
	b, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > maxBytes {
		return nil, invalidInputf("read exceeds max bytes: %d", maxBytes)
	}
	return b, nil
}

// MainName returns the file name without its extension; parent directories are ignored.
func MainName(path string) string {
	name := filepath.Base(path)
	ext := filepath.Ext(name)
	if ext == "" {
		return name
	}
	return strings.TrimSuffix(name, ext)
}

// Extension returns the file extension without the leading dot, or an empty string when absent.
func Extension(path string) string {
	ext := filepath.Ext(path)
	if ext == "" {
		return ""
	}
	return ext[1:]
}

// FileSize returns the file size in bytes, or -1 when the path is missing or not a file.
func FileSize(path string) int64 {
	st, err := os.Stat(path)
	if err != nil || st.IsDir() {
		return -1
	}
	return st.Size()
}

// ReaderFromString converts a string to an io.Reader.
func ReaderFromString(s string) io.Reader { return bytes.NewBufferString(s) }
