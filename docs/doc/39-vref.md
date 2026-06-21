# vref Quickstart

`vref` provides `reflect`-based helpers for type checks, field reads/writes, method lookup, dynamic construction, and function calls.

## When to use vref

| Scenario | Use `vref` when | Prefer direct Go code when |
| --- | --- | --- |
| Framework or adapter code needs dynamic field access | Field names, tags, or methods are discovered at runtime. | Types are known at compile time and ordinary method calls or field access are available. |
| Tests need compact reflection assertions | You are checking generic behavior across multiple struct shapes. | A table-driven test over typed functions is clearer. |
| Dynamic invocation is part of the API | Plugin, binding, or mapping code receives arbitrary functions or methods. | The call target is fixed; reflection would hide compile-time type errors. |

## Which helper should I use?

Start with type/value inspection helpers, then move to field, method, construction, or invocation helpers only when dynamic behavior is required.

| Need | Use | Notes |
| --- | --- | --- |
| Inspect dynamic type and nil-ness | `TypeOf`, `IndirectType`, `IsNil` | Useful before reflection calls that would panic on invalid or nil values. |
| Check or read struct fields | `HasField`, `GetFieldValue`, `GetFieldValueWithOptions` | Prefer exported fields. Unsafe access should stay rare and documented at the call site. |
| Modify struct fields | `SetFieldValue` | Pass a pointer to a settable value and handle errors explicitly. |
| Find and invoke methods | `GetMethod`, `InvokeMethod`, `Invoke` | Check method existence and argument compatibility before calling. |
| Construct values dynamically | `NewInstance` | Useful for factories or binding layers; validate the constructor signature first. |
| Call arbitrary functions | `InvokeFunc` | Prefer typed function calls when the function is known at compile time. |

## Reflection safety checklist

- Prefer typed Go code first. Reflection removes compile-time guarantees and should be limited to boundaries that truly need dynamic behavior.
- Guard nil, invalid, and non-pointer values before field writes. A set operation requires an addressable, settable value.
- Keep `WithUnsafeAccess(true)` out of normal application paths unless there is a documented reason to read unexported fields.
- Treat dynamic invocation errors as caller-visible validation failures; do not panic on mismatched argument counts or types.
- Keep reflection helpers close to the adapter or framework boundary so ordinary business logic remains typed and easy to refactor.

## Related packages

- Use `vobj` when object checks, defaults, comparison, or deep-copy helpers are enough without dynamic invocation.
- Use `vbean` when reflection is only needed to copy values between structs or maps.
- Use `vconv` when dynamic values need explicit scalar conversion after reflection lookup.

## When not to use vref

- Use ordinary typed Go code when types, fields, methods, or constructors are known at compile time.
- Use interfaces, generics, or small adapter functions when they preserve compile-time checks and keep call sites readable.
- Avoid unsafe field access in normal application paths; prefer exported fields, methods, or explicit test-only helpers.
- Avoid dynamic invocation for authorization, billing, quota, or other correctness-critical decisions unless inputs and errors are validated explicitly.
- Avoid reflection-heavy hot paths until metadata caching and typed alternatives have been benchmarked with representative data.

## Benchmarks and trade-offs

Benchmark reflection-heavy adapters with the same struct shapes, field counts, and method signatures used in production:

```bash
go test -bench=. -benchmem -run=^$ ./internal/ref ./vref
```

Reflection helpers reduce repeated `reflect` boilerplate and centralize error handling, but they cannot restore compile-time type safety. Field lookup, dynamic invocation, and unsafe access are all slower and harder to refactor than typed code.

Cache reflection metadata in higher-level code when the same type is inspected repeatedly. Keep `WithUnsafeAccess(true)` limited to tests, migration tools, or documented adapter boundaries.

## FAQ

### Does vref make reflection type-safe?

No. It wraps common reflection operations with helper APIs and errors, but callers still need to validate dynamic inputs and handle failures.

### When is unsafe field access acceptable?

Use unsafe access only for narrow tooling, migration, or test scenarios where exported access is unavailable and the trade-off is documented. Prefer exported fields or methods in normal application code.

### Why not use reflection everywhere for flexibility?

Reflection hides compiler checks, can be slower, and makes refactors harder. Use it at dynamic boundaries; keep core logic typed.

## Get type and value information

```go
package main

import (
	"fmt"
	"reflect"

	"github.com/imajinyun/go-knifer/vref"
)

func main() {
	type User struct{ Name string }
	u := &User{Name: "alice"}

	typ := vref.TypeOf(u)
	fmt.Println(typ.Kind() == reflect.Pointer)
	fmt.Println(vref.IndirectType(typ).Name())
	fmt.Println(vref.IsNil((*User)(nil)))
}
```

## Read and modify struct fields

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vref"
)

type User struct {
	Name string `json:"name"`
	age  int
}

func main() {
	u := &User{Name: "alice", age: 18}

	fmt.Println(vref.HasField(u, "name"))
	fmt.Println(vref.GetFieldValue(u, "name"))

	if err := vref.SetFieldValue(u, "Name", "bob"); err != nil {
		panic(err)
	}
	fmt.Println(u.Name)

	fmt.Println(vref.GetFieldValueWithOptions(u, "age", vref.WithUnsafeAccess(true)))
}
```

## Find and call methods

```go
package main

import (
	"fmt"
	"reflect"

	"github.com/imajinyun/go-knifer/vref"
)

type Counter struct{}

func (Counter) Add(a, b int) int { return a + b }

func main() {
	c := Counter{}
	method, ok := vref.GetMethod(c, false, "Add", reflect.TypeOf(1), reflect.TypeOf(2))
	if !ok {
		panic("method not found")
	}
	got, err := vref.InvokeMethod(c, method, 2, 3)
	if err != nil {
		panic(err)
	}
	fmt.Println(got)

	got, err = vref.Invoke(c, "Add", 4, 5)
	if err != nil {
		panic(err)
	}
	fmt.Println(got)
}
```

## Dynamically construct values and call functions

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vref"
)

type User struct{ Name string }

func NewUser(name string) User { return User{Name: name} }

func main() {
	created, err := vref.NewInstance(NewUser, "alice")
	if err != nil {
		panic(err)
	}
	fmt.Println(created.(User).Name)

	got, err := vref.InvokeFunc(func(a, b int) int { return a + b }, 6, 7)
	if err != nil {
		panic(err)
	}
	fmt.Println(got)
}
```
