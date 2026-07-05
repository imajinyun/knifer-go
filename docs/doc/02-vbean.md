# vbean Quickstart

`vbean` maps fields between structs and maps. Use `Copy` / `CopyProperties` for trusted Go-to-Go property copy, `Decode` / `DecodeResult` for weak string/numeric/bool input conversion with metadata, and `Merge` / `MergeResult` when multiple sources should update an existing destination from left to right.

## Golden path APIs

The first-choice API set for this facade is kept in sync with `ai-context.json` and the generated tools catalog.

- `CopyProperties`
- `ComposeDecodeHook`
- `Copy`
- `Decode`
- `DecodeResult`

## Which helper should I use?

Choose the helper by the mapping direction and how much conversion metadata the caller needs.

| Need | Use | Notes |
| --- | --- | --- |
| Convert a struct or map to `map[string]any` | `ToMap` | Good for serialization boundaries, diagnostics, and generic map workflows. |
| Fill a struct from a map or another struct | `ToStruct`, `Decode` | Destination must be a pointer. Enable weak conversion only when the input boundary needs it. |
| Copy trusted Go values between compatible shapes | `Copy`, `CopyProperties` | Use when source and destination are already trusted Go values and conversion policy is simple. |
| Track matched and unused fields | `DecodeResult`, `MergeResult` | Use metadata when callers must reject unused input or explain mapping decisions. |
| Merge layered sources into an existing destination | `Merge`, `MergeWithOptions`, `MergeResultWithOptions` | Later sources override earlier sources; keep that precedence visible at the call site. |
| Customize matching or conversion | `WithTagNames`, `WithCaseInsensitive`, `WithWeaklyTyped`, parser options, `WithDecodeHook` | Prefer explicit per-call options over hidden global conversion policy. |

## Mapping safety checklist

- Pass a pointer destination for struct writes, and handle mapping errors instead of assuming every field is assignable.
- Enable weak typing only at input boundaries that intentionally accept string/numeric/bool coercion.
- Use `WithStrictUnused(true)` or inspect `Result.Unused` when extra source fields should fail validation.
- Keep tag-name precedence explicit with `WithTagNames` when JSON, DB, form, or config tags differ.
- Be deliberate with `WithIgnoreEmpty` and `WithIgnoreZero`; they can preserve existing values but can also hide intentional zero-value updates.
- Treat map inputs from users or config files as untrusted until required fields, unused fields, and type conversions have been checked.

## Decode hook and field-path errors

`WithDecodeHook` adds a per-call conversion hook for cases such as string-to-time, string-to-duration, string-to-domain enum, or provider-backed value normalization. Hooks receive the source type, destination type, and current value. Return the original value unchanged when the hook does not handle the conversion. Do not use package-level mutable registries; keep conversion policy local to the boundary call so reviews and tests can see it.

Use `ComposeDecodeHook` to chain focused hooks from left to right. Built-in hooks cover the common scalar extension points:

- `StringToTimeHook(layout)` converts strings to `time.Time` using the supplied layout.
- `StringToDurationHook()` converts strings such as `"250ms"` or `"1h30m"` to `time.Duration`.

Decode and copy errors include the destination field path as human-readable context. Nested slice, map, and struct failures add fragments such as `bean: set field Items`, `index 0`, `map key`, `map value`, and the nested field name. Tests should assert both `errors.Is(err, knifer.ErrCodeInvalidInput)` and the relevant path fragments when adding a new mapping path.

Decode contract:

| Behavior | Contract |
| --- | --- |
| Weak conversion | Enabled by default for string/numeric/bool/slice/map/struct assignment; disable with `WithWeaklyTyped(false)`. |
| Decode hook order | `ComposeDecodeHook` runs hooks left to right; each hook sees the previous hook's value. |
| Metadata | `DecodeResult` reports `Matched`, `Skipped`, and `Unused`; use `WithStrictUnused(true)` to fail on unused input. |
| Field path errors | Assignment failures wrap the destination field path and nested index/map fragments. |

Merge strategy contract:

| Strategy point | Current behavior |
| --- | --- |
| Source precedence | Sources are applied left to right; later sources override earlier sources. |
| Zero values | Zero values overwrite by default; `WithIgnoreZero(true)` skips zero source values. |
| Empty values | Empty strings/slices/maps overwrite by default; `WithIgnoreEmpty(true)` skips empty source values. |
| Slices | Matching slice fields are replaced, not appended. |
| Maps | Matching map fields are replaced through assignment/conversion, not deep-merged. |
| Type mismatch | Returns the first field-path error and stops the merge. |

Copy contract:

| Behavior | Contract |
| --- | --- |
| Field matching | Matches exported struct fields and map keys by configured tag names, aliases, and case-insensitive matching. |
| Ignored fields | Tags with `-` are omitted from source collection and destination matching. |
| Method copying | Methods are not invoked or copied; use an explicit mapper when method-derived fields are required. |
| Pointer allocation | Destination must be a non-nil pointer or `map[string]any`; pointer fields are assigned only through normal assignment/conversion paths. |
| Required fields | Required-field validation is not implicit; inspect `Result.Unused`, use `WithStrictUnused(true)`, or validate after mapping. |

## When not to use vbean

- Use hand-written mappers when source and destination types are stable, business rules are complex, or reviewers need explicit field-by-field behavior.
- Use strict decoders or validators before mapping when input must reject unknown fields, malformed values, or policy-specific combinations.
- Avoid weak conversion for authorization, billing, quota, or compliance decisions unless invalid and unused fields are surfaced and tested.
- Use typed constructors instead of generic map-to-struct mapping for library APIs and domain objects with invariants.
- Avoid reflection-based mapping in hot paths until it has been benchmarked against direct typed code.

## Related packages

- Use `vconv` for explicit scalar conversions before or after struct mapping.
- Use `vmap` when the source data is map-shaped and does not need struct binding.
- Use `vjson` when the mapping boundary starts from JSON payloads or fixtures.
- Use `vconf` to load, layer, validate, and watch configuration before binding it into structs.
- Use `vobj` only for generic nil/empty/default/clone checks around already-mapped values.

## Boundary with vconf and vobj

`vbean` owns property mapping, not configuration source management or broad object utilities. A typical config-to-struct flow should load and validate with `vconf`, map or bind with `vbean` only when a Go value shape must change, and reserve `vobj` for generic object checks such as optional pointer defaults or serialization-based cloning.

Do not put file loading, profile precedence, environment expansion, or remote URL policy in `vbean`; those belong to `vconf`. Do not add generic emptiness, length, comparison, or clone helpers to `vbean`; those belong to `vobj` or typed Go code.

## Benchmarks and trade-offs

Use the bean benchmark suite to measure reflection and conversion overhead on your machine:

```bash
go test -bench=. -benchmem -run=^$ ./vbean
```

The suite covers representative `DecodeResult` and `Merge` workflows. Treat the output as a local baseline rather than a universal performance claim. For hot paths with known types, compare a hand-written typed mapper against the `vbean` helper.

## FAQ

### Does vbean replace hand-written mappers?

No. `vbean` reduces boilerplate for dynamic or tag-driven mapping. Hand-written mappers remain clearer and faster when source and destination types are stable and business rules are complex.

### When should I use DecodeResult instead of Decode?

Use `DecodeResult` when the caller needs to know what matched, what was unused, or why a boundary payload should be rejected. Use `Decode` when only the populated destination and error matter.

### Is weak conversion safe by default?

Weak conversion is convenient, not a validation substitute. Enable it deliberately and pair it with required-field and unused-field checks where input comes from users, config files, or external systems.

## Convert a struct to a map

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vbean"
)

type UserDTO struct {
	Name string `json:"name"`
	Age  string `json:"age"`
}

func main() {
	m, err := vbean.ToMap(UserDTO{Name: "alice", Age: "18"})
	if err != nil {
		panic(err)
	}
	fmt.Println(m["name"], m["age"])
}
```

## Fill a struct from a map

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vbean"
)

type User struct {
	Name string `json:"full_name"`
	Age  int    `json:"age"`
}

func main() {
	var user User
	err := vbean.ToStruct(map[string]any{"FULL_NAME": "drew", "age": "21"}, &user,
		vbean.WithCaseInsensitive(true),
		vbean.WithWeaklyTyped(true),
	)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s:%d\n", user.Name, user.Age)
}
```

## Use custom tags and skip zero values

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vbean"
)

type Row struct {
	Name string `db:"user_name"`
	Age  int    `db:"age"`
}

func main() {
	m, err := vbean.ToMap(Row{Name: "casey", Age: 0},
		vbean.WithTagNames("db"),
		vbean.WithIgnoreZero(true),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(m) // age is skipped
}
```

## Decode weak input with metadata

Use `DecodeResult` when callers need to know which source fields matched the destination and which inputs were unused.

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vbean"
)

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	var user User
	result, err := vbean.DecodeResult(map[string]any{"name": "Kai", "age": "34", "extra": true}, &user)
	if err != nil {
		panic(err)
	}

	fmt.Println(user)
	fmt.Println(result.Matched)
	fmt.Println(result.Unused)
}
```

## Copy fields while preserving existing non-empty values

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vbean"
)

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	dst := User{Name: "existing", Age: 30}
	err := vbean.Copy(map[string]any{"name": "", "age": "22"}, &dst,
		vbean.WithIgnoreEmpty(true),
	)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", dst)
}
```

## Merge multiple sources into an existing value

Use `Merge` when later sources should override earlier sources while preserving unmatched destination fields.

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vbean"
)

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	user := User{Name: "existing", Age: 18}
	if err := vbean.Merge(&user, map[string]any{"name": "new"}, map[string]any{"age": "21"}); err != nil {
		panic(err)
	}

	fmt.Println(user)
}
```
