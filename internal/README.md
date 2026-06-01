# Internal package guide

`internal/*` contains implementation packages for go-knifer. Application code should
import the public `v*` facade packages instead of importing from `internal/*`.

## Placement rules

- Put new code in the narrowest domain package that owns the behavior.
- Do not add broad packages such as `extra`, `misc`, `common`, or `util`.
- Keep domain-specific tests next to the internal package implementation.
- Use `vobj` only as a convenience object-level wrapper; do not place new domain
  logic there first.

## Facade rules

- When an internal symbol is exported, decide explicitly whether it should be
  forwarded by the matching `v*` package.
- Public facade APIs should stay thin and stable. They should delegate to
  `internal/*` and avoid duplicating implementation logic.
- Large modules may keep generated `facade.go` files; small modules may use
  hand-written wrappers. Both styles require the same public API review.

## Domain boundaries

- `hash` owns general hash helpers such as additive/FNV and simple digest
  shortcuts. Security-oriented digest, HMAC, encryption, key, and PEM handling
  belong to `crypto`.
- `http` owns lightweight standard-library HTTP helpers. Resty-based chainable
  client behavior belongs to `resty`.
- `codec` owns encoding/decoding algorithms such as Base64, Hex, and URL query
  escaping. URL/URI parsing, normalization, resource, and scheme semantics belong
  to `url`.
- `json` owns JSON objects, arrays, paths, and lightweight XML adapters. XML
  parsing, tree navigation, formatting, namespace handling, and XML-specific
  map/bean conversion belong to `xml`.
- Object-level defaults, nil checks, clone, and serialization convenience may be
  wrapped by `obj`, but string, slice, map, serialization, reflection, and other
  specific behavior should still be implemented in their domain packages first.

## Reserved packages

`db`, `dfa`, and `poi` are intentional placeholders that document future domain
ownership. They should not expose runtime APIs until the corresponding capability
is implemented and reviewed.

## Review checklist

Before adding or moving an internal API, verify:

1. The implementation belongs to a clear domain package.
2. No ambiguous catch-all package is introduced.
3. The matching `v*` facade decision is documented by code or review.
4. Overlapping convenience APIs include comments that point users to the primary
   domain package when the distinction may be unclear.
5. Unit tests cover the internal behavior and, when useful, facade consistency.
