# vset Quickstart

`vset` provides generic sets and common numeric/string set aliases, with support for add/remove/contains, set operations, member export, and JSON/YAML encoding.

## Which helper should I use?

| Need | Use | Notes |
| --- | --- | --- |
| Create a typed set | `New[T]` | Use for any comparable element type. |
| Create common primitive sets | `NewString`, `NewInt`, `NewInt32`, `NewInt64`, `NewUint`, `NewUint32`, `NewUint64` | Aliases reduce generic type noise in call sites and serialized structs. |
| Check or update membership | `Add`, `Remove`, `Contains` | Sets are map-backed; mutating methods update the receiver. |
| Run set algebra | `Union`, `Intersect`, `Sub`, `Equal` | Algebra helpers return new sets and leave inputs unchanged. |
| Export members | `Members` | Member order is intentionally undefined. Sort before deterministic output. |
| Encode or decode data | `MarshalJSON`, `UnmarshalJSON`, `MarshalYAML`, `UnmarshalYAML` | Encoded arrays are not ordered unless callers sort exported members themselves. |
| Inject JSON providers | `MarshalJSONWithOptions`, `UnmarshalJSONWithOptions`, `WithSetMarshalFunc`, `WithSetUnmarshalFunc` | Use for tests or integration with alternate JSON implementations. |

## Set correctness checklist

- Set members must be comparable. Use `vslice.UniqBy` or a map keyed by a derived comparable value when elements themselves are not comparable.
- `Members`, `String`, JSON arrays, and YAML sequences use map iteration order. Sort exported members before comparing snapshots, writing stable files, or returning deterministic API payloads.
- `Add` and `Remove` mutate the set. `Union`, `Intersect`, and `Sub` return new sets.
- The zero value of `Set[T]` is a nil map; construct sets with `New` before calling mutating methods such as `Add`.
- JSON and YAML decoding replace the target set with decoded members. Validate input before decoding if partial updates are not acceptable.
- Do not mutate or read/write the same set concurrently without external synchronization.

## When not to use vset

- Use `vslice` when order, duplicate counts, or stable positional operations are part of the data model.
- Use a plain `map[T]struct{}` when you need direct map operations, custom storage layout, or tight control over allocation and synchronization.
- Use a sorted slice or tree structure when deterministic iteration is a core requirement rather than a presentation step.

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

## Related packages

- Use `vslice` when order, duplicates, or index-aware operations matter.
- Use `vmap` when values should be associated with keys rather than represented as membership only.
- Use `vblf` when probabilistic membership with lower memory use is acceptable.

## Benchmarks and trade-offs

Run the focused set tests when changing set behavior:

```bash
go test ./vset
```

Set operations are map-backed: membership checks are average O(1), while `Union`, `Intersect`, `Sub`, `Members`, and serialization allocate proportional to the number of members returned. If deterministic output is needed, budget for sorting after `Members`.

## FAQ

### Are `Members` returned in insertion order?

No. Sets are map-backed, and member order follows Go map iteration. Sort the returned slice before deterministic output or assertions.

### Can I store slices, maps, or functions in a set?

No. `Set[T]` requires `T` to be comparable. Store a comparable key such as an ID, digest, or normalized string instead.

### Does `Union` mutate either input set?

No. `Union`, `Intersect`, and `Sub` return new sets. `Add` and `Remove` mutate the receiver.

### Why does JSON output order change?

JSON marshaling exports `Members`, so array order inherits map iteration order. For stable wire formats, export members, sort them, and marshal the sorted slice yourself.
