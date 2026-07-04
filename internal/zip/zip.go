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
	defaultBufferSize    = 32
	DefaultUnzipMaxBytes = 1 << 30
	maxInt64             = int64(1<<63 - 1)
	maxZipEntries        = 100000
	maxZipEntryDepth     = 1024
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

type (
	OpenFunc          func(string) (io.ReadCloser, error)
	ReadFileFunc      func(string) ([]byte, error)
	OpenFileFunc      func(string, int, os.FileMode) (io.WriteCloser, error)
	EvalSymlinksFunc  func(string) (string, error)
	StatFunc          func(string) (os.FileInfo, error)
	LstatFunc         func(string) (os.FileInfo, error)
	ReadDirFunc       func(string) ([]os.DirEntry, error)
	ReadlinkFunc      func(string) (string, error)
	MkdirAllFunc      func(string, os.FileMode) error
	RemoveFunc        func(string) error
	RenameFunc        func(string, string) error
	OpenZipReaderFunc func(string) (*archivezip.ReadCloser, error)
	CreateTempFunc    func(string, string) (TempFile, error)
)

// TempFile is the writable temporary file contract used by append operations.
type TempFile interface {
	io.WriteCloser
	Name() string
}

type archiveConfig struct {
	dirPerm           os.FileMode
	filePerm          os.FileMode
	overwrite         bool
	preserveMode      bool
	withSrcDir        bool
	setWithSrcDir     bool
	filter            FileFilter
	setFilter         bool
	compressionMethod uint16
	compressionLevel  int
	maxBytes          int64
	setMaxBytes       bool
	open              OpenFunc
	readFile          ReadFileFunc
	openFile          OpenFileFunc
	evalSymlinks      EvalSymlinksFunc
	stat              StatFunc
	lstat             LstatFunc
	readDir           ReadDirFunc
	readlink          ReadlinkFunc
	mkdirAll          MkdirAllFunc
	remove            RemoveFunc
	rename            RenameFunc
	openZipReader     OpenZipReaderFunc
	createTemp        CreateTempFunc
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
		open:              defaultOpen,
		readFile:          defaultReadFile,
		openFile:          defaultOpenFile,
		evalSymlinks:      filepath.EvalSymlinks,
		stat:              os.Stat,
		lstat:             os.Lstat,
		readDir:           os.ReadDir,
		readlink:          os.Readlink,
		mkdirAll:          os.MkdirAll,
		remove:            os.Remove,
		rename:            os.Rename,
		openZipReader:     archivezip.OpenReader,
		createTemp:        defaultCreateTemp,
	}
}

// WithEvalSymlinks sets the function used to resolve extraction paths for symlink escape checks.
func WithEvalSymlinks(evalSymlinks EvalSymlinksFunc) ArchiveOption {
	return func(c *archiveConfig) {
		if evalSymlinks != nil {
			c.evalSymlinks = evalSymlinks
		}
	}
}

func defaultOpen(path string) (io.ReadCloser, error) {
	// #nosec G304 -- archive helpers intentionally read caller-provided paths.
	return os.Open(path)
}

func defaultReadFile(path string) ([]byte, error) {
	// #nosec G304 -- archive helpers intentionally read caller-provided paths.
	return os.ReadFile(path)
}

func defaultOpenFile(path string, flag int, perm os.FileMode) (io.WriteCloser, error) {
	// #nosec G304 -- archive helpers intentionally write caller-provided paths.
	return os.OpenFile(path, flag, perm)
}

func defaultCreateTemp(dir, pattern string) (TempFile, error) {
	return os.CreateTemp(dir, pattern)
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

// WithSourceDir controls whether source directory names are included in newly created ZIP archives.
func WithSourceDir(withSrcDir bool) ArchiveOption {
	return func(c *archiveConfig) {
		c.withSrcDir = withSrcDir
		c.setWithSrcDir = true
	}
}

// WithFileFilter sets the source path filter used by newly created ZIP archives.
func WithFileFilter(filter FileFilter) ArchiveOption {
	return func(c *archiveConfig) {
		if filter != nil {
			c.filter = filter
			c.setFilter = true
		}
	}
}

// WithCompressionMethod sets the ZIP compression method used for newly created entries.
func WithCompressionMethod(method uint16) ArchiveOption {
	return func(c *archiveConfig) { c.compressionMethod = method }
}

// WithCompressionLevel sets the deflate compression level used for newly created entries.
func WithCompressionLevel(level int) ArchiveOption {
	return func(c *archiveConfig) { c.compressionLevel = level }
}

// WithMaxBytes limits bytes read from archive entries, decompressed streams, or compression inputs. Non-positive means unlimited.
func WithMaxBytes(n int64) ArchiveOption {
	return func(c *archiveConfig) {
		c.maxBytes = n
		c.setMaxBytes = true
	}
}

// WithOpen sets the function used to open source files for reading.
func WithOpen(open OpenFunc) ArchiveOption {
	return func(c *archiveConfig) {
		if open != nil {
			c.open = open
		}
	}
}

// WithReadFile sets the function used to read a complete source file.
func WithReadFile(readFile ReadFileFunc) ArchiveOption {
	return func(c *archiveConfig) {
		if readFile != nil {
			c.readFile = readFile
		}
	}
}

// WithOpenFile sets the function used to open archive/extracted files for writing.
func WithOpenFile(openFile OpenFileFunc) ArchiveOption {
	return func(c *archiveConfig) {
		if openFile != nil {
			c.openFile = openFile
		}
	}
}

// WithStat sets the function used to inspect existing archive paths.
func WithStat(stat StatFunc) ArchiveOption {
	return func(c *archiveConfig) {
		if stat != nil {
			c.stat = stat
		}
	}
}

// WithLstat sets the function used to inspect source paths without following symlinks.
func WithLstat(lstat LstatFunc) ArchiveOption {
	return func(c *archiveConfig) {
		if lstat != nil {
			c.lstat = lstat
		}
	}
}

// WithReadDir sets the function used to enumerate source directories.
func WithReadDir(readDir ReadDirFunc) ArchiveOption {
	return func(c *archiveConfig) {
		if readDir != nil {
			c.readDir = readDir
		}
	}
}

// WithReadlink sets the function used to read symlink targets.
func WithReadlink(readlink ReadlinkFunc) ArchiveOption {
	return func(c *archiveConfig) {
		if readlink != nil {
			c.readlink = readlink
		}
	}
}

// WithMkdirAll sets the function used to create directory trees.
func WithMkdirAll(mkdirAll MkdirAllFunc) ArchiveOption {
	return func(c *archiveConfig) {
		if mkdirAll != nil {
			c.mkdirAll = mkdirAll
		}
	}
}

// WithRemove sets the function used to remove temporary files.
func WithRemove(remove RemoveFunc) ArchiveOption {
	return func(c *archiveConfig) {
		if remove != nil {
			c.remove = remove
		}
	}
}

// WithRename sets the function used to move completed temporary archives into place.
func WithRename(rename RenameFunc) ArchiveOption {
	return func(c *archiveConfig) {
		if rename != nil {
			c.rename = rename
		}
	}
}

// WithOpenZipReader sets the function used to open existing ZIP archives for reading.
func WithOpenZipReader(openZipReader OpenZipReaderFunc) ArchiveOption {
	return func(c *archiveConfig) {
		if openZipReader != nil {
			c.openZipReader = openZipReader
		}
	}
}

// WithCreateTemp sets the function used to create temporary archives for append operations.
func WithCreateTemp(createTemp CreateTempFunc) ArchiveOption {
	return func(c *archiveConfig) {
		if createTemp != nil {
			c.createTemp = createTemp
		}
	}
}

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
	if cfg.open == nil {
		cfg.open = defaultOpen
	}
	if cfg.readFile == nil {
		cfg.readFile = defaultReadFile
	}
	if cfg.openFile == nil {
		cfg.openFile = defaultOpenFile
	}
	if cfg.evalSymlinks == nil {
		cfg.evalSymlinks = filepath.EvalSymlinks
	}
	if cfg.stat == nil {
		cfg.stat = os.Stat
	}
	if cfg.lstat == nil {
		cfg.lstat = os.Lstat
	}
	if cfg.readDir == nil {
		cfg.readDir = os.ReadDir
	}
	if cfg.readlink == nil {
		cfg.readlink = os.Readlink
	}
	if cfg.mkdirAll == nil {
		cfg.mkdirAll = os.MkdirAll
	}
	if cfg.remove == nil {
		cfg.remove = os.Remove
	}
	if cfg.rename == nil {
		cfg.rename = os.Rename
	}
	if cfg.openZipReader == nil {
		cfg.openZipReader = archivezip.OpenReader
	}
	if cfg.createTemp == nil {
		cfg.createTemp = defaultCreateTemp
	}
	return cfg
}

func applyUnzipOptions(opts []ArchiveOption) archiveConfig {
	cfg := applyArchiveOptions(opts)
	if !cfg.setMaxBytes {
		cfg.maxBytes = DefaultUnzipMaxBytes
	}
	return cfg
}

func applyDecompressOptions(opts []ArchiveOption) archiveConfig {
	return applyUnzipOptions(opts)
}

// Open opens a ZIP file for reading.
func Open(path string) (*archivezip.ReadCloser, error) { return OpenWithOptions(path) }

// OpenWithOptions opens a ZIP file for reading with per-call options.
func OpenWithOptions(path string, opts ...ArchiveOption) (*archivezip.ReadCloser, error) {
	return applyArchiveOptions(opts).openZipReader(path)
}

// ReadFile reads a file from disk. It is useful when composing in-memory archive entries.
func ReadFile(path string) ([]byte, error) { return ReadFileWithOptions(path) }

// ReadFileWithOptions reads a file using per-call archive options.
func ReadFileWithOptions(path string, opts ...ArchiveOption) ([]byte, error) {
	return applyArchiveOptions(opts).readFile(path)
}

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
	return AppendWithOptions(zipPath, srcPath)
}

// AppendWithOptions appends srcPath into zipPath by rewriting the archive with per-call options.
func AppendWithOptions(zipPath, srcPath string, opts ...ArchiveOption) error {
	return appendWithFilter(zipPath, srcPath, opts...)
}

// Zip creates an archive next to srcPath and returns the archive path.
func Zip(srcPath string) (string, error) {
	dest := strings.TrimSuffix(srcPath, filepath.Ext(srcPath)) + ".zip"
	return dest, ZipTo(srcPath, dest, false)
}

// ZipTo creates an archive at zipPath from srcPath.
func ZipTo(srcPath, zipPath string, withSrcDir bool) error {
	return ZipFilesWithOptions(zipPath, []string{srcPath}, WithSourceDir(withSrcDir))
}

// ZipFiles creates a ZIP archive from source files or directories.
func ZipFiles(dest string, withSrcDir bool, srcFiles ...string) (err error) {
	return ZipFilesWithOptions(dest, srcFiles, WithSourceDir(withSrcDir))
}

// ZipFilesWithOptions creates a ZIP archive from source files or directories with per-call options.
func ZipFilesWithOptions(dest string, srcFiles []string, opts ...ArchiveOption) (err error) {
	return ZipFilesFilterWithOptions(dest, false, nil, srcFiles, opts...)
}

// ZipFilesFilter creates a ZIP archive and filters source paths.
func ZipFilesFilter(dest string, withSrcDir bool, filter FileFilter, srcFiles ...string) (err error) {
	return ZipFilesFilterWithOptions(dest, withSrcDir, filter, srcFiles)
}

// ZipFilesFilterWithOptions creates a ZIP archive with source filtering and per-call options.
func ZipFilesFilterWithOptions(dest string, withSrcDir bool, filter FileFilter, srcFiles []string, opts ...ArchiveOption) (err error) {
	cfg := applyArchiveOptions(opts)
	if cfg.setWithSrcDir {
		withSrcDir = cfg.withSrcDir
	}
	if cfg.setFilter {
		filter = cfg.filter
	}
	if err := validateZipTarget(cfg, dest, srcFiles...); err != nil {
		return err
	}
	if dir := filepath.Dir(dest); dir != "." {
		if err := cfg.mkdirAll(dir, cfg.dirPerm); err != nil {
			return err
		}
	}
	flag := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if !cfg.overwrite {
		flag |= os.O_EXCL
	}
	out, err := cfg.openFile(dest, flag, cfg.filePerm)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := out.Close(); err == nil {
			err = closeErr
		}
	}()
	writerOpts := append([]ArchiveOption{WithSourceDir(withSrcDir), WithFileFilter(filter)}, opts...)
	return ZipToWriterWithOptions(out, srcFiles, writerOpts...)
}

// ZipToWriter writes source files or directories into out as a ZIP archive.
func ZipToWriter(out io.Writer, withSrcDir bool, filter FileFilter, srcFiles ...string) (err error) {
	return ZipToWriterWithOptions(out, srcFiles, WithSourceDir(withSrcDir), WithFileFilter(filter))
}

// ZipToWriterWithOptions writes source files or directories into out as a ZIP archive with per-call options.
func ZipToWriterWithOptions(out io.Writer, srcFiles []string, opts ...ArchiveOption) (err error) {
	if out == nil {
		return invalidInputf("zip writer output is nil")
	}
	cfg := applyArchiveOptions(opts)
	withSrcDir := cfg.withSrcDir
	filter := cfg.filter
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
		info, err := cfg.lstat(src)
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
		if err := cfg.mkdirAll(dir, cfg.dirPerm); err != nil {
			return err
		}
	}
	flag := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if !cfg.overwrite {
		flag |= os.O_EXCL
	}
	out, err := cfg.openFile(zipFile, flag, cfg.filePerm)
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
		if err := cfg.mkdirAll(dir, cfg.dirPerm); err != nil {
			return err
		}
	}
	flag := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if !cfg.overwrite {
		flag |= os.O_EXCL
	}
	out, err := cfg.openFile(zipFile, flag, cfg.filePerm)
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
	if out == nil {
		return invalidInputf("zip writer output is nil")
	}
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
		if entry.Reader == nil {
			return invalidInputf("zip stream reader is nil")
		}
		header := &archivezip.FileHeader{Name: name, Method: cfg.compressionMethod}
		header.SetMode(cfg.filePerm)
		w, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}
		if _, err := copyLimit(w, entry.Reader, cfg.maxBytes); err != nil {
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

// UnzipTo extracts zipFile into destDir with the default uncompressed-size limit.
func UnzipTo(zipFile, destDir string) error {
	return UnzipToWithOptions(zipFile, destDir)
}

// UnzipToLimit extracts zipFile into destDir and optionally limits total uncompressed size.
func UnzipToLimit(zipFile, destDir string, limit int64) error {
	return UnzipToWithOptions(zipFile, destDir, WithMaxBytes(limit))
}

// UnzipToWithOptions extracts zipFile into destDir with per-call options.
func UnzipToWithOptions(zipFile, destDir string, opts ...ArchiveOption) error {
	cfg := applyUnzipOptions(opts)
	r, err := cfg.openZipReader(zipFile)
	if err != nil {
		return err
	}
	defer func() { _ = r.Close() }()
	return UnzipReaderToWithOptions(&r.Reader, destDir, opts...)
}

// UnzipReaderTo extracts archive reader contents into destDir.
func UnzipReaderTo(r *archivezip.Reader, destDir string) error {
	return UnzipReaderToWithOptions(r, destDir)
}

// UnzipReaderToLimit extracts archive reader contents into destDir and optionally limits total size.
func UnzipReaderToLimit(r *archivezip.Reader, destDir string, limit int64) error {
	return UnzipReaderToWithOptions(r, destDir, WithMaxBytes(limit))
}

// UnzipReaderToWithOptions extracts archive reader contents into destDir with per-call options.
func UnzipReaderToWithOptions(r *archivezip.Reader, destDir string, opts ...ArchiveOption) error {
	cfg := applyUnzipOptions(opts)
	if r == nil {
		return invalidInputf("zip reader is nil")
	}
	if err := validateArchiveEntries(r.File); err != nil {
		return err
	}
	if err := cfg.mkdirAll(destDir, cfg.dirPerm); err != nil {
		return err
	}
	var total int64
	for _, f := range r.File {
		if cfg.maxBytes > 0 {
			if f.UncompressedSize64 > uint64(maxInt64) {
				return invalidInputf("uncompressed size exceeds int64 limit")
			}
			size := int64(f.UncompressedSize64) // #nosec G115 -- guarded by the maxInt64 check above.
			if size > cfg.maxBytes-total {
				return invalidInputf("uncompressed size exceeds limit")
			}
		}
		written, err := extractFile(f, destDir, cfg, cfg.maxBytes-total)
		if err != nil {
			return err
		}
		if cfg.maxBytes > 0 {
			if total > maxInt64-written {
				return invalidInputf("uncompressed size exceeds int64 limit")
			}
			total += written
			if total > cfg.maxBytes {
				return invalidInputf("uncompressed size exceeds limit")
			}
		}
	}
	return nil
}

// Get returns a reader for the named entry in zipFile.
func Get(zipFile, name string) (io.ReadCloser, error) {
	return GetWithOptions(zipFile, name)
}

// GetWithOptions returns a reader for the named entry in zipFile with per-call options.
func GetWithOptions(zipFile, name string, opts ...ArchiveOption) (io.ReadCloser, error) {
	r, err := applyArchiveOptions(opts).openZipReader(zipFile)
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
	rc, err := GetWithOptions(zipFile, name, opts...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rc.Close() }()
	return readAllLimit(rc, applyDecompressOptions(opts).maxBytes)
}

// Read walks every archive entry and calls consumer.
func Read(zipFile string, consumer func(*archivezip.File) error) error {
	return ReadWithOptions(zipFile, consumer)
}

// ReadWithOptions walks every archive entry and calls consumer using per-call options.
func ReadWithOptions(zipFile string, consumer func(*archivezip.File) error, opts ...ArchiveOption) error {
	if consumer == nil {
		return invalidInputf("zip entry consumer is nil")
	}
	r, err := applyArchiveOptions(opts).openZipReader(zipFile)
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
	return ListFileNamesWithOptions(zipFile, dir)
}

// ListFileNamesWithOptions returns direct file names under dir inside zipFile using per-call options.
func ListFileNamesWithOptions(zipFile, dir string, opts ...ArchiveOption) ([]string, error) {
	r, err := applyArchiveOptions(opts).openZipReader(zipFile)
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
func Gzip(data []byte) ([]byte, error) { return GzipWithOptions(data) }

// GzipWithOptions compresses data using gzip with per-call options.
func GzipWithOptions(data []byte, opts ...ArchiveOption) ([]byte, error) {
	return GzipReaderWithOptions(bytes.NewReader(data), len(data), opts...)
}

// GzipString compresses text using gzip.
func GzipString(content string) ([]byte, error) { return Gzip([]byte(content)) }

// GzipFile compresses a file using gzip and returns compressed bytes.
func GzipFile(path string) ([]byte, error) {
	return GzipFileWithOptions(path)
}

// GzipFileWithOptions compresses a file using gzip and per-call options.
func GzipFileWithOptions(path string, opts ...ArchiveOption) ([]byte, error) {
	cfg := applyArchiveOptions(opts)
	f, err := cfg.open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	info, err := cfg.stat(path)
	if err != nil {
		return nil, err
	}
	return GzipReaderWithOptions(f, int(info.Size()), opts...)
}

// GzipReader compresses all bytes from r using gzip.
func GzipReader(r io.Reader, estimatedLength int) ([]byte, error) {
	return GzipReaderWithOptions(r, estimatedLength)
}

// GzipReaderWithOptions compresses all bytes from r using gzip with per-call options.
func GzipReaderWithOptions(r io.Reader, estimatedLength int, opts ...ArchiveOption) ([]byte, error) {
	if estimatedLength <= 0 {
		estimatedLength = defaultBufferSize
	}
	cfg := applyArchiveOptions(opts)
	var buf bytes.Buffer
	buf.Grow(estimatedLength)
	w, err := gzip.NewWriterLevel(&buf, cfg.compressionLevel)
	if err != nil {
		return nil, err
	}
	if _, err := copyLimit(w, r, cfg.maxBytes); err != nil {
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
	cfg := applyDecompressOptions(opts)
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer func() { _ = zr.Close() }()
	var buf bytes.Buffer
	buf.Grow(estimatedLength)
	_, err = copyLimit(&buf, zr, cfg.maxBytes) // #nosec G110 -- this low-level helper intentionally decompresses caller-provided gzip data.
	return buf.Bytes(), err
}

// Zlib compresses data using zlib with the default compression level.
func Zlib(data []byte) ([]byte, error) { return ZlibLevel(data, flate.DefaultCompression) }

// ZlibString compresses text using zlib with the specified compression level.
func ZlibString(content string, level int) ([]byte, error) { return ZlibLevel([]byte(content), level) }

// ZlibFile compresses a file using zlib with the specified compression level.
func ZlibFile(path string, level int) ([]byte, error) {
	return ZlibFileWithOptions(path, level)
}

// ZlibFileWithOptions compresses a file using zlib with the specified compression level and per-call options.
func ZlibFileWithOptions(path string, level int, opts ...ArchiveOption) ([]byte, error) {
	cfg := applyArchiveOptions(opts)
	f, err := cfg.open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	info, err := cfg.stat(path)
	if err != nil {
		return nil, err
	}
	return ZlibReaderWithOptions(f, level, int(info.Size()), opts...)
}

// ZlibLevel compresses data using zlib with the specified compression level.
func ZlibLevel(data []byte, level int) ([]byte, error) {
	return ZlibReader(bytes.NewReader(data), level, len(data))
}

// ZlibLevelWithOptions compresses data using zlib with the specified compression level and per-call options.
func ZlibLevelWithOptions(data []byte, level int, opts ...ArchiveOption) ([]byte, error) {
	return ZlibReaderWithOptions(bytes.NewReader(data), level, len(data), opts...)
}

// ZlibReader compresses all bytes from r using zlib with the specified compression level.
func ZlibReader(r io.Reader, level, estimatedLength int) ([]byte, error) {
	return ZlibReaderWithOptions(r, level, estimatedLength)
}

// ZlibReaderWithOptions compresses all bytes from r using zlib with the specified compression level and per-call options.
func ZlibReaderWithOptions(r io.Reader, level, estimatedLength int, opts ...ArchiveOption) ([]byte, error) {
	if estimatedLength <= 0 {
		estimatedLength = defaultBufferSize
	}
	cfg := applyArchiveOptions(opts)
	var buf bytes.Buffer
	buf.Grow(estimatedLength)
	w, err := zlib.NewWriterLevel(&buf, level)
	if err != nil {
		return nil, err
	}
	if _, err := copyLimit(w, r, cfg.maxBytes); err != nil {
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
	cfg := applyDecompressOptions(opts)
	zr, err := zlib.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer func() { _ = zr.Close() }()
	var buf bytes.Buffer
	buf.Grow(estimatedLength)
	_, err = copyLimit(&buf, zr, cfg.maxBytes) // #nosec G110 -- this low-level helper intentionally decompresses caller-provided zlib data.
	return buf.Bytes(), err
}

func appendWithFilter(zipPath, srcPath string, opts ...ArchiveOption) error {
	cfg := applyArchiveOptions(opts)
	filter := cfg.filter
	tmp, err := cfg.createTemp(filepath.Dir(zipPath), ".zip-append-*.zip")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	zw := archivezip.NewWriter(tmp)
	if cfg.compressionMethod == archivezip.Deflate && cfg.compressionLevel != flate.DefaultCompression {
		zw.RegisterCompressor(archivezip.Deflate, func(w io.Writer) (io.WriteCloser, error) {
			return flate.NewWriter(w, cfg.compressionLevel)
		})
	}
	if _, err := cfg.stat(zipPath); err == nil {
		r, err := cfg.openZipReader(zipPath)
		if err != nil {
			_ = zw.Close()
			_ = tmp.Close()
			_ = cfg.remove(tmpPath)
			return err
		}
		for _, f := range r.File {
			if err := copyExistingEntry(zw, f); err != nil {
				_ = r.Close()
				_ = zw.Close()
				_ = tmp.Close()
				_ = cfg.remove(tmpPath)
				return err
			}
		}
		_ = r.Close()
	}
	info, err := cfg.lstat(srcPath)
	if err != nil {
		_ = zw.Close()
		_ = tmp.Close()
		_ = cfg.remove(tmpPath)
		return err
	}
	base := filepath.Dir(srcPath)
	name := filepath.Base(srcPath)
	if info.IsDir() && filepath.Dir(srcPath) == srcPath {
		base = srcPath
		name = ""
	}
	if err := addPath(zw, srcPath, base, name, filter, cfg); err != nil {
		_ = zw.Close()
		_ = tmp.Close()
		_ = cfg.remove(tmpPath)
		return err
	}
	if err := zw.Close(); err != nil {
		_ = tmp.Close()
		_ = cfg.remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = cfg.remove(tmpPath)
		return err
	}
	return cfg.rename(tmpPath, zipPath)
}

func addPath(zw *archivezip.Writer, path, base, name string, filter FileFilter, cfg archiveConfig) error {
	info, err := cfg.lstat(path)
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
		entries, err := cfg.readDir(path)
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
		linkTarget, err := cfg.readlink(path)
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, linkTarget)
		return err
	}
	r, err := cfg.open(path)
	if err != nil {
		return err
	}
	defer func() { _ = r.Close() }()
	_, err = io.Copy(w, r) // #nosec G110 -- copying existing archive entries preserves caller-provided archive contents.
	return err
}

func extractFile(f *archivezip.File, destDir string, cfg archiveConfig, remainingBytes int64) (int64, error) {
	target, err := safeZipTarget(destDir, f.Name)
	if err != nil {
		return 0, err
	}
	if f.FileInfo().IsDir() {
		perm := cfg.dirPerm.Perm()
		if cfg.preserveMode {
			perm = f.Mode().Perm()
		}
		if err := validateNoSymlinkAncestorEscape(cfg, destDir, target); err != nil {
			return 0, err
		}
		if err := cfg.mkdirAll(target, perm); err != nil {
			return 0, err
		}
		if err := validateNoSymlinkEscape(cfg, destDir, target); err != nil {
			return 0, err
		}
		return 0, nil
	}
	parent := filepath.Dir(target)
	if err := validateNoSymlinkAncestorEscape(cfg, destDir, parent); err != nil {
		return 0, err
	}
	if err := cfg.mkdirAll(parent, cfg.dirPerm.Perm()); err != nil {
		return 0, err
	}
	if err := validateNoSymlinkEscape(cfg, destDir, parent); err != nil {
		return 0, err
	}
	r, err := f.Open()
	if err != nil {
		return 0, err
	}
	defer func() { _ = r.Close() }()
	perm := cfg.filePerm.Perm()
	if cfg.preserveMode {
		perm = f.Mode().Perm()
	}
	flag := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if !cfg.overwrite {
		flag |= os.O_EXCL
	}
	w, err := cfg.openFile(target, flag, perm)
	if err != nil {
		return 0, err
	}
	written, err := copyLimit(w, r, remainingBytes)
	if err != nil {
		_ = w.Close()
		_ = cfg.remove(target)
		return written, err
	}
	if err := w.Close(); err != nil {
		_ = cfg.remove(target)
		return written, err
	}
	return written, nil
}

func validateNoSymlinkEscape(cfg archiveConfig, destDir, targetParent string) error {
	destReal, err := cfg.evalSymlinks(destDir)
	if err != nil {
		return err
	}
	parentReal, err := cfg.evalSymlinks(targetParent)
	if err != nil {
		return err
	}
	destAbs, err := filepath.Abs(destReal)
	if err != nil {
		return err
	}
	parentAbs, err := filepath.Abs(parentReal)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(destAbs, parentAbs)
	if err != nil {
		return err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return invalidInputf("zip entry target escapes destination through symlink")
	}
	return nil
}

func validateNoSymlinkAncestorEscape(cfg archiveConfig, destDir, targetPath string) error {
	destAbs, err := filepath.Abs(destDir)
	if err != nil {
		return err
	}
	targetAbs, err := filepath.Abs(targetPath)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(destAbs, targetAbs)
	if err != nil {
		return err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return invalidInputf("zip entry target escapes destination")
	}
	current := destAbs
	if err := validateNoSymlinkEscape(cfg, destAbs, current); err != nil {
		return err
	}
	if rel == "." {
		return nil
	}
	for _, part := range strings.Split(rel, string(os.PathSeparator)) {
		if part == "" || part == "." {
			continue
		}
		current = filepath.Join(current, part)
		if _, err := cfg.lstat(current); err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if err := validateNoSymlinkEscape(cfg, destAbs, current); err != nil {
			return err
		}
	}
	return nil
}

func validateArchiveEntries(files []*archivezip.File) error {
	if len(files) > maxZipEntries {
		return invalidInputf("zip archive has too many entries: %d", len(files))
	}
	seen := make(map[string]bool, len(files))
	seenDir := make(map[string]bool, len(files))
	requiredDirs := map[string]struct{}{}
	for _, file := range files {
		if file == nil {
			return invalidInputf("zip entry is nil")
		}
		name, err := cleanEntryName(file.Name)
		if err != nil {
			return err
		}
		if entryDepth(name) > maxZipEntryDepth {
			return invalidInputf("zip entry %q is too deep", file.Name)
		}
		modeType := file.Mode().Type()
		if modeType != 0 && modeType != os.ModeDir && modeType != os.ModeSymlink {
			return invalidInputf("zip entry %q has unsupported file type", file.Name)
		}
		isDir := file.FileInfo().IsDir()
		if seen[name] {
			return invalidInputf("duplicate zip entry %q", file.Name)
		}
		if _, ok := requiredDirs[name]; ok && !isDir {
			return invalidInputf("zip entry %q conflicts with nested path", file.Name)
		}
		for _, parent := range parentEntryNames(name) {
			if seen[parent] && !seenDir[parent] {
				return invalidInputf("zip entry %q conflicts with file parent %q", file.Name, parent)
			}
			requiredDirs[parent] = struct{}{}
		}
		seen[name] = true
		seenDir[name] = isDir
	}
	return nil
}

func entryDepth(name string) int {
	if name == "" {
		return 0
	}
	return strings.Count(name, "/") + 1
}

func parentEntryNames(name string) []string {
	parts := strings.Split(name, "/")
	if len(parts) <= 1 {
		return nil
	}
	parents := make([]string, 0, len(parts)-1)
	for i := 1; i < len(parts); i++ {
		parents = append(parents, strings.Join(parts[:i], "/"))
	}
	return parents
}

func readAllLimit(r io.Reader, maxBytes int64) ([]byte, error) {
	if r == nil {
		return nil, invalidInputf("reader is nil")
	}
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

func copyLimit(dst io.Writer, src io.Reader, maxBytes int64) (int64, error) {
	if dst == nil {
		return 0, invalidInputf("writer is nil")
	}
	if src == nil {
		return 0, invalidInputf("reader is nil")
	}
	if maxBytes <= 0 {
		return io.Copy(dst, src)
	}
	limited := &io.LimitedReader{R: src, N: maxBytes}
	n, err := io.Copy(dst, limited)
	if err != nil {
		return n, err
	}
	var extra [1]byte
	extraN, extraErr := src.Read(extra[:])
	if extraN > 0 {
		return n, invalidInputf("archive data exceeds max bytes: %d", maxBytes)
	}
	if extraErr != nil && extraErr != io.EOF {
		return n, extraErr
	}
	return n, nil
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
	if strings.Contains(name, `\`) || strings.HasPrefix(name, `\\`) || len(name) >= 2 && name[1] == ':' {
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

func validateZipTarget(cfg archiveConfig, zipFile string, srcFiles ...string) error {
	info, err := cfg.stat(zipFile)
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
		info, err := cfg.stat(src)
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
