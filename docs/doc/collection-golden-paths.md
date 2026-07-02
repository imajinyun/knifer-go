# Collection Golden Paths

Use this page when the task is collection work and the caller needs one
recommended `knifer-go` facade before comparing alternatives. It complements
[`collections-comparison.md`](collections-comparison.md): that page compares
libraries, while this page starts from the workflow.

## Task Index

| Task | Recommended facade | Shortest `knifer-go` path | Standard library path | `samber/lo` path | `duke-git/lancet` path |
| --- | --- | --- | --- | --- | --- |
| map | `vslice` / `vmap` | `vslice.Map` for slices, `vmap.MapValues` for maps | Plain `for` loop | `lo.Map` | broad map helpers |
| filter | `vslice` / `vmap` | `vslice.Filter`, `vmap.Filter` | Plain loop with append or assignment | `lo.Filter` | broad filter helpers |
| reduce | `vslice` / `vmap` | `vslice.Reduce`, `vmap.Reduce` | Plain accumulator loop | `lo.Reduce` | broad reduce helpers |
| group | `vslice` / `vmap` | `vslice.GroupBy`, `vmap.GroupBy` | Map accumulation loop | `lo.GroupBy` | broad group helpers |
| chunk | `vslice` | `vslice.Chunk` | Indexed loop | `lo.Chunk` | broad chunk helpers |
| window | `vslice` | `vslice.Window` or `vslice.Sliding` | Indexed loop | window helpers when available | broad window helpers |
| set | `vset` | `vset.New`, `Union`, `Intersect`, `Sub` | `map[T]struct{}` | set-like helpers | broad set helpers |
| zip | `vslice` | `vslice.Zip2` / `vslice.Unzip2` | Indexed loop | zip helpers | broad zip helpers |
| partition | `vmap` / `vslice` | `vmap.Partition` for maps, `vslice.PartitionBy` for grouped slice partitions | Plain loop | partition helpers | broad partition helpers |
| find | `vslice` / `vmap` | `vslice.Find`, `vmap.Find` | Plain loop with early return | `lo.Find` | broad find helpers |
| contains | `vslice` / `vmap` / `vset` | `vslice.Contains`, `vmap.ContainsKey`, `vset.Contains` | `slices.Contains`, map lookup | `lo.Contains` | broad contains helpers |

## Decision Rules

- Use the standard library when a local loop is shorter and clearer.
- Use `samber/lo` when the only need is Lodash-style generic collection helpers.
- Use `duke-git/lancet` when broad helper coverage is the main adoption reason.
- Use `knifer-go` when collection work is part of a cross-domain workflow that
  also uses safe HTTP, URL, crypto, JSON, file, config, database, CLI, cache, or
  logging facades.
- Use error-returning helpers such as `MapErr`, `FilterErr`, and `ReduceErr`
  when callbacks can fail.
- Keep map ordering explicit. Use sorted-key helpers before logging, snapshot
  tests, generated files, or API responses that require deterministic order.

Machine-readable boundaries:

- stdlib slices/maps
- workflow-first collection entry point
- standard library first for local loops
- samber/lo for collection-only lodash-style helpers
- lancet for broad helper coverage
- knifer-go for cross-domain facade workflows
- do not copy every helper from lo or lancet

## Short Examples

### Map

```go
doubled := vslice.Map([]int{1, 2}, func(n int) int { return n * 2 })
labels := vmap.MapValues(map[string]int{"a": 1}, func(k string, v int) string { return k })
```

### Filter

```go
even := vslice.Filter([]int{1, 2, 3}, func(n int) bool { return n%2 == 0 })
kept := vmap.Filter(map[string]int{"a": 1}, func(k string, v int) bool { return v > 0 })
```

### Reduce

```go
sum := vslice.Reduce([]int{1, 2, 3}, 0, func(acc, n int) int { return acc + n })
total := vmap.Reduce(map[string]int{"a": 1}, 0, func(acc int, k string, v int) int { return acc + v })
```

### Group

```go
byLen := vslice.GroupBy([]string{"go", "js", "java"}, func(s string) int { return len(s) })
grouped := vmap.GroupBy([]string{"go", "js"}, func(s string) int { return len(s) })
```

### Chunk And Window

```go
chunks := vslice.Chunk([]int{1, 2, 3, 4}, 2)
windows := vslice.Window([]int{1, 2, 3}, 2)
```

### Set

```go
unique := vset.NewString("go", "go", "tool")
hasTool := unique.Contains("tool")
```

### Zip

```go
pairs := vslice.Zip2([]string{"a", "b"}, []int{1, 2})
```

### Partition

```go
matched, rest := vmap.Partition(map[string]int{"a": 1, "b": 2}, func(k string, v int) bool { return v%2 == 0 })
groups := vslice.PartitionBy([]int{1, 1, 2}, func(n int) int { return n })
_ = matched
_ = rest
_ = groups
```

### Find And Contains

```go
first, ok := vslice.Find([]int{1, 2, 3}, func(n int) bool { return n > 1 })
exists := vmap.ContainsKey(map[string]int{"a": 1}, "a")
_ = first
_ = ok
_ = exists
```

## Follow-Up Boundary

Do not copy every helper from `lo` or `lancet`. Add collection APIs only when a
repeated workflow needs shared semantics, executable examples, benchmark
evidence, and API snapshot coverage.
