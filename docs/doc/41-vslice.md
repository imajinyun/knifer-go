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
