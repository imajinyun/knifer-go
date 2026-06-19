package vzip_test

import (
	archivezip "archive/zip"
	"bytes"
	"fmt"
	"io"
	"sort"

	"github.com/imajinyun/go-knifer/vzip"
)

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
