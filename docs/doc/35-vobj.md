# vobj Quickstart

`vobj` provides object emptiness checks, comparisons, defaults, collection membership checks, type information, and serialization-based deep copy helpers.

## Golden path APIs

The first-choice API set for this facade is kept in sync with `ai-context.json` and the generated tools catalog.

- `Clone`
- `CloneIfPossibleWithOptions`
- `Accept`
- `Equals`
- `Apply`

## Which helper should I use?

Choose helpers by whether you need nil/empty checks, generic pointer handling, collection inspection, or serialization-based copying.

| Need | Use | Notes |
| --- | --- | --- |
| Check nil/null or empty values | `IsNil`, `IsNull`, `IsEmpty`, `IsNotEmpty`, batch helpers | Useful for generic input checks; be explicit about zero-value semantics. |
| Inspect length or membership | `Length`, `Contains` | Works across common collections and strings, with reflection-like trade-offs. |
| Compare values | `Equal`, `Equals`, `NotEqual`, `Compare`, `CompareNull` | Prefer typed comparisons when types are known and simple. |
| Apply defaults for optional pointers | `DefaultIfNil`, `DefaultIfNilFunc`, `DefaultIfNilApply` | Keeps nil handling concise without losing type safety. |
| Consume or transform non-nil pointers | `Apply`, `Accept` | Good for optional values in generic helpers. |
| Clone or serialize values | `Clone`, `CloneIfPossible`, `CloneByStream`, `Serialize`, `DeserializeTo` | Serialization behavior depends on exported fields, registered types, and codec options. |
| Customize serialization | `WithEncoderFactory`, `WithDecoderFactory`, `Register`, `RegisterName` | Use when default gob behavior is not enough or tests need injected codecs. |

## Object helper correctness checklist

- Prefer typed Go code when types are known; generic object helpers should make dynamic cases clearer, not obscure simple logic.
- Be precise about nil versus empty: `nil`, `0`, `""`, empty slices, and empty maps often have different business meaning.
- Treat `Contains`, `Length`, and equality helpers as convenience wrappers with reflection-like costs.
- Do not assume serialization-based clones preserve unexported fields, functions, channels, or external resources.
- Register or accept concrete types before deserializing polymorphic values.
- Check clone and deserialize errors unless the `IfPossible` or `OrNil` fallback behavior is intentionally acceptable.

## When not to use vobj

- Use normal typed comparisons, `len`, and nil checks in straightforward code.
- Use `slices`, `maps`, or hand-written loops when collection element types are known and performance matters.
- Use domain-specific copy constructors for values with locks, file handles, network clients, contexts, or other non-serializable resources.
- Avoid generic emptiness checks in validation rules where each field's zero value has different meaning.

## Must API compatibility

`MustDeserialize` is a compatibility helper for trusted fixtures and startup data where malformed bytes should fail fast. New code should prefer `DeserializeTo` or `DeserializeToWithOptions` so malformed or untrusted bytes return errors that callers can classify, log, or recover from.

## Related packages

- Use `vref` when object inspection requires reflection, dynamic fields, or method invocation.
- Use `vbean` when object values need struct-to-struct or map-to-struct copying.
- Use `vjson` when deep copy or comparison workflows rely on serialization boundaries.
- Use `vconf` when object values come from configuration files, profiles, environment expansion, or remote config.

## Boundary with vbean and vconf

`vobj` is the broad convenience layer for dynamic object checks. It should not own configuration loading, profile precedence, schema validation, struct binding policy, or tag-driven property mapping.

Use `vconf` first when data is still configuration text, files, remote URLs, or profile overlays. Use `vbean` when a Go value needs tag-aware struct/map mapping or matched/unused metadata. Use `vobj` after those focused packages have done their work, for generic operations such as nil/default handling, object length checks, comparison, type inspection, or serialization-based cloning.

## Benchmarks and trade-offs

Benchmark object helpers in hot paths, especially collection scans and serialization-based cloning:

```bash
go test -bench=. -benchmem -run=^$ ./internal/obj ./vobj
```

Generic helpers reduce boilerplate in dynamic code but can allocate, reflect, serialize, or hide type-specific behavior. Typed code is usually faster and easier to reason about when the shape is known.

Serialization-based cloning provides deep-copy convenience for supported values, but it is not a universal object copier. Test representative structs before relying on it for isolation.

## FAQ

### Is `IsEmpty` the same as a business validation rule?

No. It is a generic zero/empty check. Business validation should decide whether `0`, `false`, an empty string, or an empty collection is acceptable for each field.

### Does `Clone` copy every Go value perfectly?

No. It depends on the serialization codec and supported field types. Use explicit copy code for resources, locks, unexported state, or custom invariants.

### When should I use `DefaultIfNil`?

Use it when a pointer represents an optional value and a default is safe. Do not use it to hide missing required data.

## Check emptiness, length, and membership

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vobj"
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

	"github.com/imajinyun/knifer-go/vobj"
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

	"github.com/imajinyun/knifer-go/vobj"
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

	"github.com/imajinyun/knifer-go/vobj"
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
