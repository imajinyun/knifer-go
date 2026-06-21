package vzip_test

import (
	archivezip "archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/imajinyun/go-knifer/vzip"
)

func writeExampleArchive(entries ...vzip.EntryData) (string, func(), error) {
	dir, err := os.MkdirTemp("", "go-knifer-vzip-example-")
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
		vzip.EntryData{Name: "config/app.setting", Data: []byte("name=go-knifer")},
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
		vzip.EntryData{Name: "config/app.yml", Data: []byte("name: go-knifer")},
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
	dir, err := os.MkdirTemp("", "go-knifer-vzip-files-")
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
	dir, err := os.MkdirTemp("", "go-knifer-vzip-filter-")
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
		vzip.EntryData{Name: "config/app.yml", Data: []byte("name: go-knifer")},
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

	dir, err := os.MkdirTemp("", "go-knifer-vzip-unzip-reader-")
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

	dir, err := os.MkdirTemp("", "go-knifer-vzip-unzip-to-")
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
