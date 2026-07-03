# Collection Advanced Backlog

Use this page before adding advanced collection APIs. The current public
collection surface is `vslice`, `vmap`, and `vset`; new helpers should only be
added when the workflow is repeated, typed, documented, benchmarked, and clearer
than direct standard-library code.

## Candidate Lanes

| Candidate | Current path | Decision before implementation |
| --- | --- | --- |
| slice partition by predicate | Use explicit loops or `vmap.Partition` for maps; use `vslice.PartitionBy` only for grouped slice partitions. | Define whether the API returns two slices, preserves order, and allocates once or twice. |
| zip N | Use `vslice.Zip2` today. | Decide whether `Zip3` / `ZipN` is worth API surface or whether explicit structs are clearer. |
| cartesian product | Use explicit nested loops. | Require benchmarks and memory-budget guidance before adding because output size grows multiplicatively. |
| channel helpers | Use Go channels, `context`, and explicit goroutines. | Avoid hiding ownership, close, cancellation, and backpressure rules behind generic helpers. |
| parallel transforms | Use `vjob` or explicit worker pools today. | Define ordering, cancellation, panic/error handling, worker limits, and benchmark scope before adding. |
| iterator-first helpers | Use Go range adapters in `vslice` / `vmap` where already available. | Prefer iterator APIs only when they reduce allocations without obscuring control flow. |

## Decision Rules

- Do not copy every helper from `samber/lo` or `duke-git/lancet`.
- Prefer the standard library when local loops are clearer.
- Prefer existing `vslice`, `vmap`, `vset`, or `vjob` paths before adding a new
  helper.
- Require an API decision card before implementation.
- Require executable examples before adding the public API to the catalog.
- Require benchmark evidence before adding helpers that allocate, parallelize,
  or produce combinatorial output.
- Keep error and cancellation contracts explicit.

## Required API Decision Card Questions

1. What repeated workflow needs a public helper instead of a local loop?
2. Which facade owns the helper: `vslice`, `vmap`, `vset`, or `vjob`?
3. Does the helper preserve order, mutate inputs, or share backing arrays?
4. How does it handle invalid sizes, nil inputs, callback errors, context
   cancellation, and panics?
5. What benchmark command proves the helper is acceptable for representative
   input sizes?
6. Which existing `lo`, `lancet`, or standard-library pattern is being compared?

## Machine-Readable Boundaries

- slice partition by predicate
- zip N
- cartesian product
- channel helpers
- parallel transforms
- iterator-first helpers
- do not copy every helper from lo or lancet
- require an API decision card before implementation
- require executable examples before public API
- require benchmark evidence before allocation-heavy helpers
- keep error and cancellation contracts explicit
