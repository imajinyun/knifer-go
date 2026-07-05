# vzip Quickstart

`vzip` provides ZIP archive creation, reading, extraction, in-memory/streaming entry writes, and gzip/zlib compression helpers, with overwrite, permission, and size-limit options.

## When to use vzip

| Scenario | Use `vzip` when | Prefer another tool when |
| --- | --- | --- |
| Build a ZIP archive in application code | You need a small helper around common ZIP entry creation, file packing, extraction, or content reads. | You need a fully custom archive pipeline with low-level `archive/zip` control over headers, methods, or streaming internals. |
| Extract archives from another system | You want explicit overwrite, permission, and size-limit options at the call site. | The input is untrusted and you still need deeper validation or quarantine logic beyond path and size controls. |
| Compress or decompress string/byte payloads | You need convenience wrappers for gzip/zlib round trips in tests, fixtures, or transport helpers. | You need streaming compression control, custom dictionaries, or direct interoperability tuning. |

## Golden path APIs

The first-choice API set for this facade is kept in sync with `ai-context.json` and the generated tools catalog.

- `Append`
- `NewWriter`
- `AppendWithOptions`
- `Get`
- `GetBytes`

## Which helper should I use?

Choose the helper that matches the archive operation you are performing: create, inspect, extract, or compress payloads.

| Need | Use | Notes |
| --- | --- | --- |
| Create a ZIP from in-memory files | `ZipEntriesWithOptions`, `ZipData` | Good for generated reports, fixtures, and small synthetic archives. |
| Create a ZIP from existing files or directories | `ZipFilesWithOptions` | Keep overwrite behavior explicit with options. |
| List entry names or inspect entries | `ListFileNames`, `Read` | Use `Read` when you need to inspect each entry before deciding whether to extract or decode it. |
| Read one entry as bytes | `GetBytes` | Prefer for small known entries; avoid whole-entry reads for uncontrolled large content. |
| Extract an archive to disk | `UnzipToWithOptions` | Set `WithOverwrite`, `WithMaxBytes`, and related options deliberately for safety. |
| gzip / ungzip short payloads | `GzipString`, `UnGzipString`, `Gzip`, `UnGzip` | Good for fixtures and small transport payload helpers. |
| zlib / unzlib short payloads | `ZlibString`, `UnZlibString`, `Zlib`, `UnZlib` | Choose when the downstream protocol expects zlib rather than gzip framing. |

## Archive safety checklist

- Treat archive extraction as a trust boundary. Keep destination directories explicit and isolated from unrelated application data.
- Use size-limiting options such as `WithMaxBytes` when archive contents come from another system or user input.
- Be explicit about overwrite behavior during extraction and archive creation; silent replacement is easy to miss in review.
- Inspect entries with `Read` or `ListFileNames` before extraction when you need policy checks or allow-lists.
- Prefer temporary directories in tests and intermediate workflows so extraction side effects are easy to clean up.
- Use whole-entry helpers like `GetBytes` only when entry size is already bounded or trusted.

## When not to use vzip

- Use `archive/zip` directly when you need custom headers, per-entry metadata, streaming internals, or nonstandard ZIP behavior that the facade does not expose.
- Use a dedicated archive scanner or quarantine pipeline for hostile uploads that require malware scanning, content-type enforcement, or policy decisions before extraction.
- Use `compress/gzip` or `compress/zlib` directly for long-running streaming compression with backpressure instead of collecting the compressed payload in memory.
- Avoid extracting archives into application-owned directories that also contain unrelated state; use a temporary or dedicated destination first.
- Avoid whole-entry reads for unbounded entries; stream with `Get` or inspect metadata before loading content into memory.

## Related packages

- Use `vfile` for ordinary filesystem path, temp-file, directory, and locking helpers outside archive workflows.
- Use `vcsv`, `vpoi`, or `vxml` when archive entries contain structured tabular or XML data that must be parsed after extraction.
- Use `vcrypto` when archive workflows need separate hashing, encryption, or integrity verification policy.

## Benchmarks and trade-offs

Use package benchmarks and examples as local baselines for archive convenience overhead:

```bash
go test -bench=. -benchmem -run=^$ ./internal/zip ./vzip
```

ZIP helpers trade low-level control for repeatable defaults: call sites can show overwrite, size, permission, compression, and provider policies without repeating archive plumbing. That readability matters most for common create, inspect, and extract flows.

Safety options are not free. Entry walking, path checks, size accounting, compression level choices, and whole-entry reads affect CPU, allocations, and I/O. Measure with representative archive sizes instead of assuming one compression level or helper is universally fastest.

## FAQ

### Does vzip make archive extraction completely safe?

No. `vzip` helps make extraction policy visible with options, but callers still own destination selection, trust-boundary checks, and any additional filtering required by the application.

### Should I read entries into memory or extract to disk?

Read into memory for small, known entries that you immediately inspect or parse. Extract to disk when files are larger, need to be consumed by other tools, or should remain on disk after the operation.

### When should I use gzip/zlib helpers instead of ZIP helpers?

Use gzip/zlib helpers for single payload compression. Use ZIP helpers when you need multiple named entries, archive metadata, or extraction to a directory.

## Create ZIP files from in-memory entries

```go
package main

import (
	"fmt"
	"path/filepath"

	"github.com/imajinyun/knifer-go/vzip"
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
	archivezip "archive/zip"
	"fmt"
	"os"
	"path/filepath"

	"github.com/imajinyun/knifer-go/vzip"
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

	"github.com/imajinyun/knifer-go/vzip"
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

	if err := vzip.Read(zipFile, func(file *archivezip.File) error {
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

	"github.com/imajinyun/knifer-go/vzip"
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
