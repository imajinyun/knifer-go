package vzip

import (
	archivezip "archive/zip"
	"io"
	"os"

	zipimpl "github.com/imajinyun/knifer-go/internal/zip"
)

// FileFilter decides whether a source path should be added to an archive.
type FileFilter = zipimpl.FileFilter

// Entry describes an archive entry.
type Entry = zipimpl.Entry

// Writer is a ZIP archive writer.
type Writer = zipimpl.Writer

// Reader is a ZIP archive reader.
type Reader = zipimpl.Reader

// Error is the ZIP module error type.
type Error = zipimpl.ZipError

// EntryData represents in-memory content to add into a ZIP archive.
type EntryData = zipimpl.EntryData

// StreamEntry represents stream content to add into a ZIP archive.
type StreamEntry = zipimpl.StreamEntry

// ArchiveOption customizes archive helpers per call.
type ArchiveOption = zipimpl.ArchiveOption

type (
	OpenFunc          = zipimpl.OpenFunc
	ReadFileFunc      = zipimpl.ReadFileFunc
	OpenFileFunc      = zipimpl.OpenFileFunc
	EvalSymlinksFunc  = zipimpl.EvalSymlinksFunc
	StatFunc          = zipimpl.StatFunc
	LstatFunc         = zipimpl.LstatFunc
	ReadDirFunc       = zipimpl.ReadDirFunc
	ReadlinkFunc      = zipimpl.ReadlinkFunc
	MkdirAllFunc      = zipimpl.MkdirAllFunc
	RemoveFunc        = zipimpl.RemoveFunc
	RenameFunc        = zipimpl.RenameFunc
	OpenZipReaderFunc = zipimpl.OpenZipReaderFunc
	CreateTempFunc    = zipimpl.CreateTempFunc
	TempFile          = zipimpl.TempFile
)

// WithDirPerm sets the directory permission used when creating archive output/extract directories.
func WithDirPerm(perm os.FileMode) ArchiveOption { return zipimpl.WithDirPerm(perm) }

// WithFilePerm sets the file permission used when archive metadata is not preserved.
func WithFilePerm(perm os.FileMode) ArchiveOption { return zipimpl.WithFilePerm(perm) }

// WithOverwrite controls whether an existing output/extracted file may be overwritten.
func WithOverwrite(overwrite bool) ArchiveOption { return zipimpl.WithOverwrite(overwrite) }

// WithPreserveMode controls whether extracted files keep mode bits from the archive.
func WithPreserveMode(preserve bool) ArchiveOption { return zipimpl.WithPreserveMode(preserve) }

// WithSourceDir controls whether source directory names are included in newly created ZIP archives.
func WithSourceDir(withSrcDir bool) ArchiveOption { return zipimpl.WithSourceDir(withSrcDir) }

// WithFileFilter sets the source path filter used by newly created ZIP archives.
func WithFileFilter(filter FileFilter) ArchiveOption { return zipimpl.WithFileFilter(filter) }

// WithCompressionMethod sets the ZIP compression method used for newly created entries.
func WithCompressionMethod(method uint16) ArchiveOption { return zipimpl.WithCompressionMethod(method) }

// WithCompressionLevel sets the deflate compression level used for newly created entries.
func WithCompressionLevel(level int) ArchiveOption { return zipimpl.WithCompressionLevel(level) }

// WithMaxBytes limits bytes read from archive entries, decompressed streams, or compression inputs.
func WithMaxBytes(n int64) ArchiveOption { return zipimpl.WithMaxBytes(n) }

// WithOpen sets the function used to open source files for reading.
func WithOpen(open OpenFunc) ArchiveOption { return zipimpl.WithOpen(open) }

// WithReadFile sets the function used to read a complete source file.
func WithReadFile(readFile ReadFileFunc) ArchiveOption { return zipimpl.WithReadFile(readFile) }

// WithOpenFile sets the function used to open archive/extracted files for writing.
func WithOpenFile(openFile OpenFileFunc) ArchiveOption { return zipimpl.WithOpenFile(openFile) }

// WithEvalSymlinks sets the function used to resolve extraction paths for symlink escape checks.
func WithEvalSymlinks(evalSymlinks EvalSymlinksFunc) ArchiveOption {
	return zipimpl.WithEvalSymlinks(evalSymlinks)
}

// WithStat sets the function used to inspect existing archive paths.
func WithStat(stat StatFunc) ArchiveOption { return zipimpl.WithStat(stat) }

// WithLstat sets the function used to inspect source paths without following symlinks.
func WithLstat(lstat LstatFunc) ArchiveOption { return zipimpl.WithLstat(lstat) }

// WithReadDir sets the function used to enumerate source directories.
func WithReadDir(readDir ReadDirFunc) ArchiveOption { return zipimpl.WithReadDir(readDir) }

// WithReadlink sets the function used to read symlink targets.
func WithReadlink(readlink ReadlinkFunc) ArchiveOption { return zipimpl.WithReadlink(readlink) }

// WithMkdirAll sets the function used to create directory trees.
func WithMkdirAll(mkdirAll MkdirAllFunc) ArchiveOption { return zipimpl.WithMkdirAll(mkdirAll) }

// WithRemove sets the function used to remove temporary files.
func WithRemove(remove RemoveFunc) ArchiveOption { return zipimpl.WithRemove(remove) }

// WithRename sets the function used to move completed temporary archives into place.
func WithRename(rename RenameFunc) ArchiveOption { return zipimpl.WithRename(rename) }

// WithOpenZipReader sets the function used to open existing ZIP archives for reading.
func WithOpenZipReader(openZipReader OpenZipReaderFunc) ArchiveOption {
	return zipimpl.WithOpenZipReader(openZipReader)
}

// WithCreateTemp sets the function used to create temporary archives for append operations.
func WithCreateTemp(createTemp CreateTempFunc) ArchiveOption {
	return zipimpl.WithCreateTemp(createTemp)
}

// Open opens a ZIP file for reading.
func Open(path string) (*archivezip.ReadCloser, error) { return OpenWithOptions(path) }

// OpenWithOptions opens a ZIP file for reading with per-call options.
func OpenWithOptions(path string, opts ...ArchiveOption) (*archivezip.ReadCloser, error) {
	return zipimpl.OpenWithOptions(path, opts...)
}

// NewWriter returns a ZIP writer for out.
func NewWriter(out io.Writer) *archivezip.Writer { return zipimpl.NewWriter(out) }

// GetStream returns a reader for entry.
func GetStream(entry *archivezip.File) (io.ReadCloser, error) { return zipimpl.GetStream(entry) }

// Append appends srcPath into zipPath by rewriting the archive.
func Append(zipPath, srcPath string) error { return AppendWithOptions(zipPath, srcPath) }

// AppendWithOptions appends srcPath into zipPath by rewriting the archive with per-call options.
func AppendWithOptions(zipPath, srcPath string, opts ...ArchiveOption) error {
	return zipimpl.AppendWithOptions(zipPath, srcPath, opts...)
}

// Zip creates an archive next to srcPath and returns the archive path.
func Zip(srcPath string) (string, error) { return zipimpl.Zip(srcPath) }

// ZipTo creates an archive at zipPath from srcPath.
func ZipTo(srcPath, zipPath string, withSrcDir bool) error {
	return zipimpl.ZipTo(srcPath, zipPath, withSrcDir)
}

// ZipFiles creates a ZIP archive from source files or directories.
func ZipFiles(dest string, withSrcDir bool, srcFiles ...string) error {
	return ZipFilesWithOptions(dest, withSrcDir, srcFiles)
}

// ZipFilesWithOptions creates a ZIP archive from source files or directories with per-call options.
func ZipFilesWithOptions(dest string, withSrcDir bool, srcFiles []string, opts ...ArchiveOption) error {
	return zipimpl.ZipFilesWithOptions(dest, srcFiles, append([]ArchiveOption{zipimpl.WithSourceDir(withSrcDir)}, opts...)...)
}

// ZipFilesUsingOptions creates a ZIP archive from source files or directories using only functional options.
func ZipFilesUsingOptions(dest string, srcFiles []string, opts ...ArchiveOption) error {
	return zipimpl.ZipFilesWithOptions(dest, srcFiles, opts...)
}

// ZipFilesFilter creates a ZIP archive and filters source paths.
func ZipFilesFilter(dest string, withSrcDir bool, filter FileFilter, srcFiles ...string) error {
	return ZipFilesFilterWithOptions(dest, withSrcDir, filter, srcFiles)
}

// ZipFilesFilterWithOptions creates a ZIP archive with source filtering and per-call options.
func ZipFilesFilterWithOptions(dest string, withSrcDir bool, filter FileFilter, srcFiles []string, opts ...ArchiveOption) error {
	return zipimpl.ZipFilesFilterWithOptions(dest, withSrcDir, filter, srcFiles, opts...)
}

// ZipToWriter writes source files or directories into out as a ZIP archive.
func ZipToWriter(out io.Writer, withSrcDir bool, filter FileFilter, srcFiles ...string) error {
	return ZipToWriterWithOptions(out, withSrcDir, filter, srcFiles)
}

// ZipToWriterWithOptions writes source files or directories into out as a ZIP archive with per-call options.
func ZipToWriterWithOptions(out io.Writer, withSrcDir bool, filter FileFilter, srcFiles []string, opts ...ArchiveOption) error {
	return zipimpl.ZipToWriterWithOptions(out, srcFiles, append([]ArchiveOption{zipimpl.WithSourceDir(withSrcDir), zipimpl.WithFileFilter(filter)}, opts...)...)
}

// ZipToWriterUsingOptions writes source files or directories into out using only functional options.
func ZipToWriterUsingOptions(out io.Writer, srcFiles []string, opts ...ArchiveOption) error {
	return zipimpl.ZipToWriterWithOptions(out, srcFiles, opts...)
}

// ZipData creates or overwrites zipFile and adds one text entry.
func ZipData(zipFile, path, data string) error { return zipimpl.ZipData(zipFile, path, data) }

// ZipBytes creates or overwrites zipFile and adds one byte entry.
func ZipBytes(zipFile, path string, data []byte) error { return zipimpl.ZipBytes(zipFile, path, data) }

// ZipEntries creates or overwrites zipFile and adds in-memory entries.
func ZipEntries(zipFile string, entries ...EntryData) error {
	return ZipEntriesWithOptions(zipFile, entries)
}

// ZipEntriesWithOptions creates or overwrites zipFile and adds in-memory entries with per-call options.
func ZipEntriesWithOptions(zipFile string, entries []EntryData, opts ...ArchiveOption) error {
	return zipimpl.ZipEntriesWithOptions(zipFile, entries, opts...)
}

// ZipEntriesToWriter writes in-memory entries into out as a ZIP archive.
func ZipEntriesToWriter(out io.Writer, entries ...EntryData) error {
	return ZipEntriesToWriterWithOptions(out, entries)
}

// ZipEntriesToWriterWithOptions writes in-memory entries into out as a ZIP archive with per-call options.
func ZipEntriesToWriterWithOptions(out io.Writer, entries []EntryData, opts ...ArchiveOption) error {
	return zipimpl.ZipEntriesToWriterWithOptions(out, entries, opts...)
}

// ZipStreams creates or overwrites zipFile and adds stream entries.
func ZipStreams(zipFile string, entries ...StreamEntry) error {
	return ZipStreamsWithOptions(zipFile, entries)
}

// ZipStreamsWithOptions creates or overwrites zipFile and adds stream entries with per-call options.
func ZipStreamsWithOptions(zipFile string, entries []StreamEntry, opts ...ArchiveOption) error {
	return zipimpl.ZipStreamsWithOptions(zipFile, entries, opts...)
}

// ZipStreamsToWriter writes stream entries into out as a ZIP archive.
func ZipStreamsToWriter(out io.Writer, entries ...StreamEntry) error {
	return ZipStreamsToWriterWithOptions(out, entries)
}

// ZipStreamsToWriterWithOptions writes stream entries into out as a ZIP archive with per-call options.
func ZipStreamsToWriterWithOptions(out io.Writer, entries []StreamEntry, opts ...ArchiveOption) error {
	return zipimpl.ZipStreamsToWriterWithOptions(out, entries, opts...)
}

// Unzip extracts zipFile into a sibling directory named after the archive.
func Unzip(zipFile string) (string, error) { return zipimpl.Unzip(zipFile) }

// UnzipTo extracts zipFile into destDir.
func UnzipTo(zipFile, destDir string) error { return UnzipToWithOptions(zipFile, destDir) }

// UnzipToLimit extracts zipFile into destDir and optionally limits total uncompressed size.
func UnzipToLimit(zipFile, destDir string, limit int64) error {
	return zipimpl.UnzipToLimit(zipFile, destDir, limit)
}

// UnzipToWithOptions extracts zipFile into destDir with per-call options.
func UnzipToWithOptions(zipFile, destDir string, opts ...ArchiveOption) error {
	return zipimpl.UnzipToWithOptions(zipFile, destDir, opts...)
}

// UnzipReaderTo extracts archive reader contents into destDir.
func UnzipReaderTo(r *archivezip.Reader, destDir string) error {
	return UnzipReaderToWithOptions(r, destDir)
}

// UnzipReaderToLimit extracts archive reader contents into destDir and optionally limits total size.
func UnzipReaderToLimit(r *archivezip.Reader, destDir string, limit int64) error {
	return zipimpl.UnzipReaderToLimit(r, destDir, limit)
}

// UnzipReaderToWithOptions extracts archive reader contents into destDir with per-call options.
func UnzipReaderToWithOptions(r *archivezip.Reader, destDir string, opts ...ArchiveOption) error {
	return zipimpl.UnzipReaderToWithOptions(r, destDir, opts...)
}

// Get returns a reader for the named entry in zipFile.
func Get(zipFile, name string) (io.ReadCloser, error) { return GetWithOptions(zipFile, name) }

// GetWithOptions returns a reader for the named entry in zipFile with per-call options.
func GetWithOptions(zipFile, name string, opts ...ArchiveOption) (io.ReadCloser, error) {
	return zipimpl.GetWithOptions(zipFile, name, opts...)
}

// GetBytes returns the content of the named entry in zipFile.
func GetBytes(zipFile, name string) ([]byte, error) { return GetBytesWithOptions(zipFile, name) }

// GetBytesWithOptions returns the content of the named entry in zipFile with per-call options.
func GetBytesWithOptions(zipFile, name string, opts ...ArchiveOption) ([]byte, error) {
	return zipimpl.GetBytesWithOptions(zipFile, name, opts...)
}

// Read walks every archive entry and calls consumer.
func Read(zipFile string, consumer func(*archivezip.File) error) error {
	return ReadWithOptions(zipFile, consumer)
}

// ReadWithOptions walks every archive entry and calls consumer using per-call options.
func ReadWithOptions(zipFile string, consumer func(*archivezip.File) error, opts ...ArchiveOption) error {
	return zipimpl.ReadWithOptions(zipFile, consumer, opts...)
}

// ListFileNames returns direct file names under dir inside zipFile.
func ListFileNames(zipFile, dir string) ([]string, error) {
	return ListFileNamesWithOptions(zipFile, dir)
}

// ListFileNamesWithOptions returns direct file names under dir inside zipFile using per-call options.
func ListFileNamesWithOptions(zipFile, dir string, opts ...ArchiveOption) ([]string, error) {
	return zipimpl.ListFileNamesWithOptions(zipFile, dir, opts...)
}

// Gzip compresses data using gzip.
func Gzip(data []byte) ([]byte, error) { return GzipWithOptions(data) }

// GzipWithOptions compresses data using gzip with per-call options.
func GzipWithOptions(data []byte, opts ...ArchiveOption) ([]byte, error) {
	return zipimpl.GzipWithOptions(data, opts...)
}

// GzipString compresses text using gzip.
func GzipString(content string) ([]byte, error) { return zipimpl.GzipString(content) }

// GzipFile compresses a file using gzip and returns compressed bytes.
func GzipFile(path string) ([]byte, error) { return GzipFileWithOptions(path) }

// GzipFileWithOptions compresses a file using gzip and per-call options.
func GzipFileWithOptions(path string, opts ...ArchiveOption) ([]byte, error) {
	return zipimpl.GzipFileWithOptions(path, opts...)
}

// GzipReader compresses all bytes from r using gzip.
func GzipReader(r io.Reader, estimatedLength int) ([]byte, error) {
	return GzipReaderWithOptions(r, estimatedLength)
}

// GzipReaderWithOptions compresses all bytes from r using gzip with per-call options.
func GzipReaderWithOptions(r io.Reader, estimatedLength int, opts ...ArchiveOption) ([]byte, error) {
	return zipimpl.GzipReaderWithOptions(r, estimatedLength, opts...)
}

// UnGzip decompresses gzip data.
func UnGzip(data []byte) ([]byte, error) { return UnGzipWithOptions(data) }

// UnGzipWithOptions decompresses gzip data with per-call options.
func UnGzipWithOptions(data []byte, opts ...ArchiveOption) ([]byte, error) {
	return zipimpl.UnGzipWithOptions(data, opts...)
}

// Gunzip decompresses gzip data.
func Gunzip(data []byte) ([]byte, error) { return zipimpl.Gunzip(data) }

// UnGzipString decompresses gzip data and returns text.
func UnGzipString(data []byte) (string, error) { return zipimpl.UnGzipString(data) }

// UnGzipReader decompresses all gzip bytes from r.
func UnGzipReader(r io.Reader, estimatedLength int) ([]byte, error) {
	return UnGzipReaderWithOptions(r, estimatedLength)
}

// UnGzipReaderWithOptions decompresses gzip bytes from r with per-call options.
func UnGzipReaderWithOptions(r io.Reader, estimatedLength int, opts ...ArchiveOption) ([]byte, error) {
	return zipimpl.UnGzipReaderWithOptions(r, estimatedLength, opts...)
}

// Zlib compresses data using zlib with the default compression level.
func Zlib(data []byte) ([]byte, error) { return zipimpl.Zlib(data) }

// ZlibString compresses text using zlib with the specified compression level.
func ZlibString(content string, level int) ([]byte, error) { return zipimpl.ZlibString(content, level) }

// ZlibFile compresses a file using zlib with the specified compression level.
func ZlibFile(path string, level int) ([]byte, error) { return ZlibFileWithOptions(path, level) }

// ZlibFileWithOptions compresses a file using zlib with the specified compression level and per-call options.
func ZlibFileWithOptions(path string, level int, opts ...ArchiveOption) ([]byte, error) {
	return zipimpl.ZlibFileWithOptions(path, level, opts...)
}

// ZlibLevel compresses data using zlib with the specified compression level.
func ZlibLevel(data []byte, level int) ([]byte, error) { return zipimpl.ZlibLevel(data, level) }

// ZlibLevelWithOptions compresses data using zlib with the specified compression level and per-call options.
func ZlibLevelWithOptions(data []byte, level int, opts ...ArchiveOption) ([]byte, error) {
	return zipimpl.ZlibLevelWithOptions(data, level, opts...)
}

// ZlibReader compresses all bytes from r using zlib with the specified compression level.
func ZlibReader(r io.Reader, level, estimatedLength int) ([]byte, error) {
	return zipimpl.ZlibReader(r, level, estimatedLength)
}

// ZlibReaderWithOptions compresses all bytes from r using zlib with the specified compression level and per-call options.
func ZlibReaderWithOptions(r io.Reader, level, estimatedLength int, opts ...ArchiveOption) ([]byte, error) {
	return zipimpl.ZlibReaderWithOptions(r, level, estimatedLength, opts...)
}

// UnZlib decompresses zlib data.
func UnZlib(data []byte) ([]byte, error) { return UnZlibWithOptions(data) }

// UnZlibWithOptions decompresses zlib data with per-call options.
func UnZlibWithOptions(data []byte, opts ...ArchiveOption) ([]byte, error) {
	return zipimpl.UnZlibWithOptions(data, opts...)
}

// Unzlib decompresses zlib data.
func Unzlib(data []byte) ([]byte, error) { return zipimpl.Unzlib(data) }

// UnZlibString decompresses zlib data and returns text.
func UnZlibString(data []byte) (string, error) { return zipimpl.UnZlibString(data) }

// UnZlibReader decompresses all zlib bytes from r.
func UnZlibReader(r io.Reader, estimatedLength int) ([]byte, error) {
	return UnZlibReaderWithOptions(r, estimatedLength)
}

// UnZlibReaderWithOptions decompresses zlib bytes from r with per-call options.
func UnZlibReaderWithOptions(r io.Reader, estimatedLength int, opts ...ArchiveOption) ([]byte, error) {
	return zipimpl.UnZlibReaderWithOptions(r, estimatedLength, opts...)
}

// ReadFile reads a file from disk. It is useful when composing in-memory archive entries.
func ReadFile(path string) ([]byte, error) { return ReadFileWithOptions(path) }

// ReadFileWithOptions reads a file using per-call archive options.
func ReadFileWithOptions(path string, opts ...ArchiveOption) ([]byte, error) {
	return zipimpl.ReadFileWithOptions(path, opts...)
}
