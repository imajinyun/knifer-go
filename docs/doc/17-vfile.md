# vfile Quickstart

`vfile` provides helpers for file reading, writing, appending, directory creation, copying, deletion, filename parsing, and bounded reads with default size protection.

## Read and write text files

```go
package main

import (
	"fmt"
	"path/filepath"

	"github.com/imajinyun/go-knifer/vfile"
)

func main() {
	path := filepath.Join("tmp", "hello.txt")
	if err := vfile.WriteFileString(path, "hello\nworld\n"); err != nil {
		panic(err)
	}

	text, err := vfile.ReadFileString(path)
	if err != nil {
		panic(err)
	}
	fmt.Print(text)
}
```

## Read by line and by chunk

```go
package main

import (
	"fmt"
	"strings"

	"github.com/imajinyun/go-knifer/vfile"
)

func main() {
	lines, err := vfile.ReadLines(strings.NewReader("a\nb\n"))
	if err != nil {
		panic(err)
	}
	fmt.Println(lines)

	err = vfile.ReadChunksWithOptions(
		strings.NewReader("abcdef"),
		func(chunk []byte) error {
			fmt.Println(string(chunk))
			return nil
		},
		vfile.WithBufferSize(3),
	)
	if err != nil {
		panic(err)
	}
}
```

## Create directories, append, and touch files

```go
package main

import (
	"path/filepath"

	"github.com/imajinyun/go-knifer/vfile"
)

func main() {
	dir := filepath.Join("tmp", "logs")
	if err := vfile.Mkdir(dir, vfile.WithMkdirPerm(0o755)); err != nil {
		panic(err)
	}

	path := filepath.Join(dir, "app.log")
	if err := vfile.Touch(path); err != nil {
		panic(err)
	}
	if err := vfile.AppendFileString(path, "started\n"); err != nil {
		panic(err)
	}
}
```

## Check, copy, and delete files

```go
package main

import (
	"fmt"
	"path/filepath"

	"github.com/imajinyun/go-knifer/vfile"
)

func main() {
	src := filepath.Join("tmp", "src.txt")
	dst := filepath.Join("tmp", "backup", "src.txt")
	_ = vfile.WriteFileString(src, "content")

	if err := vfile.CopyFile(src, dst, vfile.WithOverwrite(true)); err != nil {
		panic(err)
	}
	fmt.Println(vfile.Exists(dst), vfile.IsFile(dst), vfile.Size(dst))
	fmt.Println(vfile.MainName(dst), vfile.Extension(dst))

	_ = vfile.Del(filepath.Join("tmp", "backup"))
}
```
