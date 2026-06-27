// Package file provides file and IO helpers.
package file

import (
	"bufio"
	"bytes"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	strutil "github.com/imajinyun/knifer-go/internal/str"
)

const DefaultMaxBytes int64 = 64 << 20

// This section provides IO helpers aligned with the utility toolkit-core IoUtil.

type (
	OpenFunc      func(string) (io.ReadCloser, error)
	OpenFileFunc  func(string, int, fs.FileMode) (io.WriteCloser, error)
	StatFunc      func(string) (fs.FileInfo, error)
	MkdirAllFunc  func(string, fs.FileMode) error
	RemoveAllFunc func(string) error
)

type fileConfig struct {
	filePerm         fs.FileMode
	dirPerm          fs.FileMode
	overwrite        bool
	createParents    bool
	maxBytes         int64
	bufferSize       int
	initialLineBytes int
	maxLineBytes     int
	open             OpenFunc
	openFile         OpenFileFunc
	stat             StatFunc
	mkdirAll         MkdirAllFunc
	removeAll        RemoveAllFunc
	charset          string
}

// Option customizes file helpers.
type Option func(*fileConfig)

// WriteOption customizes file write helpers.
type WriteOption = Option

// DirOption customizes directory helpers.
type DirOption = Option

// ReadOption customizes file and stream read helpers.
type ReadOption = Option

// StatOption customizes stat-like file helpers.
type StatOption = Option

// DeleteOption customizes delete helpers.
type DeleteOption = Option

func defaultConfig() fileConfig {
	return fileConfig{
		filePerm:         0o644,
		dirPerm:          0o755,
		overwrite:        true,
		createParents:    true,
		maxBytes:         DefaultMaxBytes,
		bufferSize:       32 * 1024,
		initialLineBytes: 64 * 1024,
		maxLineBytes:     1024 * 1024,
		open:             defaultOpen,
		openFile:         defaultOpenFile,
		stat:             os.Stat,
		mkdirAll:         os.MkdirAll,
		removeAll:        os.RemoveAll,
	}
}

func defaultOpen(path string) (io.ReadCloser, error) {
	// #nosec G304 -- file helpers intentionally read caller-provided paths.
	return os.Open(path)
}

func defaultOpenFile(path string, flag int, perm fs.FileMode) (io.WriteCloser, error) {
	// #nosec G304 -- file helpers intentionally write caller-provided paths.
	return os.OpenFile(path, flag, perm)
}

// WithFilePerm sets the file permission used when creating files.
func WithFilePerm(perm fs.FileMode) WriteOption { return func(c *fileConfig) { c.filePerm = perm } }

// WithDirPerm sets the parent-directory permission used when creating directories.
func WithDirPerm(perm fs.FileMode) WriteOption { return func(c *fileConfig) { c.dirPerm = perm } }

// WithOverwrite controls whether an existing destination file may be replaced.
func WithOverwrite(overwrite bool) WriteOption {
	return func(c *fileConfig) { c.overwrite = overwrite }
}

// WithCreateParents controls whether parent directories are created automatically.
func WithCreateParents(create bool) WriteOption {
	return func(c *fileConfig) { c.createParents = create }
}

// WithMkdirPerm sets the directory permission used by Mkdir.
func WithMkdirPerm(perm fs.FileMode) DirOption { return func(c *fileConfig) { c.dirPerm = perm } }

// WithMaxBytes limits how many bytes a read helper may consume. Non-positive restores DefaultMaxBytes.
func WithMaxBytes(n int64) ReadOption {
	return func(c *fileConfig) {
		if n > 0 {
			c.maxBytes = n
		} else {
			c.maxBytes = DefaultMaxBytes
		}
	}
}

// WithUnlimitedRead disables the default read-size guard for callers that explicitly need it.
func WithUnlimitedRead() ReadOption { return func(c *fileConfig) { c.maxBytes = 0 } }

// WithBufferSize sets the buffer size used by chunk reads and limited copies.
func WithBufferSize(n int) ReadOption { return func(c *fileConfig) { c.bufferSize = n } }

// WithInitialLineBuffer sets the initial scanner buffer for line reads.
func WithInitialLineBuffer(n int) ReadOption { return func(c *fileConfig) { c.initialLineBytes = n } }

// WithMaxLineBytes sets the maximum scanner token size for line reads.
func WithMaxLineBytes(n int) ReadOption { return func(c *fileConfig) { c.maxLineBytes = n } }

// WithCharset converts file string and line reads from charset to UTF-8.
func WithCharset(charset string) ReadOption {
	return func(c *fileConfig) { c.charset = charset }
}

// WithOpen sets the function used to open files for reading.
func WithOpen(open OpenFunc) Option {
	return func(c *fileConfig) {
		if open != nil {
			c.open = open
		}
	}
}

// WithOpenFile sets the function used to open files for writing.
func WithOpenFile(openFile OpenFileFunc) Option {
	return func(c *fileConfig) {
		if openFile != nil {
			c.openFile = openFile
		}
	}
}

// WithStat sets the function used to inspect filesystem paths.
func WithStat(stat StatFunc) Option {
	return func(c *fileConfig) {
		if stat != nil {
			c.stat = stat
		}
	}
}

// WithMkdirAll sets the function used to create directory trees.
func WithMkdirAll(mkdirAll MkdirAllFunc) Option {
	return func(c *fileConfig) {
		if mkdirAll != nil {
			c.mkdirAll = mkdirAll
		}
	}
}

// WithRemoveAll sets the function used to remove file trees.
func WithRemoveAll(removeAll RemoveAllFunc) Option {
	return func(c *fileConfig) {
		if removeAll != nil {
			c.removeAll = removeAll
		}
	}
}

func applyOptions(opts []Option) fileConfig {
	cfg := defaultConfig()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.initialLineBytes <= 0 {
		cfg.initialLineBytes = 64 * 1024
	}
	if cfg.bufferSize <= 0 {
		cfg.bufferSize = 32 * 1024
	}
	if cfg.maxLineBytes <= 0 {
		cfg.maxLineBytes = 1024 * 1024
	}
	if cfg.maxLineBytes < cfg.initialLineBytes {
		cfg.maxLineBytes = cfg.initialLineBytes
	}
	if cfg.open == nil {
		cfg.open = defaultOpen
	}
	if cfg.openFile == nil {
		cfg.openFile = defaultOpenFile
	}
	if cfg.stat == nil {
		cfg.stat = os.Stat
	}
	if cfg.mkdirAll == nil {
		cfg.mkdirAll = os.MkdirAll
	}
	if cfg.removeAll == nil {
		cfg.removeAll = os.RemoveAll
	}
	return cfg
}

func applyWriteOptions(opts []WriteOption) fileConfig { return applyOptions(opts) }
func applyDirOptions(opts []DirOption) fileConfig     { return applyOptions(opts) }
func applyReadOptions(opts []ReadOption) fileConfig   { return applyOptions(opts) }
func applyStatOptions(opts []StatOption) fileConfig   { return applyOptions(opts) }
func applyDeleteOptions(opts []DeleteOption) fileConfig {
	return applyOptions(opts)
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
	cfg := applyReadOptions(opts)
	b, err := readAllLimit(r, cfg.maxBytes)
	if err != nil {
		return "", err
	}
	if cfg.charset != "" {
		b, err = decodeCharset(b, cfg.charset)
		if err != nil {
			return "", err
		}
	}
	return string(b), nil
}

// ReadLines reads all lines from r. The scanner buffer is enlarged to support lines up to 1 MiB.
func ReadLines(r io.Reader) ([]string, error) { return ReadLinesWithOptions(r) }

// ReadLinesWithOptions reads all lines from r with per-call line options.
func ReadLinesWithOptions(r io.Reader, opts ...ReadOption) ([]string, error) {
	cfg := applyReadOptions(opts)
	if cfg.charset == "" {
		return readLinesWithConfig(r, cfg)
	}
	data, err := readAllLimit(r, cfg.maxBytes)
	if err != nil {
		return nil, err
	}
	data, err = decodeCharset(data, cfg.charset)
	if err != nil {
		return nil, err
	}
	return readLinesWithConfig(bytes.NewReader(data), cfg)
}

// ReadChunks reads r in chunks and invokes handle for each chunk.
func ReadChunks(r io.Reader, handle func([]byte) error) error {
	return ReadChunksWithOptions(r, handle)
}

// ReadChunksWithOptions reads r in chunks with per-call read options.
func ReadChunksWithOptions(r io.Reader, handle func([]byte) error, opts ...ReadOption) error {
	if r == nil {
		return invalidInputf("reader is nil")
	}
	if handle == nil {
		return invalidInputf("chunk handler is nil")
	}
	return readChunksWithConfig(r, handle, applyReadOptions(opts))
}

// IoCopy copies from src to dst and returns the number of bytes written.
func IoCopy(dst io.Writer, src io.Reader) (int64, error) { return io.Copy(dst, src) }

// IoCopyWithOptions copies from src to dst using per-call read options.
func IoCopyWithOptions(dst io.Writer, src io.Reader, opts ...ReadOption) (int64, error) {
	if dst == nil {
		return 0, invalidInputf("writer is nil")
	}
	if src == nil {
		return 0, invalidInputf("reader is nil")
	}
	return copyWithConfig(dst, src, applyReadOptions(opts))
}

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
	return FileExistsWithOptions(path)
}

// FileExistsWithOptions reports whether a file or directory exists using per-call stat options.
func FileExistsWithOptions(path string, opts ...StatOption) bool {
	_, err := applyStatOptions(opts).stat(path)
	return err == nil
}

// IsFile reports whether path exists and is a regular file.
func IsFile(path string) bool {
	return IsFileWithOptions(path)
}

// IsFileWithOptions reports whether path exists and is a regular file using per-call stat options.
func IsFileWithOptions(path string, opts ...StatOption) bool {
	st, err := applyStatOptions(opts).stat(path)
	return err == nil && !st.IsDir()
}

// IsDirectory reports whether path exists and is a directory.
func IsDirectory(path string) bool {
	return IsDirectoryWithOptions(path)
}

// IsDirectoryWithOptions reports whether path exists and is a directory using per-call stat options.
func IsDirectoryWithOptions(path string, opts ...StatOption) bool {
	st, err := applyStatOptions(opts).stat(path)
	return err == nil && st.IsDir()
}

// FileReadString reads the whole file as a string.
func FileReadString(path string) (string, error) { return FileReadStringWithOptions(path) }

// FileReadStringWithOptions reads a file as a string with per-call read options.
func FileReadStringWithOptions(path string, opts ...ReadOption) (string, error) {
	cfg := applyReadOptions(opts)
	f, err := cfg.open(path)
	if err != nil {
		return "", wrapFileIO("read file "+path, err)
	}
	defer CloseQuietly(f)
	b, err := readAllLimit(f, cfg.maxBytes)
	if err != nil {
		return "", wrapFileIO("read file "+path, err)
	}
	if cfg.charset != "" {
		b, err = decodeCharset(b, cfg.charset)
		if err != nil {
			return "", err
		}
	}
	return string(b), nil
}

// FileReadBytes reads all bytes from a file.
func FileReadBytes(path string) ([]byte, error) { return FileReadBytesWithOptions(path) }

// FileReadBytesWithOptions reads bytes from a file with per-call read options.
func FileReadBytesWithOptions(path string, opts ...ReadOption) ([]byte, error) {
	cfg := applyReadOptions(opts)
	f, err := cfg.open(path)
	if err != nil {
		return nil, wrapFileIO("read file "+path, err)
	}
	defer CloseQuietly(f)
	b, err := readAllLimit(f, cfg.maxBytes)
	return b, wrapFileIO("read file "+path, err)
}

// FileReadLines reads all lines from a file.
func FileReadLines(path string) ([]string, error) { return FileReadLinesWithOptions(path) }

// FileReadLinesWithOptions reads all lines from a file with per-call read options.
func FileReadLinesWithOptions(path string, opts ...ReadOption) ([]string, error) {
	cfg := applyReadOptions(opts)
	f, err := cfg.open(path)
	if err != nil {
		return nil, wrapFileIO("open file "+path, err)
	}
	defer CloseQuietly(f)
	if cfg.charset != "" {
		data, err := readAllLimit(f, cfg.maxBytes)
		if err != nil {
			return nil, wrapFileIO("read file "+path, err)
		}
		data, err = decodeCharset(data, cfg.charset)
		if err != nil {
			return nil, err
		}
		lines, err := readLinesWithConfig(bytes.NewReader(data), cfg)
		return lines, wrapFileIO("read lines from file "+path, err)
	}
	lines, err := readLinesWithConfig(f, cfg)
	return lines, wrapFileIO("read lines from file "+path, err)
}

// FileReadChunks reads a file in chunks and invokes handle for each chunk.
func FileReadChunks(path string, handle func([]byte) error) error {
	return FileReadChunksWithOptions(path, handle)
}

// FileReadChunksWithOptions reads file chunks with per-call read options.
func FileReadChunksWithOptions(path string, handle func([]byte) error, opts ...ReadOption) error {
	if handle == nil {
		return invalidInputf("chunk handler is nil")
	}
	cfg := applyReadOptions(opts)
	f, err := cfg.open(path)
	if err != nil {
		return wrapFileIO("open file "+path, err)
	}
	defer CloseQuietly(f)
	return wrapFileIO("read chunks from file "+path, readChunksWithConfig(f, handle, cfg))
}

// FileWriteString writes content to a file, overwriting existing data and creating parent directories.
func FileWriteString(path, content string, opts ...WriteOption) error {
	return FileWriteStringWithOptions(path, content, opts...)
}

// FileWriteStringWithOptions writes content to a file with per-call write options.
func FileWriteStringWithOptions(path, content string, opts ...WriteOption) error {
	return FileWriteBytesWithOptions(path, []byte(content), opts...)
}

// FileWriteBytes writes bytes to a file, overwriting existing data and creating parent directories.
func FileWriteBytes(path string, data []byte, opts ...WriteOption) error {
	return FileWriteBytesWithOptions(path, data, opts...)
}

// FileWriteBytesWithOptions writes bytes to a file with per-call write options.
func FileWriteBytesWithOptions(path string, data []byte, opts ...WriteOption) error {
	cfg := applyWriteOptions(opts)
	if err := ensureWriteParent(path, cfg); err != nil {
		return err
	}
	flag := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if !cfg.overwrite {
		flag |= os.O_EXCL
	}
	return writeFile(path, data, flag, cfg)
}

// FileAppendString appends content to a file and creates parent directories when needed.
func FileAppendString(path, content string, opts ...WriteOption) error {
	return FileAppendStringWithOptions(path, content, opts...)
}

// FileAppendStringWithOptions appends content to a file with per-call write options.
func FileAppendStringWithOptions(path, content string, opts ...WriteOption) error {
	cfg := applyWriteOptions(opts)
	if err := ensureWriteParent(path, cfg); err != nil {
		return err
	}
	flag := os.O_APPEND | os.O_CREATE | os.O_WRONLY
	if !cfg.overwrite {
		flag |= os.O_EXCL
	}
	f, err := cfg.openFile(path, flag, cfg.filePerm)
	if err != nil {
		return wrapFileIO("open file "+path, err)
	}
	defer CloseQuietly(f)
	_, err = io.WriteString(f, content)
	return wrapFileIO("append file "+path, err)
}

// Mkdir creates a directory tree. Empty and current-directory paths are treated as no-ops.
func Mkdir(dir string, opts ...DirOption) error {
	return MkdirWithOptions(dir, opts...)
}

// MkdirWithOptions creates a directory tree with per-call directory options.
func MkdirWithOptions(dir string, opts ...DirOption) error {
	if dir == "" || dir == "." {
		return nil
	}
	cfg := applyDirOptions(opts)
	return wrapFileIO("create directory "+dir, cfg.mkdirAll(dir, cfg.dirPerm))
}

// Touch creates an empty file when it does not exist.
func Touch(path string, opts ...WriteOption) error {
	return TouchWithOptions(path, opts...)
}

// TouchWithOptions creates an empty file when it does not exist using per-call write options.
func TouchWithOptions(path string, opts ...WriteOption) error {
	cfg := applyWriteOptions(opts)
	if FileExistsWithOptions(path, WithStat(cfg.stat)) {
		return nil
	}
	if err := ensureWriteParent(path, cfg); err != nil {
		return err
	}
	f, err := cfg.openFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, cfg.filePerm)
	if err != nil {
		return wrapFileIO("touch file "+path, err)
	}
	return wrapFileIO("close file "+path, f.Close())
}

// Del removes a file or directory recursively. Missing paths are treated as success.
func Del(path string) error {
	return DelWithOptions(path)
}

// DelWithOptions removes a file or directory recursively using per-call delete options.
func DelWithOptions(path string, opts ...DeleteOption) error {
	cfg := applyDeleteOptions(opts)
	if !FileExistsWithOptions(path, WithStat(cfg.stat)) {
		return nil
	}
	return wrapFileIO("delete "+path, cfg.removeAll(path))
}

// FileCopy copies a file and overwrites the destination when it already exists.
func FileCopy(src, dst string, opts ...WriteOption) error {
	return FileCopyWithOptions(src, dst, opts...)
}

// FileCopyWithOptions copies a file using per-call write options.
func FileCopyWithOptions(src, dst string, opts ...WriteOption) error {
	cfg := applyWriteOptions(opts)
	if !IsFileWithOptions(src, WithStat(cfg.stat)) {
		return invalidInputf("source is not a file: %s", src)
	}
	if err := ensureWriteParent(dst, cfg); err != nil {
		return err
	}
	in, err := cfg.open(src)
	if err != nil {
		return wrapFileIO("open source file "+src, err)
	}
	defer CloseQuietly(in)
	flag := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if !cfg.overwrite {
		flag |= os.O_EXCL
	}
	out, err := cfg.openFile(dst, flag, cfg.filePerm)
	if err != nil {
		return wrapFileIO("open destination file "+dst, err)
	}
	defer CloseQuietly(out)
	_, err = io.Copy(out, in)
	return wrapFileIO("copy file "+src+" to "+dst, err)
}

func ensureWriteParent(path string, cfg fileConfig) error {
	if !cfg.createParents {
		return nil
	}
	dir := filepath.Dir(path)
	if dir == "" || dir == "." {
		return nil
	}
	return wrapFileIO("create directory "+dir, cfg.mkdirAll(dir, cfg.dirPerm))
}

func writeFile(path string, data []byte, flag int, cfg fileConfig) error {
	f, err := cfg.openFile(path, flag, cfg.filePerm)
	if err != nil {
		return wrapFileIO("open file "+path, err)
	}
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

func readLinesWithConfig(r io.Reader, cfg fileConfig) ([]string, error) {
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

func readChunksWithConfig(r io.Reader, handle func([]byte) error, cfg fileConfig) error {
	buf := make([]byte, cfg.bufferSize)
	var total int64
	for {
		n, err := r.Read(buf)
		if n > 0 {
			total += int64(n)
			if cfg.maxBytes > 0 && total > cfg.maxBytes {
				return invalidInputf("read exceeds max bytes: %d", cfg.maxBytes)
			}
			chunk := slices.Clone(buf[:n])
			if err := handle(chunk); err != nil {
				return err
			}
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func copyWithConfig(dst io.Writer, src io.Reader, cfg fileConfig) (int64, error) {
	buf := make([]byte, cfg.bufferSize)
	var written int64
	for {
		n, readErr := src.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			if cfg.maxBytes > 0 && written+int64(n) > cfg.maxBytes {
				allowed := int(cfg.maxBytes - written)
				if allowed > 0 {
					m, writeErr := dst.Write(chunk[:allowed])
					written += int64(m)
					if writeErr != nil {
						return written, writeErr
					}
					if m != allowed {
						return written, io.ErrShortWrite
					}
				}
				return written, invalidInputf("copy exceeds max bytes: %d", cfg.maxBytes)
			}
			m, writeErr := dst.Write(chunk)
			written += int64(m)
			if writeErr != nil {
				return written, writeErr
			}
			if m != n {
				return written, io.ErrShortWrite
			}
		}
		if readErr == io.EOF {
			return written, nil
		}
		if readErr != nil {
			return written, readErr
		}
	}
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
	return FileSizeWithOptions(path)
}

// FileSizeWithOptions returns the file size using per-call stat options.
func FileSizeWithOptions(path string, opts ...StatOption) int64 {
	st, err := applyStatOptions(opts).stat(path)
	if err != nil || st.IsDir() {
		return -1
	}
	return st.Size()
}

// ReaderFromString converts a string to an io.Reader.
func ReaderFromString(s string) io.Reader { return bytes.NewBufferString(s) }

func decodeCharset(data []byte, charset string) ([]byte, error) {
	return strutil.ToUTF8(data, charset)
}
