# vpoi Quickstart

`vpoi` provides helpers for Excel XLSX worksheet listing, row reads/writes, multi-sheet writes, and in-memory buffer export.

## Write and read workbook files

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/imajinyun/go-knifer/vpoi"
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

	"github.com/imajinyun/go-knifer/vpoi"
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

## Read and write with an in-memory buffer

```go
package main

import (
	"bytes"
	"fmt"

	"github.com/imajinyun/go-knifer/vpoi"
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

	"github.com/imajinyun/go-knifer/vpoi"
)

func main() {
	path := filepath.Join(os.TempDir(), "multi.xlsx")
	sheets := map[string][][]string{
		"Users":  {{"id", "name"}, {"1", "alice"}},
		"Orders": {{"id", "total"}, {"100", "42"}},
	}
	if err := vpoi.WriteSheets(path, sheets, vpoi.WithOverwrite(true)); err != nil {
		panic(err)
	}
	names, err := vpoi.SheetNames(path)
	if err != nil {
		panic(err)
	}
	fmt.Println(names)
}
```
