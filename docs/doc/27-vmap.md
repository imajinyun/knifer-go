# vmap Quickstart

`vmap` provides generic map construction, lookup, conversion, filtering, aggregation, merge, set-operation, and comparison helpers.

## Build and query maps

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vmap"
)

func main() {
	m, err := vmap.OfE[string, int]("a", 1, "b", 2)
	if err != nil {
		panic(err)
	}

	fmt.Println(vmap.ContainsKey(m, "a"))
	fmt.Println(vmap.GetOr(m, "missing", 99))
	fmt.Println(vmap.SortedKeys(m))
}
```

## Build from pairs and transform keys or values

```go
package main

import (
	"fmt"
	"strconv"

	"github.com/imajinyun/go-knifer/vmap"
)

func main() {
	m := vmap.FromPairs(
		vmap.Pair[string, int]{Key: "a", Value: 1},
		vmap.Pair[string, int]{Key: "b", Value: 2},
	)

	labels := vmap.MapValues(m, func(k string, v int) string {
		return k + ":" + strconv.Itoa(v)
	})
	fmt.Println(labels)
}
```

## Filter, group, and aggregate

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vmap"
)

func main() {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	odds := vmap.Filter(m, func(_ string, v int) bool { return v%2 == 1 })
	total := vmap.Reduce(m, 0, func(acc int, _ string, v int) int { return acc + v })
	byLen := vmap.GroupBy([]string{"go", "js", "java"}, func(s string) int { return len(s) })

	fmt.Println(odds)
	fmt.Println(total)
	fmt.Println(byLen[2])
}
```

## Merge and run set operations

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vmap"
)

func main() {
	a := map[string]int{"a": 1, "b": 2}
	b := map[string]int{"b": 20, "c": 3}

	fmt.Println(vmap.Merge(a, b))
	fmt.Println(vmap.Intersect(a, b))
	fmt.Println(vmap.Diff(a, b))
	fmt.Println(vmap.SymmetricDiff(a, b))
}
```
