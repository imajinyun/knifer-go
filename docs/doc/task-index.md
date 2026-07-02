# Task Index

Use this page when a developer or AI agent knows the task but does not know
which facade to import. Each task has one default facade and a small set of
related facades. Start with the default facade; move to related facades only
when the workflow crosses that boundary.

## Day-One Tasks

| Task | Default facade | Related facades | Boundary |
| --- | --- | --- | --- |
| string cleanup | `vstr` | `vregex`, `vdfa`, `vhan`, `vtok` | Use direct `strings` or `unicode` calls when a local expression is clearer. |
| slice transformation | `vslice` | `vmap`, `vset`, `vjob` | Use plain loops when allocation, side effects, or early returns must be explicit. |
| map transformation | `vmap` | `vslice`, `vset`, `vbean` | Sort keys before deterministic output; map iteration order is not stable. |
| JSON path and formatting | `vjson` | `vxml`, `vfile`, `vobj` | Use `encoding/json.Decoder` directly for streaming or strict decoder policy. |
| file IO | `vfile` | `vzip`, `vurl`, `vhttp` | Handle path policy before calling file helpers for untrusted input. |
| safe HTTP | `vhttp` | `vresty`, `vurl`, `vnet` | Use Safe/E/WithOptions variants when URLs come from users, config, queues, or service discovery. |
| crypto | `vcrypto` | `vrand`, `vjwt`, `vpass`, `vhash` | Use `vhash` only for non-cryptographic hashing. |
| configuration | `vconf` | `vbean`, `vconv`, `vform`, `vfile` | Use safe remote config helpers for untrusted URLs. |
| database | `vdb` | `vconf`, `vcli` | Values use placeholders; identifiers still need validation or static review. |
| CLI command execution | `vcli` | `vsys`, `vfile`, `vlog` | Pass command arguments as slices; avoid shell concatenation for untrusted input. |

## Star Domains

| Domain | Default facade | Related facades | Start here |
| --- | --- | --- | --- |
| Safe HTTP | `vhttp` | `vresty`, `vurl`, `vnet` | [`safe-http-cookbook.md`](safe-http-cookbook.md) |
| Safe Crypto | `vcrypto` | `vrand`, `vjwt`, `vpass` | [`safe-crypto-cookbook.md`](safe-crypto-cookbook.md) |
| Daily JSON/File | `vjson` | `vfile`, `vxml`, `vzip` | [`daily-json-file-faq.md`](daily-json-file-faq.md) |

## Daily Domains

| Domain | Default facade | Related facades | Start here |
| --- | --- | --- | --- |
| Daily developer utilities | `vcli` | `vsys`, `vfile`, `vnet`, `vjob`, `vlog`, `vconf` | [`daily-developer-utilities.md`](daily-developer-utilities.md) |
| Collection workflows | `vslice` | `vmap`, `vset` | [`collection-golden-paths.md`](collection-golden-paths.md) |
| Dynamic data workflows | `vconf` | `vbean`, `vjson`, `vobj`, `vref`, `vconv` | [`dynamic-data-toolkit-matrix.md`](dynamic-data-toolkit-matrix.md) |

## Selection Rules

- Choose one default facade first.
- Use related facades only when the workflow crosses package boundaries.
- Prefer Safe, E, context-aware, or WithOptions flows at trust boundaries.
- Prefer typed Go code when it is shorter and clearer.
- Do not import `internal/*` from application code.

## Machine-Readable Coverage

- string cleanup
- slice transformation
- map transformation
- JSON path and formatting
- file IO
- safe HTTP
- crypto
- configuration
- database
- CLI command execution
- Safe HTTP
- Safe Crypto
- Daily JSON/File
- Daily developer utilities
- Collection workflows
- Dynamic data workflows
- choose one default facade first
- related facades only when workflow crosses package boundaries
- Safe/E/WithOptions flows at trust boundaries
- do not import internal packages
