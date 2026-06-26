# vpoi Quickstart

`vpoi` provides helpers for Excel XLSX worksheet listing, row reads/writes, multi-sheet writes, and in-memory buffer export.

Worksheet names are validated before opening or saving workbooks. Use `vpoi.ValidateSheetName` or `vpoi.IsValidSheetName` when sheet names come from user input, and rely on deterministic alphabetical sheet ordering when `WriteSheets` materializes multi-sheet workbooks.

## Which helper should I use?

Choose helpers by where the workbook lives and whether you need one sheet, multiple sheets, or early worksheet validation.

| Need | Use | Notes |
| --- | --- | --- |
| Write a simple workbook file | `WriteRows` | Good for single-sheet exports with default worksheet behavior. |
| Read rows from a workbook file | `ReadRows` | Use when the file path is already trusted and available on disk. |
| Write to a named worksheet or start cell | `WriteSheetRows`, `WithStartCell` | Validate user-supplied worksheet names before writing. |
| Read a bounded region | `WithReadStartCell`, `WithReadLimit` | Use for previews, imports with known header offsets, or defensive partial reads. |
| Preserve typed values on write | `WriteAnyRows`, `ReadCells` | Use when numbers, booleans, dates, or nil cells should not be forced through string writes. |
| List workbook worksheets | `SheetNames` | Use to inspect incoming files before selecting a sheet. |
| Validate worksheet names | `ValidateSheetName`, `IsValidSheetName` | Run before using names from users, reports, or external systems. |
| Work entirely in memory | `WriteRowsToBuffer`, `ReadRowsFromReader` | Useful for HTTP responses, tests, and pipelines that should avoid temporary files. |
| Write several worksheets | `WriteSheets` | Sheet materialization is deterministic; keep map keys as the intended worksheet names. |

## XLSX safety checklist

- Validate worksheet names from user input before opening or saving workbooks.
- Use temporary directories or in-memory buffers in tests to avoid leaving workbook files behind.
- Be explicit about overwrite behavior when writing files; accidental replacement is hard to detect after export.
- Treat uploaded workbooks as untrusted input. Validate sheet names and row shape before binding rows into application structs.
- Avoid logging raw spreadsheet cells when they may contain personal data, credentials, or customer content.
- Prefer in-memory buffer helpers for HTTP export paths where no durable file is needed.

## Related packages

- Use `vcsv` when tabular data should remain plain text without XLSX workbook metadata.
- Use `vfile` when workbook paths, temporary files, or directory traversal need filesystem policy checks.
- Use `vbean` when parsed rows need to be bound into typed records after validation.

## When not to use vpoi

- Use the underlying Excel library directly when you need formulas, charts, styles, merged cells, data validation, images, comments, or advanced workbook features.
- Use CSV helpers when the workflow is plain tabular text and does not require XLSX worksheets or workbook metadata.
- Use a streaming spreadsheet writer for very large exports that cannot fit comfortably in memory.
- Use domain-specific import validation before trusting uploaded workbook rows or binding them into application records.
- Avoid file helpers when an in-memory buffer is safer for short-lived HTTP responses or tests.

## Benchmarks and trade-offs

Use the POI benchmark suite to measure workbook export overhead on your machine:

```bash
go test -bench=. -benchmem -run=^$ ./vpoi
```

The suite covers representative in-memory workbook writes. Treat the output as a local baseline rather than a universal performance claim. For large exports, benchmark with row counts, sheet counts, and cell sizes that match your workload.

## FAQ

### Does vpoi handle every Excel feature?

No. `vpoi` focuses on common XLSX row read/write workflows. Use the underlying Excel library or a dedicated spreadsheet package directly when you need formulas, charts, styles, merged cells, or advanced workbook features.

### Should exports use files or buffers?

Use files when another process needs a durable artifact. Use buffers for HTTP responses, tests, and short-lived pipelines where keeping data in memory is simpler and safe for the expected size.

### Why validate sheet names early?

Early validation turns user-facing report or worksheet naming problems into clear input errors before a workbook is partially written.

## Write and read workbook files

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/imajinyun/knifer-go/vpoi"
)

func main() {
	path := filepath.Join(os.TempDir(), "users.xlsx")
	rows := [][]string{{"id", "name"}, {"1", "alice"}}

	if err := vpoi.WriteRows(path, rows, vpoi.WithOverwrite(true)); err != nil {
		panic(err)
	}
	got, err := vpoi.ReadRows(path)
	if err != nil {
		panic(err)
	}
	fmt.Println(got[1][1])
}
```

## Specify worksheets and start cells

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/imajinyun/knifer-go/vpoi"
)

func main() {
	path := filepath.Join(os.TempDir(), "sheet-users.xlsx")
	rows := [][]string{{"id", "name"}, {"1", "alice"}}

	if err := vpoi.WriteSheetRows(path, "Users", rows,
		vpoi.WithStartCell(2, 2),
		vpoi.WithOverwrite(true),
	); err != nil {
		panic(err)
	}
	names, err := vpoi.SheetNames(path)
	if err != nil {
		panic(err)
	}
	fmt.Println(names)
}
```

## Read a bounded region

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/imajinyun/knifer-go/vpoi"
)

func main() {
	path := filepath.Join(os.TempDir(), "region-users.xlsx")
	rows := [][]string{
		{"id", "name", "score"},
		{"1", "alice", "100"},
		{"2", "bob", "98"},
	}
	if err := vpoi.WriteRows(path, rows, vpoi.WithOverwrite(true)); err != nil {
		panic(err)
	}
	got, err := vpoi.ReadRows(path,
		vpoi.WithReadStartCell(2, 2),
		vpoi.WithReadLimit(2, 2),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(got)
}
```

## Write typed values and inspect cell metadata

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/imajinyun/knifer-go/vpoi"
)

func main() {
	path := filepath.Join(os.TempDir(), "typed-users.xlsx")
	rows := [][]any{
		{"name", "score", "active"},
		{"alice", 100, true},
	}
	if err := vpoi.WriteAnyRows(path, rows, vpoi.WithOverwrite(true)); err != nil {
		panic(err)
	}
	cells, err := vpoi.ReadCells(path,
		vpoi.WithReadStartCell(2, 2),
		vpoi.WithReadLimit(1, 2),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(cells[0][0].Value)
	fmt.Println(cells[0][0].Type == vpoi.CellTypeNumber)
}
```

## Validate worksheet names early

```go
package main

import (
	"errors"
	"fmt"

	"github.com/imajinyun/knifer-go/vpoi"
)

func main() {
	err := vpoi.ValidateSheetName("bad/name")
	fmt.Println(errors.Is(err, vpoi.ErrInvalidSheetName))
	fmt.Println(vpoi.IsValidSheetName("Reports"))
}
```

## Read and write with an in-memory buffer

```go
package main

import (
	"bytes"
	"fmt"

	"github.com/imajinyun/knifer-go/vpoi"
)

func main() {
	rows := [][]string{{"id", "name"}, {"1", "alice"}}
	buf, err := vpoi.WriteRowsToBuffer("Users", rows)
	if err != nil {
		panic(err)
	}
	got, err := vpoi.ReadRowsFromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		panic(err)
	}
	fmt.Println(got[1][1])
}
```

## Write multiple worksheets at once

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/imajinyun/knifer-go/vpoi"
)

func main() {
	path := filepath.Join(os.TempDir(), "multi.xlsx")
	sheets := map[string][][]string{
		"Users":  {{"id", "name"}, {"1", "alice"}},
		"Orders": {{"id", "total"}, {"100", "42"}},
		"Audit":  {{"event"}, {"created"}},
	}
	if err := vpoi.WriteSheets(path, sheets, vpoi.WithOverwrite(true)); err != nil {
		panic(err)
	}
	names, err := vpoi.SheetNames(path)
	if err != nil {
		panic(err)
	}
	fmt.Println(names) // alphabetical order: [Audit Orders Users]
}
```
