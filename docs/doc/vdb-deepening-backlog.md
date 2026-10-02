# vdb Deepening Backlog

`vdb` is the database/SQL facade for parameterized builders, entity helpers, pagination, metadata lookup, and `database/sql` session wrappers. This backlog keeps the next database lane focused on engineering depth rather than broad ORM scope.

## Scope

| Lane | Current baseline | Next hardening target |
| --- | --- | --- |
| Context-first execution | `DB`, `Session`, metadata, query, batch, and transaction methods accept `context.Context`. | Keep every new execution helper context-first and covered by provider-backed tests. |
| Dialect depth | Placeholders, wrappers, pagination, metadata SQL, and upsert SQL cover common dialect families. | Add dialect-specific fixtures only when behavior is stable and deterministic. |
| Batch operations | `ExecBatch` exists on `DB` and `Session`. | Document partial-failure semantics and preserve ordered result/error behavior. |
| Upsert semantics | MySQL, PostgreSQL, and SQLite upsert SQL are covered; unsupported dialects return explicit errors. | Expand fixtures before expanding public API. Unsupported dialects must stay explicit. |
| Scan helpers | `ScanRows`, `ScanOne`, scalar scanning, and entity assignment have focused coverage. | Add typed scan helpers only after nil, byte slice, time, and numeric conversion behavior is specified. |
| Transaction behavior | `Tx` and `Session` helpers keep transaction boundaries explicit. | Preserve rollback-on-error and commit-error contracts; avoid hidden retries. |
| Identifier safety | Builders validate identifiers, and `Raw` is documented as a trusted escape hatch. | Keep identifier allow-list guidance visible in docs and tests. |
| Benchmark scope | Builder benchmarks cover stable paged-order query generation. | Add benchmarks only for stable hot paths, and do not claim database round-trip wins. |

## Non-Goals

- Do not turn `vdb` into an ORM.
- Do not add driver dependencies to the facade package.
- Do not own migrations, schema diffing, connection pooling policy, or driver-specific bulk APIs.
- Do not hide transaction boundaries or retry writes automatically.

## Required Evidence

- `integration/db` runs the public facade contracts against real SQLite,
  PostgreSQL and MySQL; see `integration/db/README.md` for isolated setup and
  `make db-integration-check` for the required CI gate.
- `internal/db/session_exec_test.go` covers session execution, batch, query, transaction, metadata, and unsupported behavior.
- `internal/db/builder_write_test.go` covers write builders and upsert SQL.
- `internal/db/db_sql_helpers_test.go` covers count, metadata SQL, scan helper errors, options, and provider errors.
- `vdb/session_exec_test.go` covers facade execution.
- `vdb/error_contract_test.go` keeps facade errors aligned with the shared error model.

## Validation

Run the focused package tests before changing database behavior:

```bash
go test ./internal/db ./vdb
```

Run governance gates after docs, examples, metadata, or public API changes:

```bash
make docs-check
make ai-context-check
make governance-maturity-check
make agent-security-check
```
