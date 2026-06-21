# vmap Quickstart

`vmap` provides generic map construction, lookup, conversion, filtering, aggregation, merge, set-operation, and comparison helpers.

## Which helper should I use?

| Need | Use | Notes |
| --- | --- | --- |
| Create an initialized map | `New`, `NewWithCap`, `OrEmpty` | Prefer `NewWithCap` when the expected size is known. `OrEmpty` turns nil input into a writable map. |
| Build a map from literal-like input | `OfE`, `FromPairs`, `FromEntries` | Prefer `OfE` over `Of` when invalid key/value pairs should be reported instead of dropped or panicking. |
| Read values with fallbacks | `Get`, `GetOr`, `GetAny`, `ContainsKey` | Use `GetOr` when the zero value is a valid value and absence needs a fallback. |
| Produce deterministic key order | `SortedKeys`, `SortedKeysFunc`, `SortedValues` | Plain `Keys`, `Values`, `Entries`, and iterator helpers follow Go map iteration order. |
| Transform or filter entries | `Map`, `MapValues`, `MapKeys`, `Filter`, `Reject`, `Partition` | These helpers allocate new maps and do not mutate the source map. |
| Stop on callback errors | `MapErr`, `MapKeysErr`, `MapValuesErr`, `FilterErr`, `ReduceErr` | Error order is still map-iteration dependent when more than one entry can fail. |
| Merge maps | `Merge`, `Assign`, `MergeFunc`, `Update` | `Merge`/`Assign` return a new map; `Update` mutates and returns the destination map. |
| Mutate an existing map | `Clear`, `MergeWithOverwrite`, `MergeWithoutOverwrite`, `Update` | Use only when callers intentionally share or own the destination map. |
| Compare or clone maps | `Equal`, `EqualFunc`, `Clone` | `Clone` is shallow: referenced values remain shared. |

## Map correctness checklist

- Map iteration order is intentionally unstable. Sort keys before producing logs, snapshots, examples, API responses, or tests that require deterministic output.
- Do not write to a nil map. Use `New`, `NewWithCap`, `OrEmpty`, `Clone`, or `Update(nil, src)` when a writable map is required.
- Treat `Clone`, `Merge`, `Filter`, and transform helpers as shallow copies. Pointer, slice, map, and interface values inside the map are still shared.
- Use mutation helpers only when ownership is clear. `Update`, `Clear`, `MergeWithOverwrite`, and `MergeWithoutOverwrite` change the destination map.
- Duplicate keys are resolved by each helper's documented rule. For example, `Merge` and `Assign` are last-write-wins, while `MergeFunc` lets callers provide conflict handling.
- Do not read and write the same map concurrently without external synchronization.

## When not to use vmap

- Use the standard `maps` package directly when a single primitive such as `maps.Clone`, `maps.Equal`, or `maps.Keys` is all you need.
- Use explicit loops when business logic has ordering, logging, metrics, or early-exit requirements that would be hidden inside callbacks.
- Use `vslice` helpers when the data model is ordered and duplicate preservation matters.

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

## Error-aware transforms

Use `MapErr`, `MapKeysErr`, `MapValuesErr`, `FilterErr`, and `ReduceErr` when a
map callback can fail. Map iteration order follows Go map semantics, so callers
must not rely on which entry fails first when several entries can return an
error.

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vmap"
)

func main() {
	labels, err := vmap.MapValuesErr(map[string]int{"a": 1}, func(key string, value int) (string, error) {
		return fmt.Sprintf("%s=%d", key, value), nil
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(labels["a"])
}
```

## Iterate with Go 1.23 range adapters

`Iter`, `IterKeys`, and `IterValues` expose Go iterator adapters for maps. Map
iteration order follows Go map semantics and is not stable; sort keys first when
callers need deterministic output.

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vmap"
)

func main() {
	m := map[string]int{"a": 1, "b": 2}

	for key, value := range vmap.Iter(m) {
		fmt.Println(key, value)
	}

	for key := range vmap.IterKeys(m) {
		fmt.Println(key)
	}

	for value := range vmap.IterValues(m) {
		fmt.Println(value)
	}
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

## Related packages

- Use `vslice` when order, duplicates, or indexed collection operations matter.
- Use `vset` when exact membership and uniqueness are the main requirement.
- Use `vbean` when map-shaped data needs to be bound into typed structs.

## Benchmarks and trade-offs

Run the focused map benchmark suite when changing collection-heavy code:

```bash
go test -bench=. -benchmem -run=^$ ./vmap
```

The suite covers filtering, error-aware transforms, sorted-key extraction, and merges across empty, small, medium, and large maps. Expect helpers that return new maps or sorted slices to allocate in proportion to input size; use mutation helpers only when avoiding allocations is worth the ownership constraints.

## FAQ

### Are `Keys`, `Values`, `Entries`, and `Iter` deterministic?

No. They follow Go map iteration order. Use `SortedKeys` or `SortedKeysFunc`, then index into the original map when output order matters.

### Does `Clone` deep-copy values?

No. `Clone` copies the map buckets but not referenced data inside values. If values contain pointers, slices, maps, or mutable structs, copy those values explicitly.

### Which merge helper should I choose?

Use `Merge` or `Assign` to create a new last-write-wins map. Use `MergeFunc` when duplicate keys need custom conflict resolution. Use `Update`, `MergeWithOverwrite`, or `MergeWithoutOverwrite` only when mutating a destination map is intentional.

### Why can error-aware helpers return different first errors?

Map iteration order is not stable. If several entries can fail, `MapErr`, `FilterErr`, and `ReduceErr` stop at the first error encountered in that particular iteration order.
