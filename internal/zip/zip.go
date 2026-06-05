package zip

import (
	archivezip "archive/zip"
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultBufferSize = 32
	maxInt64          = int64(1<<63 - 1)
)

// FileFilter decides whether a source path should be added to an archive.
type FileFilter func(path string, info os.FileInfo) bool

// Entry describes an archive entry.
type Entry = archivezip.File

// Writer is a ZIP archive writer.
type Writer = archivezip.Writer

// Reader is a ZIP archive reader.
type Reader = archivezip.ReadCloser

// EntryData represents in-memory content to add into a ZIP archive.
type EntryData struct {
	Name string
	Data []byte
}

// StreamEntry represents stream content to add into a ZIP archive.
type StreamEntry struct {
	Name   string
	Reader io.Reader
}

type archiveConfig struct {
	dirPerm           os.FileMode
	filePerm          os.FileMode
	overwrite         bool
	preserveMode      bool
	compressionMethod uint16
	compressionLevel  int
	maxBytes          int64
}

// ArchiveOption customizes ZIP/GZIP/ZLIB archive helpers per call.
type ArchiveOption func(*archiveConfig)

func defaultArchiveConfig() archiveConfig {
	return archiveConfig{
		dirPerm:           0o750,
		filePerm:          0o644,
		overwrite:         true,
		preserveMode:      true,
		compressionMethod: archivezip.Deflate,
		compressionLevel:  flate.DefaultCompression,
	}
}

// WithDirPerm sets the directory permission used when creating archive output/extract directories.
func WithDirPerm(perm os.FileMode) ArchiveOption { return func(c *archiveConfig) { c.dirPerm = perm } }

// WithFilePerm sets the file permission used when archive metadata is not preserved.
func WithFilePerm(perm os.FileMode) ArchiveOption {
	return func(c *archiveConfig) { c.filePerm = perm }
}

// WithOverwrite controls whether an existing output/extracted file may be overwritten.
func WithOverwrite(overwrite bool) ArchiveOption {
	return func(c *archiveConfig) { c.overwrite = overwrite }
}

// WithPreserveMode controls whether extracted files keep mode bits from the archive.
func WithPreserveMode(preserve bool) ArchiveOption {
	return func(c *archiveConfig) { c.preserveMode = preserve }
}

// WithCompressionMethod sets the ZIP compression method used for newly created entries.
func WithCompressionMethod(method uint16) ArchiveOption {
	return func(c *archiveConfig) { c.compressionMethod = method }
}

// WithCompressionLevel sets the deflate compression level used for newly created entries.
func WithCompressionLevel(level int) ArchiveOption {
	return func(c *archiveConfig) { c.compressionLevel = level }
}

// WithMaxBytes limits bytes read from archive entries or decompressed streams. Non-positive means unlimited.
func WithMaxBytes(n int64) ArchiveOption { return func(c *archiveConfig) { c.maxBytes = n } }

func applyArchiveOptions(opts []ArchiveOption) archiveConfig {
	cfg := defaultArchiveConfig()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.dirPerm == 0 {
		cfg.dirPerm = 0o750
	}
	if cfg.filePerm == 0 {
		cfg.filePerm = 0o644
	}
	return cfg
}

// Open opens a ZIP file for reading.
func Open(path string) (*archivezip.ReadCloser, error) { return archivezip.OpenReader(path) }

// NewWriter returns a ZIP writer for out.
func NewWriter(out io.Writer) *archivezip.Writer { return archivezip.NewWriter(out) }

// GetStream returns a reader for entry.
func GetStream(entry *archivezip.File) (io.ReadCloser, error) {
	if entry == nil {
		return nil, invalidInputf("zip entry is nil")
	}
	return entry.Open()
}

// Append appends srcPath into zipPath by rewriting the archive.
func Append(zipPath, srcPath string) error {
	return appendWithFilter(zipPath, srcPath, nil)
}

// Zip creates an archive next to srcPath and returns the archive path.
func Zip(srcPath string) (string, error) {
	dest := strings.TrimSuffix(srcPath, filepath.Ext(srcPath)) + ".zip"
	return dest, ZipTo(srcPath, dest, false)
}

// ZipTo creates an archive at zipPath from srcPath.
func ZipTo(srcPath, zipPath string, withSrcDir bool) error {
	return ZipFiles(zipPath, withSrcDir, srcPath)
}

// ZipFiles creates a ZIP archive from source files or directories.
func ZipFiles(dest string, withSrcDir bool, srcFiles ...string) (err error) {
	return ZipFilesFilter(dest, withSrcDir, nil, srcFiles...)
}

// ZipFilesWithOptions creates a ZIP archive from source files or directories with per-call options.
func ZipFilesWithOptions(dest string, withSrcDir bool, srcFiles []string, opts ...ArchiveOption) (err error) {
	return ZipFilesFilterWithOptions(dest, withSrcDir, nil, srcFiles, opts...)
}

// ZipFilesFilter creates a ZIP archive and filters source paths.
func ZipFilesFilter(dest string, withSrcDir bool, filter FileFilter, srcFiles ...string) (err error) {
	return ZipFilesFilterWithOptions(dest, withSrcDir, filter, srcFiles)
}

// ZipFilesFilterWithOptions creates a ZIP archive with source filtering and per-call options.
func ZipFilesFilterWithOptions(dest string, withSrcDir bool, filter FileFilter, srcFiles []string, opts ...ArchiveOption) (err error) {
	cfg := applyArchiveOptions(opts)
	if err := validateZipTarget(dest, srcFiles...); err != nil {
		return err
	}
	if dir := filepath.Dir(dest); dir != "." {
		if err := os.MkdirAll(dir, cfg.dirPerm); err != nil {
			return err
		}
	}
	flag := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if !cfg.overwrite {
		flag |= os.O_EXCL
	}
	out, err := os.OpenFile(dest, flag, cfg.filePerm) // #nosec G304 -- destination path is an explicit caller-provided archive output.
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := out.Close(); err == nil {
			err = closeErr
		}
	}()
	return ZipToWriterWithOptions(out, withSrcDir, filter, srcFiles, opts...)
}

// ZipToWriter writes source files or directories into out as a ZIP archive.
func ZipToWriter(out io.Writer, withSrcDir bool, filter FileFilter, srcFiles ...string) (err error) {
	return ZipToWriterWithOptions(out, withSrcDir, filter, srcFiles)
}

// ZipToWriterWithOptions writes source files or directories into out as a ZIP archive with per-call options.
func ZipToWriterWithOptions(out io.Writer, withSrcDir bool, filter FileFilter, srcFiles []string, opts ...ArchiveOption) (err error) {
	cfg := applyArchiveOptions(opts)
	zw := archivezip.NewWriter(out)
	if cfg.compressionMethod == archivezip.Deflate && cfg.compressionLevel != flate.DefaultCompression {
		zw.RegisterCompressor(archivezip.Deflate, func(w io.Writer) (io.WriteCloser, error) {
			return flate.NewWriter(w, cfg.compressionLevel)
		})
	}
	defer func() {
		if closeErr := zw.Close(); err == nil {
			err = closeErr
		}
	}()
	for _, src := range srcFiles {
		if src == "" {
			continue
		}
		info, err := os.Lstat(src)
		if err != nil {
			return err
		}
		base := filepath.Dir(src)
		name := filepath.Base(src)
		if info.IsDir() && !withSrcDir {
			base = src
			name = ""
		}
		if err := addPath(zw, src, base, name, filter, cfg); err != nil {
			return err
		}
	}
	return nil
}

// ZipData creates or overwrites zipFile and adds one text entry.
func ZipData(zipFile, path, data string) error {
	return ZipBytes(zipFile, path, []byte(data))
}

// ZipBytes creates or overwrites zipFile and adds one byte entry.
func ZipBytes(zipFile, path string, data []byte) error {
	return ZipEntries(zipFile, EntryData{Name: path, Data: data})
}

// ZipEntries creates or overwrites zipFile and adds in-memory entries.
func ZipEntries(zipFile string, entries ...EntryData) (err error) {
	return ZipEntriesWithOptions(zipFile, entries)
}

// ZipEntriesWithOptions creates or overwrites zipFile and adds in-memory entries with per-call options.
func ZipEntriesWithOptions(zipFile string, entries []EntryData, opts ...ArchiveOption) (err error) {
	cfg := applyArchiveOptions(opts)
	if dir := filepath.Dir(zipFile); dir != "." {
		if err := os.MkdirAll(dir, cfg.dirPerm); err != nil {
			return err
		}
	}
	flag := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if !cfg.overwrite {
		flag |= os.O_EXCL
	}
	out, err := os.OpenFile(zipFile, flag, cfg.filePerm) // #nosec G304 -- destination path is an explicit caller-provided archive output.
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := out.Close(); err == nil {
			err = closeErr
		}
	}()
	return ZipEntriesToWriterWithOptions(out, entries, opts...)
}

// ZipEntriesToWriter writes in-memory entries into out as a ZIP archive.
func ZipEntriesToWriter(out io.Writer, entries ...EntryData) (err error) {
	return ZipEntriesToWriterWithOptions(out, entries)
}

// ZipEntriesToWriterWithOptions writes in-memory entries into out as a ZIP archive with per-call options.
func ZipEntriesToWriterWithOptions(out io.Writer, entries []EntryData, opts ...ArchiveOption) (err error) {
	streams := make([]StreamEntry, 0, len(entries))
	for _, entry := range entries {
		streams = append(streams, StreamEntry{Name: entry.Name, Reader: bytes.NewReader(entry.Data)})
	}
	return ZipStreamsToWriterWithOptions(out, streams, opts...)
}

// ZipStreams creates or overwrites zipFile and adds stream entries.
func ZipStreams(zipFile string, entries ...StreamEntry) (err error) {
	return ZipStreamsWithOptions(zipFile, entries)
}

// ZipStreamsWithOptions creates or overwrites zipFile and adds stream entries with per-call options.
func ZipStreamsWithOptions(zipFile string, entries []StreamEntry, opts ...ArchiveOption) (err error) {
	cfg := applyArchiveOptions(opts)
	if dir := filepath.Dir(zipFile); dir != "." {
		if err := os.MkdirAll(dir, cfg.dirPerm); err != nil {
			return err
		}
	}
	flag := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if !cfg.overwrite {
		flag |= os.O_EXCL
	}
	out, err := os.OpenFile(zipFile, flag, cfg.filePerm) // #nosec G304 -- destination path is an explicit caller-provided archive output.
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := out.Close(); err == nil {
			err = closeErr
		}
	}()
	return ZipStreamsToWriterWithOptions(out, entries, opts...)
}

// ZipStreamsToWriter writes stream entries into out as a ZIP archive.
func ZipStreamsToWriter(out io.Writer, entries ...StreamEntry) (err error) {
	return ZipStreamsToWriterWithOptions(out, entries)
}

// ZipStreamsToWriterWithOptions writes stream entries into out as a ZIP archive with per-call options.
func ZipStreamsToWriterWithOptions(out io.Writer, entries []StreamEntry, opts ...ArchiveOption) (err error) {
	cfg := applyArchiveOptions(opts)
	zw := archivezip.NewWriter(out)
	if cfg.compressionMethod == archivezip.Deflate && cfg.compressionLevel != flate.DefaultCompression {
		zw.RegisterCompressor(archivezip.Deflate, func(w io.Writer) (io.WriteCloser, error) {
			return flate.NewWriter(w, cfg.compressionLevel)
		})
	}
	defer func() {
		if closeErr := zw.Close(); err == nil {
			err = closeErr
		}
	}()
	for _, entry := range entries {
		name, err := cleanEntryName(entry.Name)
		if err != nil {
			return err
		}
		header := &archivezip.FileHeader{Name: name, Method: cfg.compressionMethod}
		header.SetMode(cfg.filePerm)
		w, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}
		if _, err := io.Copy(w, entry.Reader); err != nil {
			return err
		}
	}
	return nil
}

// Unzip extracts zipFile into a sibling directory named after the archive.
func Unzip(zipFile string) (string, error) {
	dest := strings.TrimSuffix(zipFile, filepath.Ext(zipFile))
	return dest, UnzipTo(zipFile, dest)
}

// UnzipTo extracts zipFile into destDir.
func UnzipTo(zipFile, destDir string) error { return UnzipToLimit(zipFile, destDir, -1) }

// UnzipToLimit extracts zipFile into destDir and optionally limits total uncompressed size.
func UnzipToLimit(zipFile, destDir string, limit int64) error {
	return UnzipToWithOptions(zipFile, destDir, WithMaxBytes(limit))
}

// UnzipToWithOptions extracts zipFile into destDir with per-call options.
func UnzipToWithOptions(zipFile, destDir string, opts ...ArchiveOption) error {
	r, err := archivezip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer func() { _ = r.Close() }()
	return UnzipReaderToWithOptions(&r.Reader, destDir, opts...)
}

// UnzipReaderTo extracts archive reader contents into destDir.
func UnzipReaderTo(r *archivezip.Reader, destDir string) error {
	return UnzipReaderToLimit(r, destDir, -1)
}

// UnzipReaderToLimit extracts archive reader contents into destDir and optionally limits total size.
func UnzipReaderToLimit(r *archivezip.Reader, destDir string, limit int64) error {
	return UnzipReaderToWithOptions(r, destDir, WithMaxBytes(limit))
}

// UnzipReaderToWithOptions extracts archive reader contents into destDir with per-call options.
func UnzipReaderToWithOptions(r *archivezip.Reader, destDir string, opts ...ArchiveOption) error {
	cfg := applyArchiveOptions(opts)
	if r == nil {
		return invalidInputf("zip reader is nil")
	}
	if err := os.MkdirAll(destDir, cfg.dirPerm); err != nil {
		return err
	}
	var total int64
	for _, f := range r.File {
		if cfg.maxBytes > 0 {
			if f.UncompressedSize64 > uint64(maxInt64) {
				return invalidInputf("uncompressed size exceeds int64 limit")
			}
			size := int64(f.UncompressedSize64) // #nosec G115 -- guarded by the maxInt64 check above.
			if total > maxInt64-size {
				return invalidInputf("uncompressed size exceeds int64 limit")
			}
			total += size
			if total > cfg.maxBytes {
				return invalidInputf("uncompressed size exceeds limit")
			}
		}
		if err := extractFile(f, destDir, cfg); err != nil {
			return err
		}
	}
	return nil
}

// Get returns a reader for the named entry in zipFile.
func Get(zipFile, name string) (io.ReadCloser, error) {
	r, err := archivezip.OpenReader(zipFile)
	if err != nil {
		return nil, err
	}
	for _, f := range r.File {
		if f.Name == name {
			rc, err := f.Open()
			if err != nil {
				_ = r.Close()
				return nil, err
			}
			return &readCloserWithClose{ReadCloser: rc, close: r.Close}, nil
		}
	}
	_ = r.Close()
	return nil, notFound("zip entry not found: "+name, os.ErrNotExist)
}

// GetBytes returns the content of the named entry in zipFile.
func GetBytes(zipFile, name string) ([]byte, error) {
	return GetBytesWithOptions(zipFile, name)
}

// GetBytesWithOptions returns the content of the named entry in zipFile with per-call options.
func GetBytesWithOptions(zipFile, name string, opts ...ArchiveOption) ([]byte, error) {
	rc, err := Get(zipFile, name)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rc.Close() }()
	return readAllLimit(rc, applyArchiveOptions(opts).maxBytes)
}

// Read walks every archive entry and calls consumer.
func Read(zipFile string, consumer func(*archivezip.File) error) error {
	r, err := archivezip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer func() { _ = r.Close() }()
	for _, f := range r.File {
		if err := consumer(f); err != nil {
			return err
		}
	}
	return nil
}

// ListFileNames returns direct file names under dir inside zipFile.
func ListFileNames(zipFile, dir string) ([]string, error) {
	r, err := archivezip.OpenReader(zipFile)
	if err != nil {
		return nil, err
	}
	defer func() { _ = r.Close() }()
	if strings.TrimSpace(dir) != "" && !strings.HasSuffix(dir, "/") {
		dir += "/"
	}
	names := make([]string, 0)
	for _, f := range r.File {
		name := f.Name
		if dir != "" {
			if !strings.HasPrefix(name, dir) {
				continue
			}
			name = strings.TrimPrefix(name, dir)
		}
		if name != "" && !strings.Contains(name, "/") && !f.FileInfo().IsDir() {
			names = append(names, name)
		}
	}
	return names, nil
}

// Gzip compresses data using gzip.
func Gzip(data []byte) ([]byte, error) { return GzipReader(bytes.NewReader(data), len(data)) }

// GzipString compresses text using gzip.
func GzipString(content string) ([]byte, error) { return Gzip([]byte(content)) }

// GzipFile compresses a file using gzip and returns compressed bytes.
func GzipFile(path string) ([]byte, error) {
	// #nosec G304 -- SDK file helper intentionally opens the caller-provided path.
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	return GzipReader(f, int(info.Size()))
}

// GzipReader compresses all bytes from r using gzip.
func GzipReader(r io.Reader, estimatedLength int) ([]byte, error) {
	if estimatedLength <= 0 {
		estimatedLength = defaultBufferSize
	}
	var buf bytes.Buffer
	buf.Grow(estimatedLength)
	w := gzip.NewWriter(&buf)
	if _, err := io.Copy(w, r); err != nil {
		_ = w.Close()
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// UnGzip decompresses gzip data.
func UnGzip(data []byte) ([]byte, error) { return UnGzipWithOptions(data) }

// UnGzipWithOptions decompresses gzip data with per-call options.
func UnGzipWithOptions(data []byte, opts ...ArchiveOption) ([]byte, error) {
	return UnGzipReaderWithOptions(bytes.NewReader(data), len(data), opts...)
}

// Gunzip decompresses gzip data.
func Gunzip(data []byte) ([]byte, error) { return UnGzip(data) }

// UnGzipString decompresses gzip data and returns text.
func UnGzipString(data []byte) (string, error) {
	out, err := UnGzip(data)
	return string(out), err
}

// UnGzipReader decompresses all gzip bytes from r.
func UnGzipReader(r io.Reader, estimatedLength int) ([]byte, error) {
	return UnGzipReaderWithOptions(r, estimatedLength)
}

// UnGzipReaderWithOptions decompresses gzip bytes from r with per-call options.
func UnGzipReaderWithOptions(r io.Reader, estimatedLength int, opts ...ArchiveOption) ([]byte, error) {
	if estimatedLength <= 0 {
		estimatedLength = defaultBufferSize
	}
	cfg := applyArchiveOptions(opts)
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer func() { _ = zr.Close() }()
	var buf bytes.Buffer
	buf.Grow(estimatedLength)
	err = copyLimit(&buf, zr, cfg.maxBytes) // #nosec G110 -- this low-level helper intentionally decompresses caller-provided gzip data.
	return buf.Bytes(), err
}

// Zlib compresses data using zlib with the default compression level.
func Zlib(data []byte) ([]byte, error) { return ZlibLevel(data, flate.DefaultCompression) }

// ZlibString compresses text using zlib with the specified compression level.
func ZlibString(content string, level int) ([]byte, error) { return ZlibLevel([]byte(content), level) }

// ZlibFile compresses a file using zlib with the specified compression level.
func ZlibFile(path string, level int) ([]byte, error) {
	// #nosec G304 -- SDK file helper intentionally opens the caller-provided path.
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	return ZlibReader(f, level, int(info.Size()))
}

// ZlibLevel compresses data using zlib with the specified compression level.
func ZlibLevel(data []byte, level int) ([]byte, error) {
	return ZlibReader(bytes.NewReader(data), level, len(data))
}

// ZlibReader compresses all bytes from r using zlib with the specified compression level.
func ZlibReader(r io.Reader, level, estimatedLength int) ([]byte, error) {
	if estimatedLength <= 0 {
		estimatedLength = defaultBufferSize
	}
	var buf bytes.Buffer
	buf.Grow(estimatedLength)
	w, err := zlib.NewWriterLevel(&buf, level)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(w, r); err != nil {
		_ = w.Close()
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// UnZlib decompresses zlib data.
func UnZlib(data []byte) ([]byte, error) { return UnZlibWithOptions(data) }

// UnZlibWithOptions decompresses zlib data with per-call options.
func UnZlibWithOptions(data []byte, opts ...ArchiveOption) ([]byte, error) {
	return UnZlibReaderWithOptions(bytes.NewReader(data), len(data), opts...)
}

// Unzlib decompresses zlib data.
func Unzlib(data []byte) ([]byte, error) { return UnZlib(data) }

// UnZlibString decompresses zlib data and returns text.
func UnZlibString(data []byte) (string, error) {
	out, err := UnZlib(data)
	return string(out), err
}

// UnZlibReader decompresses all zlib bytes from r.
func UnZlibReader(r io.Reader, estimatedLength int) ([]byte, error) {
	return UnZlibReaderWithOptions(r, estimatedLength)
}

// UnZlibReaderWithOptions decompresses zlib bytes from r with per-call options.
func UnZlibReaderWithOptions(r io.Reader, estimatedLength int, opts ...ArchiveOption) ([]byte, error) {
	if estimatedLength <= 0 {
		estimatedLength = defaultBufferSize
	}
	cfg := applyArchiveOptions(opts)
	zr, err := zlib.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer func() { _ = zr.Close() }()
	var buf bytes.Buffer
	buf.Grow(estimatedLength)
	err = copyLimit(&buf, zr, cfg.maxBytes) // #nosec G110 -- this low-level helper intentionally decompresses caller-provided zlib data.
	return buf.Bytes(), err
}

func appendWithFilter(zipPath, srcPath string, filter FileFilter) error {
	tmp, err := os.CreateTemp(filepath.Dir(zipPath), ".zip-append-*.zip")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	zw := archivezip.NewWriter(tmp)
	if _, err := os.Stat(zipPath); err == nil {
		r, err := archivezip.OpenReader(zipPath)
		if err != nil {
			_ = zw.Close()
			_ = tmp.Close()
			_ = os.Remove(tmpPath)
			return err
		}
		for _, f := range r.File {
			if err := copyExistingEntry(zw, f); err != nil {
				_ = r.Close()
				_ = zw.Close()
				_ = tmp.Close()
				_ = os.Remove(tmpPath)
				return err
			}
		}
		_ = r.Close()
	}
	info, err := os.Lstat(srcPath)
	if err != nil {
		_ = zw.Close()
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return err
	}
	base := filepath.Dir(srcPath)
	name := filepath.Base(srcPath)
	if info.IsDir() && filepath.Dir(srcPath) == srcPath {
		base = srcPath
		name = ""
	}
	if err := addPath(zw, srcPath, base, name, filter, defaultArchiveConfig()); err != nil {
		_ = zw.Close()
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return err
	}
	if err := zw.Close(); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return os.Rename(tmpPath, zipPath)
}

func addPath(zw *archivezip.Writer, path, base, name string, filter FileFilter, cfg archiveConfig) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if filter != nil && !filter(path, info) {
		return nil
	}
	zipName := name
	if zipName == "" {
		rel, err := filepath.Rel(base, path)
		if err != nil {
			return err
		}
		if rel == "." {
			zipName = ""
		} else {
			zipName = rel
		}
	}
	zipName = filepath.ToSlash(filepath.Clean(zipName))
	if zipName == "." {
		zipName = ""
	}
	if zipName != "" {
		if _, err := cleanEntryName(zipName); err != nil {
			return err
		}
	}
	if info.IsDir() {
		if zipName != "" {
			header, err := archivezip.FileInfoHeader(info)
			if err != nil {
				return err
			}
			header.Name = strings.TrimSuffix(zipName, "/") + "/"
			header.SetMode(info.Mode())
			if _, err := zw.CreateHeader(header); err != nil {
				return err
			}
		}
		entries, err := os.ReadDir(path)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			child := filepath.Join(path, entry.Name())
			childName := entry.Name()
			if zipName != "" {
				childName = filepath.Join(zipName, entry.Name())
			}
			if err := addPath(zw, child, base, childName, filter, cfg); err != nil {
				return err
			}
		}
		return nil
	}
	return addFile(zw, path, zipName, info, cfg)
}

func addFile(zw *archivezip.Writer, path, zipName string, info os.FileInfo, cfg archiveConfig) error {
	header, err := archivezip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = zipName
	header.Method = cfg.compressionMethod
	header.SetMode(info.Mode())
	w, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		linkTarget, err := os.Readlink(path)
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, linkTarget)
		return err
	}
	r, err := os.Open(path) // #nosec G304 -- archive creation intentionally reads caller-provided source paths.
	if err != nil {
		return err
	}
	defer func() { _ = r.Close() }()
	_, err = io.Copy(w, r) // #nosec G110 -- copying existing archive entries preserves caller-provided archive contents.
	return err
}

func extractFile(f *archivezip.File, destDir string, cfg archiveConfig) error {
	target, err := safeZipTarget(destDir, f.Name)
	if err != nil {
		return err
	}
	if f.FileInfo().IsDir() {
		perm := cfg.dirPerm
		if cfg.preserveMode {
			perm = f.Mode()
		}
		return os.MkdirAll(target, perm)
	}
	if err := os.MkdirAll(filepath.Dir(target), cfg.dirPerm); err != nil {
		return err
	}
	r, err := f.Open()
	if err != nil {
		return err
	}
	defer func() { _ = r.Close() }()
	perm := cfg.filePerm
	if cfg.preserveMode {
		perm = f.Mode()
	}
	flag := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if !cfg.overwrite {
		flag |= os.O_EXCL
	}
	w, err := os.OpenFile(target, flag, perm) // #nosec G304 -- target is validated by safeZipTarget before extraction.
	if err != nil {
		return err
	}
	if _, err := io.Copy(w, r); err != nil { // #nosec G110 -- unzip extraction is guarded by safeZipTarget and optional UnzipToLimit size checks.
		_ = w.Close()
		return err
	}
	return w.Close()
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
		return nil, invalidInputf("archive data exceeds max bytes: %d", maxBytes)
	}
	return b, nil
}

func copyLimit(dst io.Writer, src io.Reader, maxBytes int64) error {
	if maxBytes <= 0 {
		_, err := io.Copy(dst, src)
		return err
	}
	limited := &io.LimitedReader{R: src, N: maxBytes + 1}
	n, err := io.Copy(dst, limited)
	if err != nil {
		return err
	}
	if n > maxBytes {
		return invalidInputf("archive data exceeds max bytes: %d", maxBytes)
	}
	return nil
}

func copyExistingEntry(zw *archivezip.Writer, f *archivezip.File) error {
	r, err := f.Open()
	if err != nil {
		return err
	}
	defer func() { _ = r.Close() }()
	header := f.FileHeader
	w, err := zw.CreateHeader(&header)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, r) // #nosec G110 -- copying existing archive entries preserves caller-provided archive contents.
	return err
}

func cleanEntryName(name string) (string, error) {
	if name == "" || filepath.IsAbs(name) {
		return "", invalidInputf("invalid zip entry name %q", name)
	}
	cleaned := filepath.ToSlash(filepath.Clean(name))
	if cleaned == "." || strings.HasPrefix(cleaned, "../") || cleaned == ".." || strings.HasPrefix(cleaned, "/") {
		return "", invalidInputf("invalid zip entry name %q", name)
	}
	return cleaned, nil
}

func safeZipTarget(destDir, name string) (string, error) {
	cleaned, err := cleanEntryName(name)
	if err != nil {
		return "", err
	}
	target := filepath.Join(destDir, filepath.FromSlash(cleaned))
	destAbs, err := filepath.Abs(destDir)
	if err != nil {
		return "", err
	}
	targetAbs, err := filepath.Abs(target)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(destAbs, targetAbs)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", invalidInputf("invalid zip entry name %q", name)
	}
	return target, nil
}

func validateZipTarget(zipFile string, srcFiles ...string) error {
	info, err := os.Stat(zipFile)
	if err == nil && info.IsDir() {
		return invalidInputf("zip file %q must not be a directory", zipFile)
	}
	zipAbs, err := filepath.Abs(zipFile)
	if err != nil {
		return err
	}
	zipDir := filepath.Dir(zipAbs)
	for _, src := range srcFiles {
		if src == "" {
			continue
		}
		info, err := os.Stat(src)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			continue
		}
		srcAbs, err := filepath.Abs(src)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcAbs, zipDir)
		if err != nil {
			return err
		}
		if rel == "." || (!strings.HasPrefix(rel, "..") && rel != "") {
			return invalidInputf("zip file path %q must not be inside source directory %q", zipFile, src)
		}
	}
	return nil
}

type readCloserWithClose struct {
	io.ReadCloser
	close func() error
}

func (r *readCloserWithClose) Close() error {
	err1 := r.ReadCloser.Close()
	err2 := r.close()
	if err1 != nil {
		return err1
	}
	return err2
}
