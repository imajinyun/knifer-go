# Collections Comparison

Use this page when choosing between `knifer-go` collection facades,
`samber/lo`, `duke-git/lancet`, and the Go standard library `slices` / `maps`
packages.

In governance metadata this standard-library baseline is named `stdlib slices/maps`.

## Matrix

| Workflow | Standard library | `samber/lo` | `duke-git/lancet` | `knifer-go` |
| --- | --- | --- | --- | --- |
| Map | Plain `for` loop or `slices` helpers when local code is clearer. | Strong generic `Map` mental model. | Broad helper set for slice mapping. | Use `vslice.Map` / `vmap.Map` when the project already uses `knifer-go` facades. |
| filter | Plain loop with append or map assignment. | Strong generic `Filter` / `Reject` helpers. | Broad predicate helpers. | Use `vslice.Filter`, `vmap.Filter`, or `FilterErr` when callbacks can fail. |
| Reduce | Plain loop when aggregation is local. | Generic `Reduce`. | Broad reduce-style helpers. | Use `Reduce` / `ReduceErr` when behavior should be shared across call sites. |
| Group | Map accumulation in local code. | Generic `GroupBy`. | Collection grouping helpers. | Use `vslice.GroupBy` or `vmap.GroupBy` when grouped output semantics should stay in the facade model. |
| Partition | Plain loop when predicates are domain-specific. | Specialist partition helpers. | Broad split/partition helpers. | Add or use partition helpers only when semantics are reusable and documented. |
| Window | Plain indexed loop for local algorithms. | Specialist window/chunk helpers. | Broad chunk/window helpers. | Use typed window/pair helpers where deterministic examples already describe behavior. |
| Chunk | `for` loop over index steps. | Generic chunk helpers. | Broad chunk helpers. | Use `vslice.Chunk` when chunk behavior should be shared. |
| Set-like helpers | `maps`, `slices`, and explicit map sets. | Generic set-like helpers. | Broad union/intersection helpers. | Use `vset`, `vslice`, or `vmap` when the project also needs other `knifer-go` domains. |

## Decision Rules

- Use the standard library when a local loop is shorter and more explicit.
- Use `samber/lo` when the only need is Lodash-style generic collections.
- Use `duke-git/lancet` when broad helper coverage is the main adoption reason.
- Use `knifer-go` when collection logic is part of a cross-domain workflow that
  also uses safe HTTP, URL, crypto, JSON, file, config, DB, or CLI facades.
- Prefer error-returning helpers such as `MapErr`, `FilterErr`, and `ReduceErr`
  for fallible callbacks.

## Follow-Up

Do not copy every helper from `lo` or `lancet`. Add collection APIs only when a
repeated workflow needs shared semantics, examples, benchmark evidence, and API
snapshot coverage.
