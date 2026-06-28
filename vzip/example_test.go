package vzip_test

import (
	archivezip "archive/zip"
	"bytes"
	"compress/flate"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/imajinyun/knifer-go/vzip"
)

func writeExampleArchive(entries ...vzip.EntryData) (string, func(), error) {
	dir, err := os.MkdirTemp("", "knifer-go-vzip-example-")
	if err != nil {
		return "", nil, err
	}
	cleanup := func() { _ = os.RemoveAll(dir) }

	archivePath := filepath.Join(dir, "example.zip")
	if err := vzip.ZipEntries(archivePath, entries...); err != nil {
		cleanup()
		return "", nil, err
	}
	return archivePath, cleanup, nil
}

func ExampleZipEntriesToWriter() {
	var archive bytes.Buffer
	if err := vzip.ZipEntriesToWriter(&archive,
		vzip.EntryData{Name: "config/app.setting", Data: []byte("name=knifer-go")},
		vzip.EntryData{Name: "README.txt", Data: []byte("docs")},
	); err != nil {
		fmt.Println(err)
		return
	}

	reader, err := archivezip.NewReader(bytes.NewReader(archive.Bytes()), int64(archive.Len()))
	if err != nil {
		fmt.Println(err)
		return
	}

	names := make([]string, 0)
	for _, file := range reader.File {
		if file.FileInfo().IsDir() || file.Name != "config/app.setting" {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			fmt.Println(err)
			return
		}
		_, _ = io.Copy(io.Discard, rc)
		_ = rc.Close()
		names = append(names, "app.setting")
	}
	sort.Strings(names)
	fmt.Println(names)
	// Output: [app.setting]
}

func ExampleZipEntriesToWriterWithOptions() {
	var archive bytes.Buffer
	if err := vzip.ZipEntriesToWriterWithOptions(&archive, []vzip.EntryData{
		{Name: "stored.txt", Data: []byte("stored")},
	}, vzip.WithCompressionMethod(archivezip.Store)); err != nil {
		fmt.Println(err)
		return
	}

	reader, err := archivezip.NewReader(bytes.NewReader(archive.Bytes()), int64(archive.Len()))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(reader.File[0].Name, reader.File[0].Method == archivezip.Store)
	// Output: stored.txt true
}

func ExampleNewWriter() {
	var archive bytes.Buffer
	writer := vzip.NewWriter(&archive)
	entry, err := writer.Create("manual.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	if _, err := entry.Write([]byte("manual")); err != nil {
		fmt.Println(err)
		return
	}
	if err := writer.Close(); err != nil {
		fmt.Println(err)
		return
	}

	reader, err := archivezip.NewReader(bytes.NewReader(archive.Bytes()), int64(archive.Len()))
	if err != nil {
		fmt.Println(err)
		return
	}
	stream, err := reader.File[0].Open()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer stream.Close()
	data, _ := io.ReadAll(stream)
	fmt.Println(reader.File[0].Name, string(data))
	// Output: manual.txt manual
}

func ExampleZipEntries() {
	archivePath, cleanup, err := writeExampleArchive(
		vzip.EntryData{Name: "config/app.yml", Data: []byte("name: knifer-go")},
		vzip.EntryData{Name: "README.md", Data: []byte("docs")},
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer cleanup()

	reader, err := vzip.Open(archivePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer reader.Close()

	names := make([]string, 0, len(reader.File))
	for _, file := range reader.File {
		names = append(names, file.Name)
	}
	sort.Strings(names)
	fmt.Println(names)
	// Output: [README.md config/app.yml]
}

func ExampleZipFiles() {
	dir, err := os.MkdirTemp("", "knifer-go-vzip-files-")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer os.RemoveAll(dir)
	file := filepath.Join(dir, "app.txt")
	if err := os.WriteFile(file, []byte("app"), 0o644); err != nil {
		fmt.Println(err)
		return
	}
	archivePath := filepath.Join(dir, "out.zip")
	if err := vzip.ZipFiles(archivePath, false, file); err != nil {
		fmt.Println(err)
		return
	}
	names, err := vzip.ListFileNames(archivePath, "")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(names)
	// Output: [app.txt]
}

func ExampleZipFilesFilter() {
	dir, err := os.MkdirTemp("", "knifer-go-vzip-filter-")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer os.RemoveAll(dir)
	keep := filepath.Join(dir, "keep.txt")
	skip := filepath.Join(dir, "skip.log")
	if err := os.WriteFile(keep, []byte("keep"), 0o644); err != nil {
		fmt.Println(err)
		return
	}
	if err := os.WriteFile(skip, []byte("skip"), 0o644); err != nil {
		fmt.Println(err)
		return
	}
	archivePath := filepath.Join(dir, "filtered.zip")
	filter := func(path string, info os.FileInfo) bool {
		return info.IsDir() || filepath.Ext(path) == ".txt"
	}
	if err := vzip.ZipFilesFilter(archivePath, false, filter, keep, skip); err != nil {
		fmt.Println(err)
		return
	}
	names, err := vzip.ListFileNames(archivePath, "")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(names)
	// Output: [keep.txt]
}

func ExampleZipData() {
	archivePath, cleanup, err := writeExampleArchive(vzip.EntryData{Name: "placeholder", Data: []byte("ignored")})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer cleanup()

	if err := vzip.ZipData(archivePath, "config/app.env", "enabled=true"); err != nil {
		fmt.Println(err)
		return
	}
	data, err := vzip.GetBytes(archivePath, "config/app.env")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(data))
	// Output: enabled=true
}

func ExampleZipBytes() {
	archivePath, cleanup, err := writeExampleArchive(vzip.EntryData{Name: "placeholder", Data: []byte("ignored")})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer cleanup()

	if err := vzip.ZipBytes(archivePath, "payload.bin", []byte{1, 2, 3}); err != nil {
		fmt.Println(err)
		return
	}
	data, err := vzip.GetBytes(archivePath, "payload.bin")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("% x\n", data)
	// Output: 01 02 03
}

func ExampleOpen() {
	archivePath, cleanup, err := writeExampleArchive(vzip.EntryData{Name: "docs/readme.txt", Data: []byte("docs")})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer cleanup()

	reader, err := vzip.Open(archivePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer reader.Close()

	fmt.Println(len(reader.File), reader.File[0].Name)
	// Output: 1 docs/readme.txt
}

func ExampleGet() {
	archivePath, cleanup, err := writeExampleArchive(vzip.EntryData{Name: "data.txt", Data: []byte("payload")})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer cleanup()

	reader, err := vzip.Get(archivePath, "data.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer reader.Close()
	data, _ := io.ReadAll(reader)
	fmt.Println(string(data))
	// Output: payload
}

func ExampleGetBytes() {
	archivePath, cleanup, err := writeExampleArchive(vzip.EntryData{Name: "data.txt", Data: []byte("payload")})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer cleanup()

	data, err := vzip.GetBytes(archivePath, "data.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(data))
	// Output: payload
}

func ExampleRead() {
	archivePath, cleanup, err := writeExampleArchive(
		vzip.EntryData{Name: "config/app.yml", Data: []byte("name: knifer-go")},
		vzip.EntryData{Name: "data.txt", Data: []byte("payload")},
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer cleanup()

	names := make([]string, 0)
	if err := vzip.Read(archivePath, func(file *archivezip.File) error {
		names = append(names, file.Name)
		return nil
	}); err != nil {
		fmt.Println(err)
		return
	}
	sort.Strings(names)
	fmt.Println(names)
	// Output: [config/app.yml data.txt]
}

func ExampleListFileNames() {
	archivePath, cleanup, err := writeExampleArchive(
		vzip.EntryData{Name: "config/app.yml", Data: []byte("app")},
		vzip.EntryData{Name: "config/db.yml", Data: []byte("db")},
		vzip.EntryData{Name: "config/nested/secret.yml", Data: []byte("secret")},
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer cleanup()

	names, err := vzip.ListFileNames(archivePath, "config")
	if err != nil {
		fmt.Println(err)
		return
	}
	sort.Strings(names)
	fmt.Println(names)
	// Output: [app.yml db.yml]
}

func ExampleUnzipReaderTo() {
	var archive bytes.Buffer
	if err := vzip.ZipEntriesToWriter(&archive, vzip.EntryData{Name: "docs/readme.txt", Data: []byte("docs")}); err != nil {
		fmt.Println(err)
		return
	}
	reader, err := archivezip.NewReader(bytes.NewReader(archive.Bytes()), int64(archive.Len()))
	if err != nil {
		fmt.Println(err)
		return
	}

	dir, err := os.MkdirTemp("", "knifer-go-vzip-unzip-reader-")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer os.RemoveAll(dir)
	if err := vzip.UnzipReaderTo(reader, dir); err != nil {
		fmt.Println(err)
		return
	}
	data, err := os.ReadFile(filepath.Join(dir, "docs", "readme.txt"))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(data))
	// Output: docs
}

func ExampleUnzipTo() {
	archivePath, cleanup, err := writeExampleArchive(vzip.EntryData{Name: "hello.txt", Data: []byte("hello")})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer cleanup()

	dir, err := os.MkdirTemp("", "knifer-go-vzip-unzip-to-")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer os.RemoveAll(dir)
	if err := vzip.UnzipTo(archivePath, dir); err != nil {
		fmt.Println(err)
		return
	}
	data, err := os.ReadFile(filepath.Join(dir, "hello.txt"))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(data))
	// Output: hello
}

func ExampleGzipString() {
	compressed, err := vzip.GzipString("hello")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(len(compressed) > 0)
	// Output: true
}

func ExampleUnGzipString() {
	compressed, _ := vzip.GzipString("hello")
	plain, err := vzip.UnGzipString(compressed)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(plain)
	// Output: hello
}

func ExampleZlibString() {
	compressed, err := vzip.ZlibString("hello", 6)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(len(compressed) > 0)
	// Output: true
}

func ExampleUnZlibString() {
	compressed, _ := vzip.ZlibString("hello", 6)
	plain, err := vzip.UnZlibString(compressed)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(plain)
	// Output: hello
}

func ExampleZipStreamsToWriter() {
	var archive bytes.Buffer
	if err := vzip.ZipStreamsToWriter(&archive,
		vzip.StreamEntry{Name: "stream.txt", Reader: bytes.NewReader([]byte("stream"))},
	); err != nil {
		fmt.Println(err)
		return
	}

	reader, err := archivezip.NewReader(bytes.NewReader(archive.Bytes()), int64(archive.Len()))
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, file := range reader.File {
		fmt.Println(file.Name)
	}
	// Output: stream.txt
}

func ExampleGetStream() {
	var archive bytes.Buffer
	if err := vzip.ZipEntriesToWriter(&archive, vzip.EntryData{Name: "data.txt", Data: []byte("payload")}); err != nil {
		fmt.Println(err)
		return
	}

	reader, err := archivezip.NewReader(bytes.NewReader(archive.Bytes()), int64(archive.Len()))
	if err != nil {
		fmt.Println(err)
		return
	}
	rc, err := vzip.GetStream(reader.File[0])
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rc.Close()

	data, _ := io.ReadAll(rc)
	fmt.Println(string(data))
	// Output: payload
}

func ExampleGzip_byteSlice() {
	compressed, err := vzip.Gzip([]byte("hello"))
	if err != nil {
		fmt.Println(err)
		return
	}
	plain, err := vzip.UnGzip(compressed)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(plain))
	// Output: hello
}

func ExampleAppend() {
	archivePath, cleanup, err := writeExampleArchive(vzip.EntryData{Name: "base.txt", Data: []byte("base")})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer cleanup()

	extra := filepath.Join(filepath.Dir(archivePath), "extra.txt")
	if err := os.WriteFile(extra, []byte("extra"), 0o644); err != nil {
		fmt.Println(err)
		return
	}
	if err := vzip.Append(archivePath, extra); err != nil {
		fmt.Println(err)
		return
	}
	data, _ := vzip.GetBytes(archivePath, "extra.txt")
	fmt.Println(string(data))
	// Output: extra
}

func ExampleAppendWithOptions() {
	archivePath, cleanup, err := writeExampleArchive(vzip.EntryData{Name: "base.txt", Data: []byte("base")})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer cleanup()

	extra := filepath.Join(filepath.Dir(archivePath), "extra.log")
	if err := os.WriteFile(extra, []byte("skip"), 0o644); err != nil {
		fmt.Println(err)
		return
	}
	filter := func(path string, info os.FileInfo) bool {
		return info.IsDir() || filepath.Ext(path) == ".txt"
	}
	err = vzip.AppendWithOptions(archivePath, extra, vzip.WithFileFilter(filter))
	fmt.Println(err == nil)
	names, _ := vzip.ListFileNames(archivePath, "")
	fmt.Println(names)
	// Output:
	// true
	// [base.txt]
}

func ExampleZip() {
	dir, err := os.MkdirTemp("", "knifer-go-vzip-zip-")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer os.RemoveAll(dir)

	source := filepath.Join(dir, "source.txt")
	if err := os.WriteFile(source, []byte("source"), 0o644); err != nil {
		fmt.Println(err)
		return
	}
	archivePath, err := vzip.Zip(source)
	if err != nil {
		fmt.Println(err)
		return
	}
	data, _ := vzip.GetBytes(archivePath, "source.txt")
	fmt.Println(string(data))
	// Output: source
}

func ExampleZipTo() {
	dir, err := os.MkdirTemp("", "knifer-go-vzip-zip-to-")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer os.RemoveAll(dir)

	src := filepath.Join(dir, "src")
	if err := os.MkdirAll(src, 0o755); err != nil {
		fmt.Println(err)
		return
	}
	if err := os.WriteFile(filepath.Join(src, "keep.txt"), []byte("keep"), 0o644); err != nil {
		fmt.Println(err)
		return
	}
	archivePath := filepath.Join(dir, "out.zip")
	if err := vzip.ZipTo(src, archivePath, true); err != nil {
		fmt.Println(err)
		return
	}
	data, _ := vzip.GetBytes(archivePath, "src/keep.txt")
	fmt.Println(string(data))
	// Output: keep
}

func ExampleZipFilesWithOptions() {
	dir, err := os.MkdirTemp("", "knifer-go-vzip-files-options-")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer os.RemoveAll(dir)

	source := filepath.Join(dir, "data.txt")
	if err := os.WriteFile(source, []byte("data"), 0o644); err != nil {
		fmt.Println(err)
		return
	}
	archivePath := filepath.Join(dir, "data.zip")
	err = vzip.ZipFilesWithOptions(
		archivePath,
		false,
		[]string{source},
		vzip.WithCompressionLevel(flate.BestSpeed),
	)
	fmt.Println(err == nil)
	data, _ := vzip.GetBytes(archivePath, "data.txt")
	fmt.Println(string(data))
	// Output:
	// true
	// data
}

func ExampleZipFilesUsingOptions() {
	dir, err := os.MkdirTemp("", "knifer-go-vzip-files-using-options-")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer os.RemoveAll(dir)

	source := filepath.Join(dir, "data.txt")
	if err := os.WriteFile(source, []byte("data"), 0o644); err != nil {
		fmt.Println(err)
		return
	}
	archivePath := filepath.Join(dir, "data.zip")
	if err := vzip.ZipFilesUsingOptions(archivePath, []string{source}, vzip.WithSourceDir(false)); err != nil {
		fmt.Println(err)
		return
	}
	names, _ := vzip.ListFileNames(archivePath, "")
	fmt.Println(names)
	// Output: [data.txt]
}

func ExampleZipFilesFilterWithOptions() {
	dir, err := os.MkdirTemp("", "knifer-go-vzip-filter-options-")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer os.RemoveAll(dir)

	keep := filepath.Join(dir, "keep.txt")
	skip := filepath.Join(dir, "skip.log")
	if err := os.WriteFile(keep, []byte("keep"), 0o644); err != nil {
		fmt.Println(err)
		return
	}
	if err := os.WriteFile(skip, []byte("skip"), 0o644); err != nil {
		fmt.Println(err)
		return
	}
	filter := func(path string, info os.FileInfo) bool {
		return info.IsDir() || filepath.Ext(path) == ".txt"
	}
	archivePath := filepath.Join(dir, "filtered.zip")
	if err := vzip.ZipFilesFilterWithOptions(archivePath, false, filter, []string{keep, skip}); err != nil {
		fmt.Println(err)
		return
	}
	names, _ := vzip.ListFileNames(archivePath, "")
	fmt.Println(names)
	// Output: [keep.txt]
}

func ExampleZipToWriter() {
	dir, err := os.MkdirTemp("", "knifer-go-vzip-writer-")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer os.RemoveAll(dir)

	source := filepath.Join(dir, "writer.txt")
	if err := os.WriteFile(source, []byte("writer"), 0o644); err != nil {
		fmt.Println(err)
		return
	}
	var archive bytes.Buffer
	if err := vzip.ZipToWriter(&archive, false, nil, source); err != nil {
		fmt.Println(err)
		return
	}
	reader, _ := archivezip.NewReader(bytes.NewReader(archive.Bytes()), int64(archive.Len()))
	fmt.Println(reader.File[0].Name)
	// Output: writer.txt
}

func ExampleZipToWriterWithOptions() {
	dir, err := os.MkdirTemp("", "knifer-go-vzip-writer-options-")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer os.RemoveAll(dir)

	source := filepath.Join(dir, "writer.txt")
	if err := os.WriteFile(source, []byte("writer"), 0o644); err != nil {
		fmt.Println(err)
		return
	}
	var archive bytes.Buffer
	err = vzip.ZipToWriterWithOptions(
		&archive,
		false,
		nil,
		[]string{source},
		vzip.WithCompressionLevel(flate.BestSpeed),
	)
	fmt.Println(err == nil)
	// Output: true
}

func ExampleZipToWriterUsingOptions() {
	dir, err := os.MkdirTemp("", "knifer-go-vzip-writer-using-options-")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer os.RemoveAll(dir)

	source := filepath.Join(dir, "writer.txt")
	if err := os.WriteFile(source, []byte("writer"), 0o644); err != nil {
		fmt.Println(err)
		return
	}
	var archive bytes.Buffer
	if err := vzip.ZipToWriterUsingOptions(&archive, []string{source}, vzip.WithSourceDir(false)); err != nil {
		fmt.Println(err)
		return
	}
	reader, _ := archivezip.NewReader(bytes.NewReader(archive.Bytes()), int64(archive.Len()))
	fmt.Println(reader.File[0].Name)
	// Output: writer.txt
}

func ExampleZipEntriesWithOptions() {
	dir, err := os.MkdirTemp("", "knifer-go-vzip-entries-options-")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer os.RemoveAll(dir)

	archivePath := filepath.Join(dir, "entries.zip")
	err = vzip.ZipEntriesWithOptions(
		archivePath,
		[]vzip.EntryData{{Name: "stored.txt", Data: []byte("stored")}},
		vzip.WithCompressionMethod(archivezip.Store),
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	reader, _ := vzip.Open(archivePath)
	defer reader.Close()
	fmt.Println(reader.File[0].Name, reader.File[0].Method == archivezip.Store)
	// Output: stored.txt true
}

func ExampleZipStreams() {
	dir, err := os.MkdirTemp("", "knifer-go-vzip-streams-")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer os.RemoveAll(dir)

	archivePath := filepath.Join(dir, "streams.zip")
	if err := vzip.ZipStreams(archivePath, vzip.StreamEntry{Name: "stream.txt", Reader: bytes.NewReader([]byte("stream"))}); err != nil {
		fmt.Println(err)
		return
	}
	data, _ := vzip.GetBytes(archivePath, "stream.txt")
	fmt.Println(string(data))
	// Output: stream
}

func ExampleZipStreamsWithOptions() {
	dir, err := os.MkdirTemp("", "knifer-go-vzip-streams-options-")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer os.RemoveAll(dir)

	archivePath := filepath.Join(dir, "streams.zip")
	err = vzip.ZipStreamsWithOptions(
		archivePath,
		[]vzip.StreamEntry{{Name: "stream.txt", Reader: bytes.NewReader([]byte("stream"))}},
		vzip.WithMaxBytes(16),
	)
	fmt.Println(err == nil)
	data, _ := vzip.GetBytes(archivePath, "stream.txt")
	fmt.Println(string(data))
	// Output:
	// true
	// stream
}

func ExampleZipStreamsToWriterWithOptions() {
	var archive bytes.Buffer
	err := vzip.ZipStreamsToWriterWithOptions(
		&archive,
		[]vzip.StreamEntry{{Name: "stream.txt", Reader: bytes.NewReader([]byte("stream"))}},
		vzip.WithCompressionLevel(flate.BestSpeed),
	)
	fmt.Println(err == nil)
	// Output: true
}

func ExampleUnzip() {
	archivePath, cleanup, err := writeExampleArchive(vzip.EntryData{Name: "hello.txt", Data: []byte("hello")})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer cleanup()

	dest, err := vzip.Unzip(archivePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	data, _ := os.ReadFile(filepath.Join(dest, "hello.txt"))
	fmt.Println(string(data))
	// Output: hello
}

func ExampleUnzipToLimit() {
	archivePath, cleanup, err := writeExampleArchive(vzip.EntryData{Name: "hello.txt", Data: []byte("hello")})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer cleanup()

	dest := filepath.Join(filepath.Dir(archivePath), "limit")
	err = vzip.UnzipToLimit(archivePath, dest, 32)
	fmt.Println(err == nil)
	data, _ := os.ReadFile(filepath.Join(dest, "hello.txt"))
	fmt.Println(string(data))
	// Output:
	// true
	// hello
}

func ExampleUnzipReaderToLimit() {
	var archive bytes.Buffer
	if err := vzip.ZipEntriesToWriter(&archive, vzip.EntryData{Name: "hello.txt", Data: []byte("hello")}); err != nil {
		fmt.Println(err)
		return
	}
	reader, _ := archivezip.NewReader(bytes.NewReader(archive.Bytes()), int64(archive.Len()))
	dir, err := os.MkdirTemp("", "knifer-go-vzip-reader-limit-")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer os.RemoveAll(dir)
	err = vzip.UnzipReaderToLimit(reader, dir, 32)
	fmt.Println(err == nil)
	// Output: true
}

func ExampleUnzipReaderToWithOptions() {
	var archive bytes.Buffer
	if err := vzip.ZipEntriesToWriter(&archive, vzip.EntryData{Name: "hello.txt", Data: []byte("hello")}); err != nil {
		fmt.Println(err)
		return
	}
	reader, _ := archivezip.NewReader(bytes.NewReader(archive.Bytes()), int64(archive.Len()))
	dir, err := os.MkdirTemp("", "knifer-go-vzip-reader-options-")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer os.RemoveAll(dir)
	err = vzip.UnzipReaderToWithOptions(reader, dir, vzip.WithMaxBytes(32))
	fmt.Println(err == nil)
	// Output: true
}

func ExampleUnzipToWithOptions() {
	archivePath, cleanup, err := writeExampleArchive(vzip.EntryData{Name: "hello.txt", Data: []byte("hello")})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer cleanup()

	dest := filepath.Join(filepath.Dir(archivePath), "options")
	err = vzip.UnzipToWithOptions(archivePath, dest, vzip.WithMaxBytes(32))
	fmt.Println(err == nil)
	// Output: true
}

func ExampleGetWithOptions() {
	archivePath, cleanup, err := writeExampleArchive(vzip.EntryData{Name: "data.txt", Data: []byte("payload")})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer cleanup()

	reader, err := vzip.GetWithOptions(archivePath, "data.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer reader.Close()
	data, _ := io.ReadAll(reader)
	fmt.Println(string(data))
	// Output: payload
}

func ExampleGetBytesWithOptions() {
	archivePath, cleanup, err := writeExampleArchive(vzip.EntryData{Name: "data.txt", Data: []byte("payload")})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer cleanup()

	data, err := vzip.GetBytesWithOptions(archivePath, "data.txt", vzip.WithMaxBytes(16))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(data))
	// Output: payload
}

func ExampleReadWithOptions() {
	archivePath, cleanup, err := writeExampleArchive(
		vzip.EntryData{Name: "b.txt", Data: []byte("b")},
		vzip.EntryData{Name: "a.txt", Data: []byte("a")},
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer cleanup()

	names := make([]string, 0)
	if err := vzip.ReadWithOptions(archivePath, func(file *archivezip.File) error {
		names = append(names, file.Name)
		return nil
	}); err != nil {
		fmt.Println(err)
		return
	}
	sort.Strings(names)
	fmt.Println(names)
	// Output: [a.txt b.txt]
}

func ExampleListFileNamesWithOptions() {
	archivePath, cleanup, err := writeExampleArchive(
		vzip.EntryData{Name: "config/app.yml", Data: []byte("app")},
		vzip.EntryData{Name: "config/db.yml", Data: []byte("db")},
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer cleanup()

	names, err := vzip.ListFileNamesWithOptions(archivePath, "config")
	if err != nil {
		fmt.Println(err)
		return
	}
	sort.Strings(names)
	fmt.Println(names)
	// Output: [app.yml db.yml]
}

func ExampleGzipWithOptions() {
	compressed, err := vzip.GzipWithOptions([]byte("hello"), vzip.WithCompressionLevel(flate.BestSpeed))
	if err != nil {
		fmt.Println(err)
		return
	}
	plain, _ := vzip.UnGzip(compressed)
	fmt.Println(string(plain))
	// Output: hello
}

func ExampleGunzip() {
	compressed, _ := vzip.Gzip([]byte("hello"))
	plain, err := vzip.Gunzip(compressed)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(plain))
	// Output: hello
}

func ExampleGzipFile() {
	dir, err := os.MkdirTemp("", "knifer-go-vzip-gzip-file-")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer os.RemoveAll(dir)
	source := filepath.Join(dir, "payload.txt")
	if err := os.WriteFile(source, []byte("payload"), 0o644); err != nil {
		fmt.Println(err)
		return
	}
	compressed, err := vzip.GzipFile(source)
	if err != nil {
		fmt.Println(err)
		return
	}
	plain, _ := vzip.UnGzip(compressed)
	fmt.Println(string(plain))
	// Output: payload
}

func ExampleGzipFileWithOptions() {
	dir, err := os.MkdirTemp("", "knifer-go-vzip-gzip-file-options-")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer os.RemoveAll(dir)
	source := filepath.Join(dir, "payload.txt")
	if err := os.WriteFile(source, []byte("payload"), 0o644); err != nil {
		fmt.Println(err)
		return
	}
	compressed, err := vzip.GzipFileWithOptions(source, vzip.WithMaxBytes(16))
	if err != nil {
		fmt.Println(err)
		return
	}
	plain, _ := vzip.UnGzip(compressed)
	fmt.Println(string(plain))
	// Output: payload
}

func ExampleGzipReader() {
	compressed, err := vzip.GzipReader(bytes.NewReader([]byte("hello")), 5)
	if err != nil {
		fmt.Println(err)
		return
	}
	plain, _ := vzip.UnGzip(compressed)
	fmt.Println(string(plain))
	// Output: hello
}

func ExampleGzipReaderWithOptions() {
	compressed, err := vzip.GzipReaderWithOptions(
		bytes.NewReader([]byte("hello")),
		5,
		vzip.WithCompressionLevel(flate.NoCompression),
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	plain, _ := vzip.UnGzip(compressed)
	fmt.Println(string(plain))
	// Output: hello
}

func ExampleUnGzipWithOptions() {
	compressed, _ := vzip.Gzip([]byte("hello"))
	plain, err := vzip.UnGzipWithOptions(compressed, vzip.WithMaxBytes(16))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(plain))
	// Output: hello
}

func ExampleUnGzipReader() {
	compressed, _ := vzip.Gzip([]byte("hello"))
	plain, err := vzip.UnGzipReader(bytes.NewReader(compressed), 5)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(plain))
	// Output: hello
}

func ExampleUnGzipReaderWithOptions() {
	compressed, _ := vzip.Gzip([]byte("hello"))
	plain, err := vzip.UnGzipReaderWithOptions(bytes.NewReader(compressed), 5, vzip.WithMaxBytes(16))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(plain))
	// Output: hello
}

func ExampleZlib() {
	compressed, err := vzip.Zlib([]byte("hello"))
	if err != nil {
		fmt.Println(err)
		return
	}
	plain, _ := vzip.UnZlib(compressed)
	fmt.Println(string(plain))
	// Output: hello
}

func ExampleZlibFile() {
	dir, err := os.MkdirTemp("", "knifer-go-vzip-zlib-file-")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer os.RemoveAll(dir)
	source := filepath.Join(dir, "payload.txt")
	if err := os.WriteFile(source, []byte("payload"), 0o644); err != nil {
		fmt.Println(err)
		return
	}
	compressed, err := vzip.ZlibFile(source, flate.BestSpeed)
	if err != nil {
		fmt.Println(err)
		return
	}
	plain, _ := vzip.UnZlib(compressed)
	fmt.Println(string(plain))
	// Output: payload
}

func ExampleZlibFileWithOptions() {
	dir, err := os.MkdirTemp("", "knifer-go-vzip-zlib-file-options-")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer os.RemoveAll(dir)
	source := filepath.Join(dir, "payload.txt")
	if err := os.WriteFile(source, []byte("payload"), 0o644); err != nil {
		fmt.Println(err)
		return
	}
	compressed, err := vzip.ZlibFileWithOptions(source, flate.BestSpeed, vzip.WithMaxBytes(16))
	if err != nil {
		fmt.Println(err)
		return
	}
	plain, _ := vzip.UnZlib(compressed)
	fmt.Println(string(plain))
	// Output: payload
}

func ExampleZlibLevel() {
	compressed, err := vzip.ZlibLevel([]byte("hello"), flate.BestSpeed)
	if err != nil {
		fmt.Println(err)
		return
	}
	plain, _ := vzip.UnZlib(compressed)
	fmt.Println(string(plain))
	// Output: hello
}

func ExampleZlibLevelWithOptions() {
	compressed, err := vzip.ZlibLevelWithOptions([]byte("hello"), flate.BestSpeed, vzip.WithMaxBytes(16))
	if err != nil {
		fmt.Println(err)
		return
	}
	plain, _ := vzip.UnZlib(compressed)
	fmt.Println(string(plain))
	// Output: hello
}

func ExampleZlibReader() {
	compressed, err := vzip.ZlibReader(bytes.NewReader([]byte("hello")), flate.BestSpeed, 5)
	if err != nil {
		fmt.Println(err)
		return
	}
	plain, _ := vzip.UnZlib(compressed)
	fmt.Println(string(plain))
	// Output: hello
}

func ExampleZlibReaderWithOptions() {
	compressed, err := vzip.ZlibReaderWithOptions(
		bytes.NewReader([]byte("hello")),
		flate.NoCompression,
		5,
		vzip.WithMaxBytes(16),
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	plain, _ := vzip.UnZlib(compressed)
	fmt.Println(string(plain))
	// Output: hello
}

func ExampleUnZlib() {
	compressed, _ := vzip.Zlib([]byte("hello"))
	plain, err := vzip.UnZlib(compressed)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(plain))
	// Output: hello
}

func ExampleUnZlibWithOptions() {
	compressed, _ := vzip.Zlib([]byte("hello"))
	plain, err := vzip.UnZlibWithOptions(compressed, vzip.WithMaxBytes(16))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(plain))
	// Output: hello
}

func ExampleUnzlib() {
	compressed, _ := vzip.Zlib([]byte("hello"))
	plain, err := vzip.Unzlib(compressed)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(plain))
	// Output: hello
}

func ExampleUnZlibReader() {
	compressed, _ := vzip.Zlib([]byte("hello"))
	plain, err := vzip.UnZlibReader(bytes.NewReader(compressed), 5)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(plain))
	// Output: hello
}

func ExampleUnZlibReaderWithOptions() {
	compressed, _ := vzip.Zlib([]byte("hello"))
	plain, err := vzip.UnZlibReaderWithOptions(bytes.NewReader(compressed), 5, vzip.WithMaxBytes(16))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(plain))
	// Output: hello
}

func ExampleReadFile() {
	dir, err := os.MkdirTemp("", "knifer-go-vzip-read-file-")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer os.RemoveAll(dir)
	source := filepath.Join(dir, "payload.txt")
	if err := os.WriteFile(source, []byte("payload"), 0o644); err != nil {
		fmt.Println(err)
		return
	}
	data, err := vzip.ReadFile(source)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(data))
	// Output: payload
}

func ExampleReadFileWithOptions() {
	data, err := vzip.ReadFileWithOptions("/virtual/payload.txt", vzip.WithReadFile(func(path string) ([]byte, error) {
		if path != "/virtual/payload.txt" {
			return nil, os.ErrNotExist
		}
		return []byte("virtual"), nil
	}))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(data))
	// Output: virtual
}
