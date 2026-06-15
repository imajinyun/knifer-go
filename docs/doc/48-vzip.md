# vzip Quickstart

`vzip` provides ZIP archive creation, reading, extraction, in-memory/streaming entry writes, and gzip/zlib compression helpers, with overwrite, permission, and size-limit options.

## Create ZIP files from in-memory entries

```go
package main

import (
	"fmt"
	"path/filepath"

	"github.com/imajinyun/go-knifer/vzip"
)

func main() {
	zipFile := filepath.Join("/tmp", "docs.zip")
	err := vzip.ZipEntriesWithOptions(zipFile, []vzip.EntryData{
		{Name: "README.txt", Data: []byte("hello zip")},
	}, vzip.WithOverwrite(true))
	if err != nil {
		panic(err)
	}

	data, err := vzip.GetBytes(zipFile, "README.txt")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
}
```

## Create archives from files or directories

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/imajinyun/go-knifer/vzip"
)

func main() {
	dir, err := os.MkdirTemp("", "vzip-src-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	src := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(src, []byte("a"), 0o644); err != nil {
		panic(err)
	}

	zipFile := filepath.Join(dir, "out.zip")
	if err := vzip.ZipFilesWithOptions(zipFile, false, []string{src}, vzip.WithOverwrite(true)); err != nil {
		panic(err)
	}

	names, err := vzip.ListFileNames(zipFile, "")
	if err != nil {
		panic(err)
	}
	fmt.Println(names)
}
```

## Read and extract ZIP files

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/imajinyun/go-knifer/vzip"
)

func main() {
	dir, err := os.MkdirTemp("", "vzip-read-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	zipFile := filepath.Join(dir, "archive.zip")
	if err := vzip.ZipData(zipFile, "data.txt", "payload"); err != nil {
		panic(err)
	}

	if err := vzip.Read(zipFile, func(file *vzip.Entry) error {
		fmt.Println(file.Name)
		return nil
	}); err != nil {
		panic(err)
	}

	dest := filepath.Join(dir, "out")
	if err := vzip.UnzipToWithOptions(zipFile, dest, vzip.WithOverwrite(true), vzip.WithMaxBytes(1024)); err != nil {
		panic(err)
	}
}
```

## gzip and zlib compression

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vzip"
)

func main() {
	gz, err := vzip.GzipString("hello")
	if err != nil {
		panic(err)
	}
	plain, err := vzip.UnGzipString(gz)
	if err != nil {
		panic(err)
	}
	fmt.Println(plain)

	z, err := vzip.ZlibString("world", 1)
	if err != nil {
		panic(err)
	}
	unzip, err := vzip.UnZlibString(z)
	if err != nil {
		panic(err)
	}
	fmt.Println(unzip)
}
```
