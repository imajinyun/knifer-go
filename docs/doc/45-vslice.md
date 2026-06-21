# vslice Quickstart

`vslice` provides generic slice helpers for emptiness checks, lookup, deduplication, mapping/filtering, set operations, pagination, and string joining.

## Which helper should I use?

| Need | Use | Notes |
| --- | --- | --- |
| Check or locate values | `IsEmpty`, `IsNotEmpty`, `Contains`, `IndexOf`, `LastIndexOf`, `Find`, `FindIndex` | Use `Find`/`FindIndex` when equality is not enough. |
| Transform values | `Map`, `FilterMap`, `FlatMap`, `Reduce`, `ForEach` | Prefer explicit loops when callbacks obscure side effects or complex branching. |
| Stop on callback errors | `MapErr`, `FilterErr`, `ReduceErr` | These preserve slice order and stop at the first failing element. |
| Deduplicate values | `Distinct`, `Uniq`, `UniqBy`, `Compact` | `Distinct` and `Uniq` preserve the first occurrence order. |
| Group or index records | `GroupBy`, `CountBy`, `KeyBy`, `Associate`, `SliceToMap` | Duplicate keys in keying helpers are resolved by later assignments in the produced map. |
| Work with windows | `Chunk`, `Window`, `Sliding`, `Sub`, `Page` | Invalid sizes, ranges, or pages return empty slices instead of panicking. |
| Pair values | `Zip2`, `Unzip2` | `Zip2` truncates to the shorter input length. |
| Set-like operations | `Union`, `Intersection`, `Subtract` | These require comparable element types and deduplicate results. |
| Iterate with Go range adapters | `Iter`, `IterIndexed` | Slice iteration is stable and follows index order. |

## Slice correctness checklist

- `Reverse` mutates the input slice in place. Clone first when callers still need the original order.
- Result slices may contain the same element values as the input. If elements are pointers, maps, slices, or structs with mutable fields, copy the elements explicitly before mutating them.
- `Sub`, `Chunk`, `Window`, and `Sliding` create slices that can share backing storage with the input. Treat returned windows as views unless the implementation contract for your use case requires explicit cloning.
- `Distinct`, `Uniq`, `Union`, `Intersection`, and `Subtract` use maps internally and require comparable keys or derived keys.
- `Page` uses one-based page numbers. Invalid page numbers, non-positive page sizes, and out-of-range pages return an empty slice.
- Do not mutate the same slice concurrently from multiple goroutines without synchronization.

## When not to use vslice

- Use the standard `slices` package directly for simple primitives such as `slices.Sort`, `slices.Clone`, `slices.DeleteFunc`, or `slices.Contains` when no go-knifer-specific helper is needed.
- Use explicit loops when you need precise allocation control, early returns with custom cleanup, or clear side-effect sequencing.
- Use `vset` when membership is the primary operation and ordering or duplicate counts are not meaningful.

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

## Related packages

- Use `vset` when uniqueness and membership are more important than order or duplicate preservation.
- Use `vmap` when slice elements should be grouped, indexed, or transformed into keyed data.
- Use `vjob` when slice processing needs batching, sharding, or merge callbacks.

## Benchmarks and trade-offs

Run the focused slice benchmark suite when changing collection-heavy code:

```bash
go test -bench=. -benchmem -run=^$ ./vslice
```

The suite covers filtering, mapping, error-aware mapping, windows, zipping, and deduplication across empty, small, medium, and large slices. Helpers that create transformed slices allocate proportional to the result size; set-like helpers also allocate maps to track membership.

## FAQ

### Does `Reverse` return a new slice?

No. `Reverse` reverses the input slice in place and returns the same slice. Clone the input first when mutation would surprise callers.

### Are window and chunk results independent copies?

Treat them as views over the input unless you explicitly clone before storing or mutating them. This avoids accidental backing-array aliasing bugs.

### What happens when `Zip2` receives different lengths?

`Zip2` pairs elements up to the shorter input length and ignores the remaining tail from the longer input.

### Should I use `vslice.Union` or `vset`?

Use `vslice.Union` when you need a slice result and first-seen order matters. Use `vset` when repeated membership checks and set algebra are the main operations.
