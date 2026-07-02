# Utility Library Comparison

`knifer-go` competes with several Go utility libraries, but it does not try to win by having the shortest helper name for every local loop. It is strongest when a project wants a broad toolkit with explicit package boundaries, generated API catalogs, safety defaults, and governance checks.

## Comparison Matrix

| Library | Best fit | Use `knifer-go` when |
| --- | --- | --- |
| `samber/lo` | Lodash-style generic collection helpers with compact map/filter/group APIs. | The same project also needs safe HTTP, URL, crypto, JWT, JSON, file, config, DB, logging, or provider-injected helpers. |
| `duke-git/lancet` | Broad utility coverage with a simple "many helpers in one toolkit" adoption story. | The decision depends on explicit safety boundaries, generated API metadata, facade packages, and machine-checked governance. |
| `thoas/go-funk` | Reflection-heavy functional helpers where generic typing is not the main concern. | The call site benefits from focused `vslice`, `vmap`, or `vstr` helpers with clearer package ownership and less reflection-driven behavior. |
| `gookit/goutil` | Daily development utilities across strings, arrays, maps, structs, env, filesystem, system, and CLI helpers. | A project wants daily utilities plus security-focused HTTP/URL/crypto/JWT/database boundaries in one facade model. |
| `spf13/cast` | Standalone type conversion helpers with a strong "cast this value" mental model. | Conversion is part of a broader workflow involving `vconv`, `vbean`, `vconf`, object mapping, config loading, or explicit error contracts. |

## Decision Rules

- Use the standard library when a short local loop or direct call is clearer.
- Use a specialist library when the task is only one narrow domain, such as collection transforms, casting, or struct copying.
- Use `knifer-go` when the workflow crosses domains or needs a cross-domain toolkit with safety defaults, explicit errors, provider injection, generated tool metadata, or governance gates.
- Do not import `internal/*` packages from applications. Public APIs live in top-level `v*` facade packages.

## Follow-Up Lanes

- Collections comparison belongs in `docs/doc/collections-comparison.md`.
- Conversion and bean mapping migration belongs in a `vconv` / `vbean` matrix.
- Daily developer utilities belong in a guide that groups `vcli`, `vsys`, `vfile`, `vnet`, `vjob`, and `vlog`.
- Benchmark claims belong in public benchmark trust documentation, not in broad marketing copy.
