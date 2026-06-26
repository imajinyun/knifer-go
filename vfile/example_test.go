package vfile_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/imajinyun/knifer-go"
	"github.com/imajinyun/knifer-go/vfile"
)

func Example_cookbookReadWriteTextFile() {
	dir, cleanup := exampleTempDir()
	defer cleanup()

	path := filepath.Join(dir, "note.txt")
	if err := vfile.WriteFileString(path, "hello"); err != nil {
		fmt.Println(err)
		return
	}
	content, err := vfile.ReadFileString(path)
	fmt.Println(content, err)
	// Output: hello <nil>
}

func Example_cookbookAppendWithoutClobbering() {
	dir, cleanup := exampleTempDir()
	defer cleanup()

	path := filepath.Join(dir, "app.log")
	_ = vfile.WriteFileString(path, "start")
	_ = vfile.AppendFileString(path, "\nstop")
	content, _ := vfile.ReadFileString(path)
	fmt.Println(content)
	// Output:
	// start
	// stop
}

func Example_cookbookCopyThenMove() {
	dir, cleanup := exampleTempDir()
	defer cleanup()

	src := filepath.Join(dir, "src.txt")
	backup := filepath.Join(dir, "backup", "src.txt")
	moved := filepath.Join(dir, "moved.txt")
	_ = vfile.WriteFileString(src, "payload")
	_ = vfile.CopyFile(src, backup)
	_ = vfile.CopyFile(src, moved)
	_ = vfile.Del(src)

	fmt.Println(vfile.Exists(backup), vfile.Exists(src), vfile.Exists(moved))
	// Output: true false true
}

func Example_cookbookCheckExistenceBeforeOptionalWork() {
	dir, cleanup := exampleTempDir()
	defer cleanup()

	path := filepath.Join(dir, "optional.txt")
	if !vfile.Exists(path) {
		_ = vfile.WriteFileString(path, "created")
	}
	fmt.Println(vfile.Exists(path))
	// Output: true
}

func Example_cookbookTemporaryDirectory() {
	dir, cleanup := exampleTempDir()
	defer cleanup()

	path := filepath.Join(dir, "nested", "note.txt")
	_ = vfile.WriteFileString(path, "isolated")
	fmt.Println(vfile.IsFile(path), strings.HasPrefix(path, dir))
	// Output: true true
}

func Example_cookbookExplicitFileError() {
	dir, cleanup := exampleTempDir()
	defer cleanup()

	missing := filepath.Join(dir, "missing.txt")
	_, err := vfile.ReadFileString(missing)
	fmt.Println(errors.Is(err, knifer.ErrCodeNotFound))
	// Output: true
}

func ExampleMainName() {
	fmt.Println(vfile.MainName("/tmp/report.csv"))
	// Output: report
}

func ExampleExtension() {
	fmt.Println(vfile.Extension("/tmp/report.csv"))
	// Output: csv
}

func ExampleReadString() {
	content, _ := vfile.ReadString(strings.NewReader("hello"))
	fmt.Println(content)
	// Output: hello
}

func ExampleWriteFileString() {
	dir, err := os.MkdirTemp("", "knifer-go-vfile-example-*")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "note.txt")
	if err := vfile.WriteFileString(path, "hello"); err != nil {
		fmt.Println(err)
		return
	}
	content, _ := vfile.ReadFileString(path)
	fmt.Println(content)
	// Output: hello
}

func ExampleExists() {
	dir, err := os.MkdirTemp("", "knifer-go-vfile-example-*")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer os.RemoveAll(dir)

	fmt.Println(vfile.Exists(dir))
	fmt.Println(vfile.Exists(filepath.Join(dir, "missing.txt")))
	// Output:
	// true
	// false
}

func ExampleAppendFileString() {
	dir, cleanup := exampleTempDir()
	defer cleanup()

	path := filepath.Join(dir, "note.txt")
	_ = vfile.WriteFileString(path, "hello")
	_ = vfile.AppendFileString(path, "!")
	content, _ := vfile.ReadFileString(path)
	fmt.Println(content)
	// Output: hello!
}

func ExampleAppendFileStringWithOptions() {
	dir, cleanup := exampleTempDir()
	defer cleanup()

	path := filepath.Join(dir, "note.txt")
	_ = vfile.AppendFileStringWithOptions(path, "first", vfile.WithOverwrite(false))
	err := vfile.AppendFileStringWithOptions(path, "second", vfile.WithOverwrite(false))
	fmt.Println(errors.Is(err, knifer.ErrCodeInternal))
	// Output: true
}

func ExampleCloseQuietly() {
	closer := &exampleCloser{}
	vfile.CloseQuietly(closer)
	vfile.CloseQuietly(nil)
	fmt.Println(closer.closed)
	// Output: true
}

func ExampleCopy() {
	var dst bytes.Buffer
	n, _ := vfile.Copy(&dst, strings.NewReader("copy"))
	fmt.Println(n, dst.String())
	// Output: 4 copy
}

func ExampleCopyWithOptions() {
	var dst bytes.Buffer
	n, err := vfile.CopyWithOptions(&dst, strings.NewReader("abcdef"), vfile.WithMaxBytes(3))
	fmt.Println(n, dst.String())
	fmt.Println(errors.Is(err, knifer.ErrCodeInvalidInput))
	// Output:
	// 3 abc
	// true
}

func ExampleCopyFile() {
	dir, cleanup := exampleTempDir()
	defer cleanup()

	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")
	_ = vfile.WriteFileString(src, "copy")
	_ = vfile.CopyFile(src, dst)
	content, _ := vfile.ReadFileString(dst)
	fmt.Println(content)
	// Output: copy
}

func ExampleCopyFileWithOptions() {
	dir, cleanup := exampleTempDir()
	defer cleanup()

	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")
	_ = vfile.WriteFileString(src, "copy")
	_ = vfile.CopyFileWithOptions(src, dst, vfile.WithFilePerm(0o600))
	info, _ := os.Stat(dst)
	fmt.Printf("%o\n", info.Mode().Perm())
	// Output: 600
}

func ExampleDel() {
	dir, cleanup := exampleTempDir()
	defer cleanup()

	path := filepath.Join(dir, "note.txt")
	_ = vfile.WriteFileString(path, "delete")
	_ = vfile.Del(path)
	fmt.Println(vfile.Exists(path))
	// Output: false
}

func ExampleDelWithOptions() {
	removed := false
	err := vfile.DelWithOptions("virtual.txt",
		vfile.WithStat(func(string) (fs.FileInfo, error) {
			return exampleFileInfo{name: "virtual.txt"}, nil
		}),
		vfile.WithRemoveAll(func(path string) error {
			removed = path == "virtual.txt"
			return nil
		}),
	)
	fmt.Println(err == nil, removed)
	// Output: true true
}

func ExampleExistsWithOptions() {
	exists := vfile.ExistsWithOptions("virtual.txt", vfile.WithStat(func(string) (fs.FileInfo, error) {
		return exampleFileInfo{name: "virtual.txt"}, nil
	}))
	fmt.Println(exists)
	// Output: true
}

func ExampleIsFile() {
	dir, cleanup := exampleTempDir()
	defer cleanup()

	path := filepath.Join(dir, "note.txt")
	_ = vfile.WriteFileString(path, "file")
	fmt.Println(vfile.IsFile(path))
	fmt.Println(vfile.IsFile(dir))
	// Output:
	// true
	// false
}

func ExampleIsFileWithOptions() {
	isFile := vfile.IsFileWithOptions("virtual.txt", vfile.WithStat(func(string) (fs.FileInfo, error) {
		return exampleFileInfo{name: "virtual.txt"}, nil
	}))
	fmt.Println(isFile)
	// Output: true
}

func ExampleIsDirectory() {
	dir, cleanup := exampleTempDir()
	defer cleanup()

	fmt.Println(vfile.IsDirectory(dir))
	fmt.Println(vfile.IsDirectory(filepath.Join(dir, "missing")))
	// Output:
	// true
	// false
}

func ExampleIsDirectoryWithOptions() {
	isDir := vfile.IsDirectoryWithOptions("virtual-dir", vfile.WithStat(func(string) (fs.FileInfo, error) {
		return exampleFileInfo{name: "virtual-dir", dir: true}, nil
	}))
	fmt.Println(isDir)
	// Output: true
}

func ExampleMkdir() {
	dir, cleanup := exampleTempDir()
	defer cleanup()

	path := filepath.Join(dir, "a", "b")
	_ = vfile.Mkdir(path)
	fmt.Println(vfile.IsDirectory(path))
	// Output: true
}

func ExampleMkdirWithOptions() {
	var gotPerm fs.FileMode
	_ = vfile.MkdirWithOptions("virtual-dir",
		vfile.WithMkdirPerm(0o700),
		vfile.WithMkdirAll(func(_ string, perm fs.FileMode) error {
			gotPerm = perm
			return nil
		}),
	)
	fmt.Printf("%o\n", gotPerm)
	// Output: 700
}

func ExampleReadAll() {
	b, _ := vfile.ReadAll(strings.NewReader("all"))
	fmt.Println(string(b))
	// Output: all
}

func ExampleReadAllWithOptions() {
	b, _ := vfile.ReadAllWithOptions(strings.NewReader("abcd"), vfile.WithMaxBytes(1), vfile.WithUnlimitedRead())
	fmt.Println(string(b))
	// Output: abcd
}

func ExampleReadStringWithOptions() {
	s, _ := vfile.ReadStringWithOptions(strings.NewReader("abc"), vfile.WithMaxBytes(3))
	fmt.Println(s)
	// Output: abc
}

func ExampleReadLines() {
	lines, _ := vfile.ReadLines(strings.NewReader("a\nb\n"))
	fmt.Println(strings.Join(lines, ","))
	// Output: a,b
}

func ExampleReadLinesWithOptions() {
	lines, _ := vfile.ReadLinesWithOptions(strings.NewReader("abc"), vfile.WithInitialLineBuffer(1), vfile.WithMaxLineBytes(4))
	fmt.Println(lines[0])
	// Output: abc
}

func ExampleReadChunks() {
	var chunks []string
	_ = vfile.ReadChunks(strings.NewReader("abcd"), func(chunk []byte) error {
		chunks = append(chunks, string(chunk))
		return nil
	})
	fmt.Println(strings.Join(chunks, ""))
	// Output: abcd
}

func ExampleReadChunksWithOptions() {
	var chunks []string
	_ = vfile.ReadChunksWithOptions(strings.NewReader("abcdef"), func(chunk []byte) error {
		chunks = append(chunks, string(chunk))
		return nil
	}, vfile.WithBufferSize(2))
	fmt.Println(strings.Join(chunks, "|"))
	// Output: ab|cd|ef
}

func ExampleReadFileString() {
	path, cleanup := exampleFile("note.txt", "hello")
	defer cleanup()

	content, _ := vfile.ReadFileString(path)
	fmt.Println(content)
	// Output: hello
}

func ExampleReadFileStringWithOptions() {
	content, _ := vfile.ReadFileStringWithOptions("virtual.txt", vfile.WithOpen(func(path string) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader("virtual:" + path)), nil
	}))
	fmt.Println(content)
	// Output: virtual:virtual.txt
}

func ExampleReadFileBytes() {
	path, cleanup := exampleFile("note.txt", "bytes")
	defer cleanup()

	data, _ := vfile.ReadFileBytes(path)
	fmt.Println(string(data))
	// Output: bytes
}

func ExampleReadFileBytesWithOptions() {
	data, _ := vfile.ReadFileBytesWithOptions("virtual.txt", vfile.WithOpen(func(string) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader("bytes")), nil
	}))
	fmt.Println(string(data))
	// Output: bytes
}

func ExampleReadFileLines() {
	path, cleanup := exampleFile("lines.txt", "a\nb\n")
	defer cleanup()

	lines, _ := vfile.ReadFileLines(path)
	fmt.Println(strings.Join(lines, ","))
	// Output: a,b
}

func ExampleReadFileLinesWithOptions() {
	lines, _ := vfile.ReadFileLinesWithOptions("virtual.txt",
		vfile.WithOpen(func(string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("abc")), nil
		}),
		vfile.WithInitialLineBuffer(1),
		vfile.WithMaxLineBytes(4),
	)
	fmt.Println(lines[0])
	// Output: abc
}

func ExampleReadFileChunks() {
	path, cleanup := exampleFile("chunks.txt", "abcd")
	defer cleanup()

	var chunks []string
	_ = vfile.ReadFileChunks(path, func(chunk []byte) error {
		chunks = append(chunks, string(chunk))
		return nil
	})
	fmt.Println(strings.Join(chunks, ""))
	// Output: abcd
}

func ExampleReadFileChunksWithOptions() {
	path, cleanup := exampleFile("chunks.txt", "abcde")
	defer cleanup()

	var chunks []string
	_ = vfile.ReadFileChunksWithOptions(path, func(chunk []byte) error {
		chunks = append(chunks, string(chunk))
		return nil
	}, vfile.WithBufferSize(2))
	fmt.Println(strings.Join(chunks, "|"))
	// Output: ab|cd|e
}

func ExampleReaderFromString() {
	content, _ := io.ReadAll(vfile.ReaderFromString("reader"))
	fmt.Println(string(content))
	// Output: reader
}

func ExampleSize() {
	path, cleanup := exampleFile("size.txt", "12345")
	defer cleanup()

	fmt.Println(vfile.Size(path))
	// Output: 5
}

func ExampleSizeWithOptions() {
	size := vfile.SizeWithOptions("virtual.txt", vfile.WithStat(func(string) (fs.FileInfo, error) {
		return exampleFileInfo{name: "virtual.txt", size: 42}, nil
	}))
	fmt.Println(size)
	// Output: 42
}

func ExampleTouch() {
	dir, cleanup := exampleTempDir()
	defer cleanup()

	path := filepath.Join(dir, "touch.txt")
	_ = vfile.Touch(path)
	fmt.Println(vfile.IsFile(path))
	// Output: true
}

func ExampleTouchWithOptions() {
	dir, cleanup := exampleTempDir()
	defer cleanup()

	path := filepath.Join(dir, "touch.txt")
	_ = vfile.TouchWithOptions(path, vfile.WithFilePerm(0o600))
	info, _ := os.Stat(path)
	fmt.Printf("%o\n", info.Mode().Perm())
	// Output: 600
}

func ExampleWriteFileBytes() {
	dir, cleanup := exampleTempDir()
	defer cleanup()

	path := filepath.Join(dir, "bytes.txt")
	_ = vfile.WriteFileBytes(path, []byte("bytes"))
	content, _ := vfile.ReadFileString(path)
	fmt.Println(content)
	// Output: bytes
}

func ExampleWriteFileBytesWithOptions() {
	dir, cleanup := exampleTempDir()
	defer cleanup()

	path := filepath.Join(dir, "bytes.txt")
	_ = vfile.WriteFileBytesWithOptions(path, []byte("bytes"), vfile.WithFilePerm(0o600))
	info, _ := os.Stat(path)
	fmt.Printf("%o\n", info.Mode().Perm())
	// Output: 600
}

func ExampleWriteFileStringWithOptions() {
	dir, cleanup := exampleTempDir()
	defer cleanup()

	path := filepath.Join(dir, "note.txt")
	_ = vfile.WriteFileStringWithOptions(path, "hello", vfile.WithFilePerm(0o600))
	content, _ := vfile.ReadFileString(path)
	fmt.Println(content)
	// Output: hello
}

func ExampleWithBufferSize() {
	var chunks []string
	_ = vfile.ReadChunksWithOptions(strings.NewReader("abcd"), func(chunk []byte) error {
		chunks = append(chunks, string(chunk))
		return nil
	}, vfile.WithBufferSize(2))
	fmt.Println(strings.Join(chunks, ","))
	// Output: ab,cd
}

func ExampleWithCreateParents() {
	dir, cleanup := exampleTempDir()
	defer cleanup()

	path := filepath.Join(dir, "missing", "note.txt")
	err := vfile.WriteFileString(path, "hello", vfile.WithCreateParents(false))
	fmt.Println(errors.Is(err, knifer.ErrCodeNotFound))
	// Output: true
}

func ExampleWithDirPerm() {
	var gotPerm fs.FileMode
	_ = vfile.WriteFileString("virtual/note.txt", "hello",
		vfile.WithDirPerm(0o700),
		vfile.WithMkdirAll(func(_ string, perm fs.FileMode) error {
			gotPerm = perm
			return nil
		}),
		vfile.WithOpenFile(func(string, int, fs.FileMode) (io.WriteCloser, error) {
			return exampleWriteCloser{Writer: io.Discard}, nil
		}),
	)
	fmt.Printf("%o\n", gotPerm)
	// Output: 700
}

func ExampleWithFilePerm() {
	var gotPerm fs.FileMode
	_ = vfile.WriteFileString("virtual.txt", "hello", vfile.WithFilePerm(0o600), vfile.WithOpenFile(func(_ string, _ int, perm fs.FileMode) (io.WriteCloser, error) {
		gotPerm = perm
		return exampleWriteCloser{Writer: io.Discard}, nil
	}))
	fmt.Printf("%o\n", gotPerm)
	// Output: 600
}

func ExampleWithInitialLineBuffer() {
	lines, _ := vfile.ReadLinesWithOptions(strings.NewReader("abc"), vfile.WithInitialLineBuffer(1), vfile.WithMaxLineBytes(4))
	fmt.Println(lines[0])
	// Output: abc
}

func ExampleWithMaxBytes() {
	_, err := vfile.ReadAllWithOptions(strings.NewReader("abcd"), vfile.WithMaxBytes(3))
	fmt.Println(errors.Is(err, knifer.ErrCodeInvalidInput))
	// Output: true
}

func ExampleWithMaxLineBytes() {
	_, err := vfile.ReadLinesWithOptions(strings.NewReader("abcd"), vfile.WithInitialLineBuffer(1), vfile.WithMaxLineBytes(3))
	fmt.Println(err != nil)
	// Output: true
}

func ExampleWithMkdirAll() {
	created := ""
	_ = vfile.MkdirWithOptions("virtual-dir", vfile.WithMkdirAll(func(path string, _ fs.FileMode) error {
		created = path
		return nil
	}))
	fmt.Println(created)
	// Output: virtual-dir
}

func ExampleWithMkdirPerm() {
	var gotPerm fs.FileMode
	_ = vfile.MkdirWithOptions("virtual-dir",
		vfile.WithMkdirPerm(0o700),
		vfile.WithMkdirAll(func(_ string, perm fs.FileMode) error {
			gotPerm = perm
			return nil
		}),
	)
	fmt.Printf("%o\n", gotPerm)
	// Output: 700
}

func ExampleWithOpen() {
	content, _ := vfile.ReadFileStringWithOptions("virtual.txt", vfile.WithOpen(func(path string) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader("opened:" + path)), nil
	}))
	fmt.Println(content)
	// Output: opened:virtual.txt
}

func ExampleWithOpenFile() {
	var dst bytes.Buffer
	_ = vfile.WriteFileString("virtual.txt", "hello", vfile.WithOpenFile(func(string, int, fs.FileMode) (io.WriteCloser, error) {
		return exampleWriteCloser{Writer: &dst}, nil
	}))
	fmt.Println(dst.String())
	// Output: hello
}

func ExampleWithOverwrite() {
	dir, cleanup := exampleTempDir()
	defer cleanup()

	path := filepath.Join(dir, "note.txt")
	_ = vfile.WriteFileString(path, "first")
	err := vfile.WriteFileString(path, "second", vfile.WithOverwrite(false))
	fmt.Println(errors.Is(err, knifer.ErrCodeInternal))
	// Output: true
}

func ExampleWithRemoveAll() {
	removed := ""
	_ = vfile.DelWithOptions("virtual.txt",
		vfile.WithStat(func(string) (fs.FileInfo, error) { return exampleFileInfo{name: "virtual.txt"}, nil }),
		vfile.WithRemoveAll(func(path string) error {
			removed = path
			return nil
		}),
	)
	fmt.Println(removed)
	// Output: virtual.txt
}

func ExampleWithStat() {
	exists := vfile.ExistsWithOptions("virtual.txt", vfile.WithStat(func(string) (fs.FileInfo, error) {
		return exampleFileInfo{name: "virtual.txt"}, nil
	}))
	fmt.Println(exists)
	// Output: true
}

func ExampleWithUnlimitedRead() {
	content, _ := vfile.ReadStringWithOptions(strings.NewReader("abcd"), vfile.WithMaxBytes(1), vfile.WithUnlimitedRead())
	fmt.Println(content)
	// Output: abcd
}

func exampleTempDir() (string, func()) {
	dir, err := os.MkdirTemp("", "knifer-go-vfile-example-*")
	if err != nil {
		panic(err)
	}
	return dir, func() { _ = os.RemoveAll(dir) }
}

func exampleFile(name, content string) (string, func()) {
	dir, cleanup := exampleTempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		cleanup()
		panic(err)
	}
	return path, cleanup
}

type exampleCloser struct{ closed bool }

func (c *exampleCloser) Close() error {
	c.closed = true
	return errors.New("ignored")
}

type exampleWriteCloser struct{ io.Writer }

func (w exampleWriteCloser) Close() error { return nil }

type exampleFileInfo struct {
	name string
	size int64
	dir  bool
}

func (i exampleFileInfo) Name() string       { return i.name }
func (i exampleFileInfo) Size() int64        { return i.size }
func (i exampleFileInfo) Mode() fs.FileMode  { return 0o644 }
func (i exampleFileInfo) ModTime() time.Time { return time.Unix(0, 0) }
func (i exampleFileInfo) IsDir() bool        { return i.dir }
func (i exampleFileInfo) Sys() any           { return nil }
