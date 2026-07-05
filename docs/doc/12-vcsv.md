# vcsv Quickstart

`vcsv` provides CSV reading, writing, and structured conversion helpers, with support for readers, strings, maps, structs, BOM handling, delimiters, and CRLF options.

## Golden path APIs

The first-choice API set for this facade is kept in sync with `ai-context.json` and the generated tools catalog.

- `ForEach`
- `MapsToRecords`
- `Read`
- `ReadMaps`
- `ReadString`

## Which helper should I use?

| Goal | Start with | Notes |
| --- | --- | --- |
| Read all records from a reader | `Read` | Returns `[][]string`; best for small to medium CSV payloads. |
| Read CSV from an in-memory string | `ReadString` | Convenience wrapper for tests, examples, and configuration snippets. |
| Read rows keyed by header | `ReadMaps` / `ReadStringMaps` | Uses the first record as headers; validate header names before relying on map keys. |
| Process records one at a time | `ForEach` | Avoids building the full result slice and is better for large streams. |
| Write raw records | `Write` / `WriteString` | Use when headers and field ordering are already explicit. |
| Write maps with selected headers | `WriteMaps` / `WriteStringMaps` | Header order controls output order and missing map values become empty fields. |
| Convert records to maps | `RecordsToMaps` | Good after reading when downstream code expects named columns. |
| Convert maps to records | `MapsToRecords` | Use explicit headers to keep output deterministic. |
| Convert structs | `StructsToRecords`, `WriteStructs`, `WriteStringStructs` | Uses struct fields and `csv` tags for header names. |
| Tune parser behavior | `WithComma`, `WithFieldsPerRecord`, `WithTrimUTF8BOM`, `WithLazyQuotes` | Keep loose parsing options local to known legacy inputs. |
| Tune output compatibility | `WithUTF8BOM`, `WithUseCRLF` | Use for Excel or platform-specific consumers that require BOM or CRLF. |

## CSV correctness checklist

- Prefer `ForEach` for large files so the process does not allocate every record at once.
- Validate required headers after `ReadMaps`; duplicate or unexpected headers can make map-based processing ambiguous.
- Use explicit header slices with `WriteMaps` and `MapsToRecords` for deterministic column order.
- Enable `WithTrimUTF8BOM` for files that may come from Excel or Windows tools.
- Avoid `WithLazyQuotes` unless you knowingly accept malformed legacy CSV; strict parsing catches data-quality issues earlier.
- Keep `WithReuseRecord` away from handlers that retain record slices after the callback or loop iteration.
- Treat CSV formula strings from untrusted users carefully before exporting to spreadsheets.

## Read CSV as records and maps

```go
package main

import (
	"fmt"
	"strings"

	"github.com/imajinyun/knifer-go/vcsv"
)

func main() {
	records, err := vcsv.ReadString("name,age\nalice,30\n")
	if err != nil {
		panic(err)
	}
	fmt.Println(records[1][0])

	rows, err := vcsv.ReadMaps(strings.NewReader("name,age\nbob,20\n"))
	if err != nil {
		panic(err)
	}
	fmt.Println(rows[0]["name"])
}
```

## Use read options

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vcsv"
)

func main() {
	records, err := vcsv.ReadString("\ufeffname;age\n alice;30\n",
		vcsv.WithComma(';'),
		vcsv.WithTrimUTF8BOM(true),
		vcsv.WithTrimLeadingSpace(true),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(records[1][0])
}
```

## Write CSV strings

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vcsv"
)

func main() {
	out, err := vcsv.WriteString([][]string{{"name"}, {"alice"}}, vcsv.WithUseCRLF(true))
	if err != nil {
		panic(err)
	}
	fmt.Print(out)
}
```

## When not to use vcsv

- Use `encoding/csv` directly when a small call site only needs standard reader/writer behavior and no facade conversion helpers.
- Use a streaming ETL library when files are huge, need schema inference, or require backpressure across pipeline stages.
- Use a typed decoder with explicit validation when CSV input is an external contract with required fields and domain rules.
- Avoid map conversion when duplicate headers or strict column ordering are part of the data contract.

## Related packages

- Use `vpoi` when the workflow requires XLSX worksheets, workbook metadata, or in-memory spreadsheet export.
- Use `vfile` when CSV paths, temporary files, or directory traversal need filesystem policy checks.
- Use `vbean` when CSV records need to be mapped into typed structs after parsing.

## Benchmarks and trade-offs

- `Read` and `ReadString` are simple but allocate the full `[][]string`. `ForEach` trades convenience for lower peak memory.
- Map helpers improve readability but allocate a map per row and depend on stable headers.
- Struct conversion provides typed field names through reflection, which is convenient but slower than hand-written conversion in hot paths.
- BOM and CRLF options improve compatibility with spreadsheet tools while adding bytes or preprocessing work.
- `WithReuseRecord` can reduce allocations, but retained records must be copied by the caller.

## FAQ

### When should I use `ForEach`?

Use it for large CSV streams or imports where each record can be handled independently. It avoids retaining every row in memory.

### How is map column order decided when writing?

The `headers` argument defines output order. Values are read from each map by those header names, so pass a stable header slice for deterministic files.

### Why are my first header characters unexpected?

The file may include a UTF-8 BOM. Read with `WithTrimUTF8BOM(true)` or write with `WithUTF8BOM(true)` only for consumers that require it.

### Should I use lazy quotes?

Only for known legacy inputs that contain malformed quoting. For new data contracts, strict CSV parsing is safer.

## Convert between structs and CSV

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vcsv"
)

type Person struct {
	Name string `csv:"name"`
	Age  int    `csv:"age"`
}

func main() {
	records, err := vcsv.StructsToRecords([]Person{{Name: "alice", Age: 30}})
	if err != nil {
		panic(err)
	}
	fmt.Println(records)

	out, err := vcsv.WriteStringStructs([]Person{{Name: "bob", Age: 20}}, vcsv.WithUTF8BOM(true))
	if err != nil {
		panic(err)
	}
	fmt.Print(out)
}
```
