// Package bean provides struct/map property mapping helpers.
//
// Use ToMap and FillMap to materialize struct or map properties as map[string]any.
// Use Copy or CopyProperties when source and destination are already trusted Go values.
// Use Decode or DecodeResult when weak string/numeric/bool conversion is expected.
// Use Merge or MergeResult when multiple sources should be applied to an existing
// destination from left to right.
//
// Boundary with neighboring modules:
//   - internal/ref owns low-level reflection primitives such as field lookup,
//     method discovery, invocation, and raw field mutation.
//   - internal/obj owns object-level predicates, equality, defaults, comparison,
//     cloning, and generic container helpers.
//   - internal/json owns JSON parsing/serialization and JSON-driven Bean/List
//     conversion.
//   - internal/bean owns direct struct/map property mapping: copy properties,
//     tag/alias matching, and weak type conversion without going through JSON.
package bean
