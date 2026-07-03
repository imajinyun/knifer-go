# Utility Library Comparison

`knifer-go` competes with several Go utility libraries, but it does not try to win by having the shortest helper name for every local loop. It is strongest when a project wants a broad toolkit with explicit package boundaries, generated API catalogs, safety defaults, and governance checks.

Last checked: 2026-07-02.

## GitHub Top 5 Utility Libraries

This list uses the project comparison scope "Go utility libraries", not web
frameworks, ORMs, CLI frameworks, or test frameworks.

| Rank | Library | Stars | Scope | Last pushed | License |
| --- | --- | ---: | --- | --- | --- |
| 1 | `samber/lo` | 21,365 | Lodash-style generic collection helpers. | 2026-07-02 | MIT |
| 2 | `duke-git/lancet` | 5,295 | Broad Go utility toolkit with many helper domains. | 2026-03-07 | MIT |
| 3 | `thoas/go-funk` | 4,939 | Reflection-heavy functional helpers for map/find/filter-style workflows. | 2024-07-24 | MIT |
| 4 | `spf13/cast` | 3,981 | Type conversion helpers used by configuration-heavy projects. | 2026-04-12 | MIT |
| 5 | `gookit/goutil` | 2,353 | Daily developer utilities across strings, arrays, maps, env, filesystem, system, CLI, test/assert, and more. | 2026-06-25 | MIT |

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

## Gap Summary

| Competitor | Current gap | Governed path |
| --- | --- | --- |
| `samber/lo` | Collection mindshare is stronger than `knifer-go`'s broad toolkit entry point. | [`collection-golden-paths.md`](collection-golden-paths.md) and [`collections-comparison.md`](collections-comparison.md). |
| `duke-git/lancet` | Broad "many helpers in one toolkit" adoption story is simpler. | [`task-index.md`](task-index.md), [`facade-tiering.md`](facade-tiering.md), and focused `v*` facade docs. |
| `thoas/go-funk` | Dynamic/reflection-heavy workflows are easy to discover in one package. | [`dynamic-data-toolkit-matrix.md`](dynamic-data-toolkit-matrix.md) and `vref` / `vobj` / `vbean` boundaries. |
| `spf13/cast` | Conversion-only mental model is simpler. | [`vconv-cast-migration.md`](vconv-cast-migration.md) and `vconv` explicit-error examples. |
| `gookit/goutil` | Daily utility entry point covers debug/test/dump habits directly. | [`daily-developer-utilities.md`](daily-developer-utilities.md) and [`developer-debug-test-backlog.md`](developer-debug-test-backlog.md). |

## TODO Lanes

- Collections comparison belongs in `docs/doc/collections-comparison.md`.
- Conversion and bean mapping migration belongs in a `vconv` / `vbean` matrix.
- Daily developer utilities belong in a guide that groups `vcli`, `vsys`, `vfile`, `vnet`, `vjob`, and `vlog`.
- Benchmark claims belong in [`benchmark-trust.md`](benchmark-trust.md), not in broad marketing copy.
- Facade tiering belongs in `docs/doc/facade-tiering.md` so day-one imports stay separate from heavy extensions and provider contracts.
- Debug/test helper candidates belong in `docs/doc/developer-debug-test-backlog.md` until `vtest` or `vdump` are implemented.
- Keep this top5 comparison governed by current GitHub metadata and refresh the stars/date deliberately.

## Refresh Workflow

Run `make utility-comparison-refresh` only when intentionally refreshing the
GitHub Top 5 table. The target calls `bin/update_utility_comparison.py --write`,
uses the GitHub API, and updates `docs/doc/utility-library-comparison.md` plus
`ai-context.json` together.

This refresh is explicit opt-in because it requires network access and writes
workspace files. Ordinary gates such as `make docs-check`, `make quick-check`,
`make agent-check`, and `make ci-test` must not depend on it.

## Sources

- GitHub API: `https://api.github.com/repos/samber/lo`
- GitHub API: `https://api.github.com/repos/duke-git/lancet`
- GitHub API: `https://api.github.com/repos/thoas/go-funk`
- GitHub API: `https://api.github.com/repos/spf13/cast`
- GitHub API: `https://api.github.com/repos/gookit/goutil`
