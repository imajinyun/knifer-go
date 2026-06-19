package vfile_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/imajinyun/go-knifer/vfile"
)

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
	dir, err := os.MkdirTemp("", "go-knifer-vfile-example-*")
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
	dir, err := os.MkdirTemp("", "go-knifer-vfile-example-*")
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
