// Package bean provides struct/map property mapping helpers.
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
