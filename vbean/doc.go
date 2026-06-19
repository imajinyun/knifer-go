// Package vbean provides public APIs for struct and map property mapping.
//
// Use ToMap and FillMap to materialize struct or map properties as map[string]any.
// Use Copy or CopyProperties when source and destination are already trusted Go values.
// Use Decode or DecodeResult when weak string/numeric/bool conversion is expected.
// Use Merge or MergeResult when multiple sources should be applied to an existing
// destination from left to right.
//
// Boundary with neighboring packages:
//   - vref exposes raw reflection primitives and field/method access.
//   - vobj exposes object predicates, equality, defaults, clone/serialization,
//     and generic container helpers.
//   - vjson exposes JSON parsing/serialization and JSON-based Bean/List
//     conversion.
//   - vbean owns direct map/struct property copy, tag/alias matching, and weak
//     type conversion without serializing through JSON.
package vbean
