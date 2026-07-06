# Adoption Trust

Use this page when deciding whether to adopt `knifer-go` in a project. It lists
the evidence that humans and AI agents can inspect before relying on the public
facade packages.

## Trust Signals

| Area | Evidence |
| --- | --- |
| Release notes | User-visible changes are tracked in [`CHANGELOG.md`](../../CHANGELOG.md). |
| Compatibility policy | Public compatibility is the top-level `v*` facade surface plus [`docs/api/exports.txt`](../api/exports.txt). |
| Deprecation policy | Deprecated APIs must name a replacement, explain migration, stay available for at least two minor releases, and appear in release notes before removal. |
| Security policy | Vulnerability reporting and disclosure expectations are documented in [`SECURITY.md`](../../SECURITY.md). |
| Generated API catalog | [`docs/api/tools.json`](../api/tools.json) and [`docs/api/tools.md`](../api/tools.md) are generated from public facade source and examples. |
| Validation gates | `make agent-check`, `make agent-full-check`, `make release-check`, `make api-freeze-check`, and `make release-notes-check` cover local and release readiness. |
| Benchmark evidence | [`benchmark-trust.md`](benchmark-trust.md) separates quick benchmark gates from manual opt-in benchmark evidence. |

Machine-readable phrases:

- generated API catalog
- validation gates
- why trust this library
- docs/doc/benchmark-trust.md

## Why Trust This Library

- Public APIs live in top-level `v*` facade packages and are checked against
  generated API snapshots.
- Implementation details stay under `internal/*`, so application code has one
  public boundary to import.
- Generated docs and AI metadata are part of the gate, not separate prose that
  can silently drift.
- Security-sensitive packages document Safe/E/WithOptions paths for trust
  boundaries such as URLs, paths, archives, config, SQL, command arguments,
  tokens, and credentials.
- Benchmark output is treated as evidence, not a universal performance claim.
- Release notes, compatibility policy, deprecation policy, and security policy
  are linked from the root README before adoption.

## Adoption Checklist

1. Read the first-use path in [`first-use-golden-paths.md`](first-use-golden-paths.md).
2. Check the package quickstart for the facade you plan to import.
3. Review [`docs/api/exports.txt`](../api/exports.txt) for public API stability.
4. Review [`CHANGELOG.md`](../../CHANGELOG.md) before upgrading.
5. Review [`SECURITY.md`](../../SECURITY.md) before reporting security issues.
6. Run `make quick-check` locally before small changes and `make agent-check`
   before publishing agent-generated changes.

## Governance Validation Contracts

Use this release summary template when a change adds, removes, or tightens a
governance gate. Keep the summary user-facing: describe what is now enforced,
who must act, and which command proves the contract.

```markdown
### Governance Validation Contracts

- Contract changed: <machine gate, metadata section, generated artifact, or CI target>
- User impact: <what maintainers or adopters must do differently>
- Required action: <command, metadata update, evidence field, or release note>
- Validation evidence: <exact command output or artifact path>
- Compatibility note: <whether existing public APIs or workflows changed>
```
