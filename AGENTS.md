# AGENTS.md

AI agents working in this repository must follow the same project boundary and validation rules as human contributors.

## Relationship with CLAUDE.md

- `AGENTS.md` is the cross-agent entrypoint. Keep it concise and portable across AI coding agents.
- `CLAUDE.md` is the Claude-specific deep guide. It may contain longer workflow details, package catalogs, and validation playbooks.
- Do not duplicate long policy text in both files. When rules change, update the canonical detailed text in `CLAUDE.md` and keep this file as the short operational summary.
- If an agent only reads one file, this file must still provide enough project contract and validation context to avoid unsafe edits.

## Project contract

- Public APIs live in top-level `v*` facade packages; implementations live under `internal/*`.
- Do not add `v*` to `v*` imports. Shared implementation belongs in `internal/*`.
- Every public facade package must keep `doc.go`, useful exported doc comments, and focused facade tests current.
- Public API changes must update `docs/api/exports.txt` with `UPDATE_API=1 make api-check`.
- Facade, doc comment, or Example changes must update `docs/api/tools.json` and `docs/api/tools.md` with `make tools-gen` or `make docs-gen`.
- Generated documentation artifacts are guarded by `make docs-check` and `make tools-check`.

## Required Go skills

For Go changes, load `golang-how-to` first, then use the relevant Go skills it selects. Common routes:

- Tests or examples: `golang-testing`.
- Doc comments, package docs, AI-readable docs, or generated docs: `golang-documentation`.
- Style, readability, and formatting: `golang-code-style`.
- CI, Makefile, governance, or generated-artifact gates: `golang-continuous-integration`.
- Error contracts: `golang-error-handling` and `golang-safety`.
- Security-sensitive packages (`vhttp`, `vresty`, `vurl`, `vconf`, `vzip`, `vfile`, `vcrypto`, `vjwt`, `vrand`, `vid`, `vdb`): `golang-security` plus `golang-safety`.

## Agent workflow

1. Inspect the existing package, tests, docs, `ai-context.json`, and generated snapshots before editing.
2. Keep changes scoped to the requested logical task; never stage unrelated local work or secrets.
3. Prefer adding behavior tests before implementation changes.
4. Run focused checks first, then the required governance checks for the touched area.
5. If generated artifacts are expected to change, run the generator, review the diff, then run the corresponding check target.
6. Report exact commands run and whether they passed, failed, or were skipped with a reason.

## Validation shortcuts

- `make quick-check`: fast local gate for normal changes.
- `make docs-check`: verifies generated documentation artifacts, including `docs/api/tools.json` and `docs/api/tools.md`.
- `make tools-gen`: regenerates the machine-readable facade tool catalog.
- `make docs-gen`: regenerates generated documentation artifacts.
- `make ai-context-check`: validates AI metadata and command side-effect declarations.
- `make agent-check`: default AI/Agent-safe validation gate.
- `make agent-full-check COVERAGE_FILE=/tmp/go-knifer-coverage.out`: full AI/Agent validation gate when a broad change requires coverage, lint, and vulnerability checks.

## Generated artifacts

- `docs/api/exports.txt` is the public API snapshot.
- `docs/api/tools.json` is the machine-readable tool catalog for public facade functions.
- `docs/api/tools.md` is the human-readable catalog generated from the same source as `tools.json`.
- Do not hand-edit generated JSON, Markdown, or snapshots except to resolve generator bugs; update the generator instead.
