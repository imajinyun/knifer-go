# vslice Quickstart

`vslice` provides generic slice helpers for emptiness checks, lookup, deduplication, mapping/filtering, set operations, pagination, and string joining.

## Check and find elements

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vslice"
)

func main() {
	items := []string{"go", "rust", "go"}

	fmt.Println(vslice.IsNotEmpty(items))
	fmt.Println(vslice.Contains(items, "go"))
	fmt.Println(vslice.IndexOf(items, "go"))
	fmt.Println(vslice.LastIndexOf(items, "go"))
}
```

## Deduplicate, reverse, and join

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vslice"
)

func main() {
	nums := []int{1, 2, 2, 3}

	fmt.Println(vslice.Distinct(nums))
	fmt.Println(vslice.Reverse(nums))
	fmt.Println(vslice.Join([]string{"a", "b", "c"}, ","))
}
```

## Map and filter

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vslice"
)

func main() {
	nums := []int{1, 2, 3, 4}

	doubled := vslice.Map(nums, func(n int) int { return n * 2 })
	even := vslice.Filter(nums, func(n int) bool { return n%2 == 0 })

	fmt.Println(doubled)
	fmt.Println(even)
}
```

## Error-aware transforms and windows

Use `MapErr`, `FilterErr`, and `ReduceErr` when a callback can fail and the
caller should stop on the first error. Use `Window` for overlapping windows,
`Sliding` for stepped windows, and `Zip2` / `Unzip2` for typed pairing without
reflection.

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vslice"
)

func main() {
	lengths, err := vslice.MapErr([]string{"go", "knifer"}, func(s string) (int, error) {
		return len(s), nil
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(lengths)
	fmt.Println(vslice.Window([]int{1, 2, 3, 4}, 3))
	fmt.Println(vslice.Sliding([]int{1, 2, 3, 4, 5}, 2, 2))
	fmt.Println(vslice.Zip2([]string{"a", "b"}, []int{1, 2, 3}))
}
```

## Iterate with Go 1.23 range adapters

`Iter` yields values in slice index order. `IterIndexed` yields index-value pairs
in the same stable order.

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vslice"
)

func main() {
	items := []string{"go", "knifer"}

	for value := range vslice.Iter(items) {
		fmt.Println(value)
	}

	for index, value := range vslice.IterIndexed(items) {
		fmt.Println(index, value)
	}
}
```

## Set operations and pagination

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vslice"
)

func main() {
	a := []int{1, 2, 3}
	b := []int{3, 4}

	fmt.Println(vslice.Union(a, b))
	fmt.Println(vslice.Intersection(a, b))
	fmt.Println(vslice.Subtract(a, b))
	fmt.Println(vslice.Page([]string{"a", "b", "c", "d"}, 2, 2))
}
```
