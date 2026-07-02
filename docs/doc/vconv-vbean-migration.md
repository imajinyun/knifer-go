# vconv and vbean Migration Matrix

Use this page when choosing between `knifer-go` conversion/mapping facades and
single-purpose libraries such as `spf13/cast`, `jinzhu/copier`,
`mitchellh/mapstructure`, and `mergo`.

## Matrix

| Workflow | Common specialist | `knifer-go` path | Boundary |
| --- | --- | --- | --- |
| Strict conversion | `strconv` or hand-written parsing | `vconv` `E` helpers | Use strict conversion when invalid input must become an explicit error. |
| Weak conversion | `spf13/cast` | `vconv` defaulting or weak helpers | Use weak conversion only when fallback semantics are intentional and documented. |
| Struct copy | `jinzhu/copier` | `vbean.Copy` / copy result helpers | Use when direct struct/map assignment needs shared tag and field behavior. |
| Map/struct decode | `mitchellh/mapstructure` | `vbean.Decode`, `vconf` binding | Use when config or boundary data needs decode hooks, weak input policy, or unused metadata. |
| Merge | `mergo` or local merge code | `vbean.Merge` / merge result helpers | Use when layered defaults and explicit overwrite behavior matter. |
| Unused metadata | custom validation | `vbean.DecodeResult`, `MergeResult` | Use when the caller must see matched, skipped, or unused keys. |

## Decision Rules

- Use `spf13/cast` when the only task is converting one value and defaulting is
  acceptable.
- Use `vconv` when conversion is part of a broader `knifer-go` workflow or when
  `E` helpers should preserve explicit errors.
- Use `jinzhu/copier` when the only task is struct copying and its behavior is
  already accepted by the project.
- Use `mitchellh/mapstructure` when a project has already standardized on its
  decode hook model.
- Use `vbean` when copy, decode, merge, and unused-key metadata should share the
  same facade and governance model.
- Use `vconf` when decode behavior is part of configuration binding rather than
  a standalone object mapping task.

## Follow-Up

Future implementation work should add examples by behavior group instead of
copying every API from cast, copier, mapstructure, or mergo.
