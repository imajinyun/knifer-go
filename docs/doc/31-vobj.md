# vobj Quickstart

`vobj` provides object emptiness checks, comparisons, defaults, collection membership checks, type information, and serialization-based deep copy helpers.

## Check emptiness, length, and membership

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vobj"
)

func main() {
	fmt.Println(vobj.IsEmpty([]int{}))
	fmt.Println(vobj.Length(map[string]int{"go": 1}))
	fmt.Println(vobj.Contains([]string{"go", "knifer"}, "go"))
}
```

## Compare values and handle nil defaults

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vobj"
)

func main() {
	fmt.Println(vobj.Equal(1, int64(1)))
	fmt.Println(vobj.NotEqual("go", "knifer"))

	name := "go"
	fmt.Println(vobj.DefaultIfNil(&name, "fallback"))
	fmt.Println(vobj.DefaultIfNil[string](nil, "fallback"))
}
```

## Transform or consume non-nil pointers

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vobj"
)

func main() {
	name := "go"
	length := vobj.Apply(&name, func(s string) int { return len(s) })
	fmt.Println(length)

	vobj.Accept(&name, func(s string) {
		fmt.Println("hello", s)
	})
}
```

## Serialize, deserialize, and deep copy

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vobj"
)

type Profile struct {
	Name string
	Tags []string
}

func main() {
	src := Profile{Name: "alice", Tags: []string{"go"}}
	clone, err := vobj.Clone(src)
	if err != nil {
		panic(err)
	}
	clone.Tags[0] = "knifer"
	fmt.Println(src.Tags[0], clone.Tags[0])

	data, err := vobj.Serialize(src)
	if err != nil {
		panic(err)
	}
	decoded, err := vobj.DeserializeTo[Profile](data, Profile{})
	if err != nil {
		panic(err)
	}
	fmt.Println(decoded.Name)
}
```
