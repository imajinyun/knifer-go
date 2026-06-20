# vcsv Quickstart

`vcsv` provides CSV reading, writing, and structured conversion helpers, with support for readers, strings, maps, structs, BOM handling, delimiters, and CRLF options.

## Read CSV as records and maps

```go
package main

import (
	"fmt"
	"strings"

	"github.com/imajinyun/go-knifer/vcsv"
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

	"github.com/imajinyun/go-knifer/vcsv"
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

	"github.com/imajinyun/go-knifer/vcsv"
)

func main() {
	out, err := vcsv.WriteString([][]string{{"name"}, {"alice"}}, vcsv.WithUseCRLF(true))
	if err != nil {
		panic(err)
	}
	fmt.Print(out)
}
```

## Convert between structs and CSV

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vcsv"
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
