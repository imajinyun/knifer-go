# vfile Quickstart

`vfile` provides helpers for file reading, writing, appending, directory creation, copying, deletion, filename parsing, magic-number file type detection, and bounded reads with default size protection.

Prefer temporary directories in tests. Keep user-controlled paths separate from trusted base directories. Use safe extraction or safe path helpers when dealing with archives or untrusted filenames. Do not ignore file I/O errors.

## Golden path APIs

The first-choice API set for this facade is kept in sync with `ai-context.json` and the generated tools catalog.

- `SafeJoin`
- `AppendFileString`
- `ExistsWithOptions`
- `CloseQuietly`
- `AppendFileStringWithOptions`

## Which helper should I use?

Choose the helper that makes the filesystem side effect explicit. Use temporary directories in tests and keep trusted base directories separate from user-controlled relative names.

| Need | Use | Notes |
| --- | --- | --- |
| Read or write a whole text file | `ReadFileString`, `WriteFileString` | Good for small configuration, fixtures, or generated text. Check the returned error. |
| Read lines or stream chunks | `ReadLines`, `ReadChunksWithOptions` | Prefer chunked reads when the input size is not tightly bounded. Set `WithBufferSize` deliberately. |
| Append to an existing log-like file | `AppendFileString` | Use when preserving existing content matters. Decide whether missing files should be created or rejected. |
| Create parent directories or touch a file | `Mkdir`, `Touch` | Set permissions explicitly with options when defaults are not appropriate. |
| Check file state before optional work | `Exists`, `IsFile`, `IsDirectory`, `Size` | Treat checks as hints, not synchronization; another process can change the path after the check. |
| Copy, move, or delete files | `CopyFile`, `Del` | Keep overwrite and deletion behavior visible at the call site. Do not ignore cleanup errors in production code. |
| Inspect names and extensions | `MainName`, `Extension` | These are string/path helpers; they do not validate whether the path is safe to open. |
| Detect a file family from bytes | `DetectFileType`, `DetectFileTypeBytes`, `DetectFileTypeFromPath` | Uses leading magic-number bytes to identify common images, archives, documents, audio/video, fonts, and executables. |

## Filesystem safety checklist

- Use `t.TempDir` or another temporary directory for tests and examples that write files.
- Keep untrusted path fragments relative to a trusted root; do not let callers provide arbitrary absolute paths for writes or deletes.
- Prefer bounded or chunked reads for content whose size is not controlled by your process.
- Check every returned error. A failed write, partial copy, or cleanup failure can leave stale data behind.
- Be explicit about overwrite and permission policy. Defaults are convenient, but reviewers should be able to see destructive behavior.
- Do not rely on `Exists` as an authorization or locking mechanism. It is useful for optional work, not for race-free decisions.
- Treat magic-number detection as identification only. Do not use it as the only upload, malware, authorization, or content-safety control.
- Do not trust file extensions alone. If routing depends on file family, inspect bytes and still enforce size, path, storage, and downstream parser limits.

## When not to use vfile

- Use `os`, `io`, or `fs` directly when you need platform-specific flags, file descriptors, memory mapping, or precise syscall behavior.
- Use `os.Root` on Go 1.24+ or a dedicated sandbox when untrusted names need a hardened filesystem boundary rather than convenience path helpers.
- Use archive-specific helpers such as `vzip` when the path is coming from an archive entry and extraction policy matters.
- Use streaming APIs instead of whole-file helpers for large, remote, or attacker-controlled content.
- Avoid mutating package-level or shared filesystem locations in reusable libraries; accept explicit paths or injected provider functions.
- Use a dedicated content scanning, sandboxing, or media validation service when accepting untrusted uploads in a security-sensitive workflow.

## Related packages

- Use `vzip` when filesystem work crosses into archive creation, extraction, or zip-entry path policy.
- Use `vcsv` or `vpoi` when files contain tabular data that needs CSV or XLSX parsing.
- Use `vurl` when file paths are derived from URLs or need URL-specific normalization first.
- Use `vimg` when detected image bytes need decoding, resizing, or transformation after the file family has been identified.

## Benchmarks and trade-offs

Use the focused file benchmarks to compare whole-file, chunked, copy, and option-provider paths on your machine:

```bash
go test -bench=. -benchmem -run=^$ ./internal/file ./vfile
```

Whole-file helpers are concise and easy to review, but they allocate enough memory for the content. Chunked reads and `CopyWithOptions` are better for large inputs and for call sites that need bounded memory use.

Magic-number detection reads only a bounded header. It is useful before dispatching to a parser, but it is not a substitute for parser-level validation, archive safety checks, size limits, or content scanning.

Provider options such as `WithOpen`, `WithOpenFile`, `WithStat`, `WithMkdirAll`, and `WithRemoveAll` make tests hermetic and reviewable. They add indirection, so keep production call sites simple unless injection is needed for policy, testing, or observability.

## FAQ

### Does vfile make arbitrary user paths safe?

No. `vfile` reduces common I/O boilerplate; callers still own trust boundaries. Join user-provided names under a trusted directory, reject path traversal where applicable, and use archive-safe helpers when extracting files.

### Should I use whole-file or chunked reads?

Use whole-file helpers for small, trusted inputs. Use `ReadChunksWithOptions` when input may be large, streamed, or controlled by another system.

### Why not ignore cleanup errors?

Cleanup errors can hide permission issues, stale files, or partial deletion. Tests may use best-effort cleanup, but production paths should decide whether cleanup failure is observable or fatal.

### Can magic-number detection prove a file is safe?

No. It only identifies known byte signatures. A file can have a valid signature and still be malicious, malformed, oversized, or unsafe for a downstream parser. Use it as one input to routing and validation, not as a security boundary.

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

### Detect file type from bytes

```go
ft := vfile.DetectFileTypeBytes([]byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A})
fmt.Println(ft.MIME, ft.Extension, vfile.IsImage(ft))

unknown := vfile.DetectFileTypeBytes([]byte("plain text"))
fmt.Println(unknown == vfile.UnknownFileType)
```

### Detect file type from a path

```go
dir, cleanup := exampleTempDir()
defer cleanup()

path := filepath.Join(dir, "archive.bin")
if err := os.WriteFile(path, []byte{'P', 'K', 0x03, 0x04, 0x00}, 0o600); err != nil {
	panic(err)
}
ft, err := vfile.DetectFileTypeFromPath(path)
if err != nil {
	panic(err)
}
fmt.Println(ft.MIME, vfile.IsArchive(ft))
```

## Read and write text files

```go
package main

import (
	"fmt"
	"path/filepath"

	"github.com/imajinyun/knifer-go/vfile"
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

	"github.com/imajinyun/knifer-go/vfile"
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

	"github.com/imajinyun/knifer-go/vfile"
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

	"github.com/imajinyun/knifer-go/vfile"
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

## Detect file type by magic number

```go
package main

import (
	"fmt"
	"strings"

	"github.com/imajinyun/knifer-go/vfile"
)

func main() {
	ft, err := vfile.DetectFileType(strings.NewReader("%PDF-1.7\n"))
	if err != nil {
		panic(err)
	}

	fmt.Println(ft.MIME)
	fmt.Println(ft.Extension)
	fmt.Println(vfile.IsDocument(ft))
}
```

## Classify bytes before routing

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vfile"
)

func main() {
	ft := vfile.DetectFileTypeBytes([]byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A})
	if vfile.IsImage(ft) {
		fmt.Println(ft.MIME)
	}

	unknown := vfile.DetectFileTypeBytes([]byte("plain text"))
	fmt.Println(unknown == vfile.UnknownFileType)
}
```
