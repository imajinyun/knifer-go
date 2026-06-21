# vfile Quickstart

`vfile` provides helpers for file reading, writing, appending, directory creation, copying, deletion, filename parsing, and bounded reads with default size protection.

Prefer temporary directories in tests. Keep user-controlled paths separate from trusted base directories. Use safe extraction or safe path helpers when dealing with archives or untrusted filenames. Do not ignore file I/O errors.

## Cookbook

### Read and write a text file

```go
dir, cleanup := exampleTempDir()
defer cleanup()

path := filepath.Join(dir, "note.txt")
if err := vfile.WriteFileString(path, "hello"); err != nil {
	panic(err)
}
content, err := vfile.ReadFileString(path)
if err != nil {
	panic(err)
}
fmt.Println(content)
```

### Append without clobbering existing content

```go
dir, cleanup := exampleTempDir()
defer cleanup()

path := filepath.Join(dir, "app.log")
if err := vfile.WriteFileString(path, "start"); err != nil {
	panic(err)
}
if err := vfile.AppendFileString(path, "\nstop"); err != nil {
	panic(err)
}
content, err := vfile.ReadFileString(path)
if err != nil {
	panic(err)
}
fmt.Println(content)
```

### Copy or move files

```go
dir, cleanup := exampleTempDir()
defer cleanup()

src := filepath.Join(dir, "src.txt")
backup := filepath.Join(dir, "backup", "src.txt")
moved := filepath.Join(dir, "moved.txt")
if err := vfile.WriteFileString(src, "payload"); err != nil {
	panic(err)
}
if err := vfile.CopyFile(src, backup); err != nil {
	panic(err)
}
if err := vfile.CopyFile(src, moved); err != nil {
	panic(err)
}
if err := vfile.Del(src); err != nil {
	panic(err)
}
fmt.Println(vfile.Exists(backup), vfile.Exists(src), vfile.Exists(moved))
```

### Check existence before optional work

```go
dir, cleanup := exampleTempDir()
defer cleanup()

path := filepath.Join(dir, "optional.txt")
if !vfile.Exists(path) {
	if err := vfile.WriteFileString(path, "created"); err != nil {
		panic(err)
	}
}
fmt.Println(vfile.Exists(path))
```

### Use temporary directories in tests

```go
dir, cleanup := exampleTempDir()
defer cleanup()

path := filepath.Join(dir, "nested", "note.txt")
if err := vfile.WriteFileString(path, "isolated"); err != nil {
	panic(err)
}
fmt.Println(vfile.IsFile(path))
```

### Handle explicit file errors

```go
dir, cleanup := exampleTempDir()
defer cleanup()

missing := filepath.Join(dir, "missing.txt")
if _, err := vfile.ReadFileString(missing); err != nil {
	// Decide whether this is expected optional work or a hard failure.
	panic(err)
}
```

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
