// Package vbean provides public APIs for struct/map property mapping.
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
