# vconv Cast Migration Cookbook

Use this page when migrating from `spf13/cast` or deciding whether scalar
conversion should stay in `vconv`. `cast` is a strong single-purpose conversion
library; `vconv` is best when conversion is part of the wider `knifer-go`
facade model and invalid input needs explicit boundaries.

## Matrix

| Workflow | `spf13/cast` habit | `vconv` path | Boundary |
| --- | --- | --- | --- |
| strict conversion | use `ToXxxE` variants from cast | `ToIntE`, `ToInt64E`, `ToFloat64E`, `ToBoolE` | Use strict conversion when invalid input must return an error. |
| weak conversion | use permissive `ToXxx` helpers | `ToInt`, `ToInt64`, `ToFloat64`, `ToBool`, `ToString`, `ToBytes` | Use weak conversion only when zero-value fallback is intentional. |
| default fallback | use `ToXxxDefault` helpers | `ToIntDefault`, `ToInt64Default`, `ToFloat64Default`, `ToBoolDefault`, `ToStringDefault` | Use default fallback when zero, false, or empty string is a valid value. |
| custom parser policy | wrap parser logic around cast | `WithParseIntFunc`, `WithParseFloatFunc`, `WithBoolParser`, formatter options | Keep custom conversion policy visible at the call site. |
| slice/map conversion | use cast helpers for broad shapes | use `vslice`, `vmap`, `vbean`, or typed loops | `vconv` is scalar-first; keep collection or struct mapping in the focused facade. |
| duration/time conversion | use cast duration/time helpers | use `time.ParseDuration`, `vdate`, or `vconv` only for scalar duration representation | Duration/time grammar is a domain contract, not just generic conversion. |
| overflow handling | rely on cast behavior | use `E` helpers for overflow-aware integer narrowing | `E` helpers reject overflow; zero/default helpers preserve legacy permissive behavior. |

## Migration Rules

- Use `vconv` for scalar conversion when the project already uses `knifer-go`
  facades or when generated API metadata and error contracts matter.
- Use `spf13/cast` when conversion is the only dependency need and its behavior
  is already accepted by the project.
- Use `E` helpers at trust boundaries: request values, config values, database
  values, CLI flags, and user-controlled data.
- Use default helpers when fallback behavior should be obvious in code review.
- Do not move collection conversion into `vconv`; use `vslice`, `vmap`, `vbean`,
  or direct typed loops.
- Do not use generic conversion as validation for currency, quota, identity,
  authorization, or security policy decisions.

## Recipes

### strict conversion

```go
port, err := vconv.ToIntE("8080")
if err != nil {
	return err
}
_ = port
```

### weak conversion

```go
count := vconv.ToInt("not-a-number")
_ = count // 0
```

### default fallback

```go
limit := vconv.ToIntDefault("bad", 100)
_ = limit
```

### custom parser policy

```go
value, err := vconv.ToIntEWithOptions("max", vconv.WithParseIntFunc(func(s string, base, bitSize int) (int64, error) {
	if s == "max" {
		return 100, nil
	}
	return strconv.ParseInt(s, base, bitSize)
}))
_ = value
_ = err
```

### slice/map conversion

```go
numbers := vslice.Map([]string{"1", "2"}, vconv.ToInt)
picked := vmap.Pick(map[string]any{"port": 8080, "debug": true}, "port")
_ = numbers
_ = picked
```

### duration/time conversion

```go
timeout, err := time.ParseDuration("150ms")
_ = timeout
_ = err
```

### overflow handling

```go
_, err := vconv.ToInt64E(uint64(math.MaxInt64) + 1)
_ = err
```

## Machine-Readable Boundaries

- strict conversion
- weak conversion
- default fallback
- custom parser policy
- slice/map conversion
- duration/time conversion
- overflow handling
- spf13/cast migration
- vconv is scalar-first
- E helpers at trust boundaries
- do not move collection conversion into vconv
