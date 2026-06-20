# vset Quickstart

`vset` provides generic sets and common numeric/string set aliases, with support for add/remove/contains, set operations, member export, and JSON/YAML encoding.

## Create sets and add, remove, or check members

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vset"
)

func main() {
	s := vset.NewString("go", "knifer")
	s.Add("tool")
	fmt.Println(s.Contains("tool"))
	s.Remove("go")
	fmt.Println(s.Members())
}
```

## Use generic sets and set operations

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vset"
)

func main() {
	a := vset.New(1, 2, 3)
	b := vset.New(3, 4)

	fmt.Println(a.Union(b).Members())
	fmt.Println(a.Intersect(b).Members())
	fmt.Println(a.Sub(b).Members())
	fmt.Println(a.Equal(vset.New(1, 2, 3)))
}
```

## Use numeric set aliases

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vset"
)

func main() {
	ints := vset.NewInt(1, 2).Union(vset.NewInt(2, 3))
	fmt.Println(ints.Equal(vset.NewInt(1, 2, 3)))

	uints := vset.NewUint64(10, 20).Sub(vset.NewUint64(10))
	fmt.Println(uints.Members())
}
```

## Encode and decode sets as JSON

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/imajinyun/go-knifer/vset"
)

func main() {
	original := vset.NewString("go", "knifer")
	b, err := json.Marshal(original)
	if err != nil {
		panic(err)
	}

	var decoded vset.String
	if err := json.Unmarshal(b, &decoded); err != nil {
		panic(err)
	}
	fmt.Println(decoded.Equal(original))
}
```
