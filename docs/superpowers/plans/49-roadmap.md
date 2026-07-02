# knifer-go Next-Phase Capability Roadmap

This roadmap tracks the next development phase after the tool catalog quality sprints. It prioritizes scenario mindshare, deeper high-value modules, documentation and benchmark trust, and explicit ecosystem adapter lanes.

## Baseline

This baseline is derived from `docs/api/tools.json.summary`. `make governance-maturity-check` validates the table so roadmap numbers cannot silently drift from the generated public API catalog.

| Metric | Value |
| --- | ---: |
| Public facade packages | 55 |
| Public functions | 2757 |
| Functions with executable examples | 1712 |
| Context-aware functions | 36 |
| Functions returning errors | 687 |
| Recommended public functions | 2735 |
| Compatibility public functions | 22 |
| Empty function synopses | 0 |
| Facade-sourced function synopses | 2104 |
| Internal-sourced function synopses | 653 |

## 90-Day Star Domain Scorecard

| Domain | Public functions | Examples | Example ratio | Internal coverage | Facade coverage | Benchmark count | Recommended API docs status | FAQ status | Comparison page status | Cookbook status |
| --- | ---: | ---: | ---: | --- | --- | ---: | --- | --- | --- | --- |
| Safe HTTP (`vhttp`, `vresty`, `vurl`) | 364 | 349 | 95.9% | `internal/httpx/http` 85.0%, `internal/httpx/resty` 80.4%, `internal/url` 87.7%, shared helpers 86.8% | `vhttp` 100.0%, `vresty` 100.0%, `vurl` 100.0% | 10 | Present in `docs/doc/README.md`, `docs/doc/22-vhttp.md`, and `docs/doc/41-vresty.md` | Present in `docs/doc/22-vhttp.md` and `docs/doc/41-vresty.md` | Present in `docs/doc/22-vhttp.md` and `docs/doc/41-vresty.md` | Present in `docs/doc/safe-http-cookbook.md` |
| Safe crypto (`vcrypto`, `vrand`, `vjwt`) | 244 | 197 | 80.7% | `internal/crypto` 94.1%, `internal/rand` 94.2%, `internal/jwt` 85.5% | `vcrypto` 100.0%, `vrand` 100.0%, `vjwt` 100.0% | 7 | Present in `docs/doc/11-vcrypto.md` and `docs/doc/38-vrand.md` | Present in `docs/doc/11-vcrypto.md` and `docs/doc/38-vrand.md` | Present in `docs/doc/safe-crypto-cookbook.md` | Present in `docs/doc/safe-crypto-cookbook.md` |
| Daily JSON/file (`vjson`, `vfile`) | 133 | 133 | 100.0% | `internal/json` 88.9%, `internal/file` 88.6% | `vjson` 100.0%, `vfile` 100.0% | 4 | Present in `docs/doc/27-vjson.md` for JSON and `docs/doc/17-vfile.md` for file workflows | Present in `docs/doc/daily-json-file-faq.md` | Present in `docs/doc/27-vjson.md` for JSON stdlib boundary; filesystem safety guidance present in `docs/doc/17-vfile.md` | Present in `docs/doc/27-vjson.md` and `docs/doc/17-vfile.md` |

## Strategic themes

1. **Scenario mindshare** — help users choose the right package for a concrete task before adding more APIs.
2. **Deep modules** — improve high-value existing packages where users already expect depth.
3. **Documentation and benchmark trust** — make claims testable through examples, benchmarks, release signals, and clear boundaries.
4. **Ecosystem adapters** — schedule AI, FTP, SSH/SFTP, pinyin, tokenization, multi-template engines, and CLI utilities as explicit development lanes.

## Recent capability closure

The Hutool gap-closure lane is implemented and the active work has moved from feature parity to governance depth. Recent commits landed the main Chinese-market and high-frequency utility gaps:

| Area | Package | Status | Evidence |
| --- | --- | --- | --- |
| SM2/SM3/SM4 national crypto | `vcrypto` | Completed | `b9f7459 feat(vcrypto): add SM2 SM3 SM4 helpers` |
| Image rotation and watermarks | `vimg` | Completed | `6447fa7 feat(vimg): add rotation and watermark helpers` |
| Consistent hashing contracts | `vhash` | Completed | `81aa58d test(vhash): harden consistent hash contracts` |
| Magic-number file type detection | `vfile` | Completed | `e72742d test(vfile): expand file type detection coverage` |
| BOM and charset contracts | `vstr` | Completed | `6e92a59 test(vstr): expand bom and charset contracts` |
| Lunar calendar, credit code, coordinates, codec, weighted random | `vdate`, `vident`, `vgeo`, `vcodec`, `vrand` | Completed | Covered by facade docs, executable examples, generated catalog entries, and focused tests |

## Capability matrix

| Area | Current package | Gap | Priority | First deliverable |
| --- | --- | --- | --- | --- |
| Collections | `vslice`, `vmap` | Error-aware transforms and typed window/zip helpers are now implemented; remaining depth is advanced partitioning and benchmark comparison narrative | P1 | Completed Sprint 10 with tests, examples, benchmarks, and generated catalogs |
| Bean mapping | `vbean`, `vmap` | copy/decode/merge semantics, deep copy, decode hooks, unused-key metadata, default merge | P1 | Define copy/decode/merge contract docs and focused tests |
| Conversion | `vconv` | explicit error-returning scalar conversions, weak-input docs, overflow-safe integer narrowing, and named-type examples are implemented | P1 | Completed Sprint 19 with contract docs, examples, tests, and generated catalogs |
| Validation | `vform`, `vident` | struct tag validation remains delegated; identity validation now includes unified social credit code parsing | P1 | Completed validation direction and credit-code helper coverage |
| Benchmarks | existing benchmark files | public benchmark narrative and competitor-neutral baselines | P1 | Add benchmark trust section and stable benchmark commands |
| Examples | all large facades | low function-level example ratio in large packages | P1 | Raise reader-facing examples in `vhttp`, `vnet`, `vnum`, `vresty`, and `vzip` |
| CLI utilities | `vcli` | command args, environment helpers, process execution, terminal IO helpers | P1 | Completed `vcli` MVP with context-aware execution |
| AI adapters | `vai` | provider abstraction for chat, embeddings, streaming, and tool calls | P1 | Completed provider-injected adapter foundation with fake-provider tests |
| Multi-template engines | `vtpl` | engine abstraction beyond `html/template` | P2 | Define adapter interface and preserve the standard template baseline |
| FTP | `vftp` | client helpers, upload/download/list, context support, provider injection | P2 | Completed provider-injected contract; remaining depth is real-client adapter examples |
| SSH/SFTP | `vssh` | SSH command execution and SFTP file transfer helpers | P2 | Completed provider-injected contract; remaining depth is host-key and transfer examples |
| Pinyin | `vhan` | Chinese transliteration helpers | P2 | Completed deterministic provider contract |
| Tokenization | `vtok` | Chinese text segmentation adapters | P2 | Completed deterministic provider contract |
| Database | `vdb` | context-first APIs, dialect depth, batch/upsert/scan helpers | P2 | Create a `vdb` deepening backlog and focused tests |
| Crypto | `vcrypto`, `vjwt`, `vrand` | Advanced safe-crypto depth is completed across SM2/SM3/SM4, weighted random, TOTP/HOTP, Argon2id password hashing, RSA JWK/JWKS, secret handling, interoperability, and benchmark scope | P2 | Closed by `safe_crypto_advanced_closeout_governance`; future work should be specific implementation lanes, not broad crypto gap closure |
| Office | `vpoi` | streaming Excel, styles, formulas, images, Word/OFD scope decision | P2 | Decide the `vpoi` scope before adding broad Office dependencies |
| Image | `vimg` | crop, resize, rotate, flip, grayscale, JPEG compression, and watermarks are implemented; remaining depth is EXIF/color profile/streaming scope | P3 | Add deterministic fixtures and benchmark notes for advanced image lanes |

## Sprint order

| Sprint | Status | Name | Outcome |
| --- | --- | --- | --- |
| 9 | Completed (`16a7e1b`) | Capability Matrix and Trust Roadmap | Published this roadmap and linked it from the documentation hub. |
| 10 | Completed (`f9101b5`) | Collection Mindshare | Deepened `vslice` and `vmap` with error-aware transforms, typed windows/pairs, examples, benchmarks, docs, and generated catalogs. |
| 11 | Completed (`72fbab1`) | Bean Copy/Decode/Merge Semantics | Split `vbean` behavior into documented copy, decode, and merge lanes. |
| 12 | Completed | Validation Direction | Recorded the validation boundary and kept struct-tag validation delegated to `go-playground/validator/v10`. |
| 13 | Completed (`4189f9e`) | Benchmark and Example Trust | Added benchmark-suite coverage, deterministic examples, benchmark-trust guidance, generated catalogs, and full governance validation. |
| 14 | Completed (`17915cf`) | Developer Experience Adapters | Implemented the `vcli` MVP with context-aware execution, flag parsing, subcommand routing, examples, benchmarks, docs, and generated catalogs. |
| 15 | Completed | Ecosystem Adapter Lane 1 | Implemented the `vai` AI adapter foundation with provider injection, context-aware chat and embedding APIs, examples, benchmarks, docs, and generated catalogs. |
| 16 | Completed | Ecosystem Adapter Lane 2 | Implemented dependency-free `vftp`, `vssh`, `vhan`, and `vtok` adapter contracts with provider injection, limits, examples, benchmarks, docs, catalogs, and governance validation. |
| 17 | Completed | Deep Business Modules | Completed `vdb`, `vcrypto`, `vpoi`, and `vimg` contract deepening with tests, examples, benchmarks, docs, generated catalogs, and governance validation. |
| 18 | Completed | Multi-Template Adapter Lane | Deepened `vtpl` with engine-neutral adapters, explicit HTML/text engine selection, context-first rendering, and contract tests before any optional third-party adapters. |
| 19 | Completed | Conversion Contract Clarity | Added explicit-error scalar conversions to `vconv`, documented weak-input semantics, added overflow-safe E integer narrowing, covered named scalar types, and kept zero/default helpers backward-compatible. |
| 20 | Completed | Hutool Gap Capability Closure | Completed national crypto, lunar calendar, credit code, coordinate conversion, codec expansion, weighted random, file type detection, BOM/charset, image operations, and consistent hash coverage across facades, docs, examples, and tests. |
| 21 | Completed | Roadmap Governance Drift Control | Kept roadmap state synchronized with generated API catalog evidence and enforced baseline plus star-domain metric drift through governance gates. |
| 22 | Completed | Large Facade Example Depth Governance | Enforced non-regression baselines for `vhttp`, `vnet`, `vnum`, `vresty`, and `vzip`; implementation passes raised `vnum` to 53, `vzip` to 68, `vnet` to 97, `vhttp` to 146, and `vresty` to 120 covered APIs. |
| 23 | Completed | Safe HTTP Cookbook Governance | Added cookbook-grade scenario guidance for `vhttp`, `vresty`, and `vurl`, then guarded the lane with generated catalog and governance evidence. |
| 24 | Completed | Safe Crypto Cookbook Governance | Added cookbook-grade scenario and comparison guidance for `vcrypto`, `vrand`, and `vjwt`, then guarded the lane with governance evidence. |
| 25 | Completed | Daily JSON/File FAQ Governance | Added cross-package FAQ guidance for `vjson` and `vfile`, then guarded the lane with governance evidence. |
| 26 | Completed | Star-Domain No-Missing Governance | Enforced that star-domain Recommended API docs, FAQ, comparison page, and cookbook status columns no longer contain `Missing` once the lanes have governance evidence. |
| 27 | Completed | vdb Deepening Backlog Governance | Added a `vdb` deepening backlog for context-first execution, dialect depth, batch/upsert behavior, scan helpers, transaction contracts, identifier safety, and benchmark scope. |
| 28 | Completed | vdb Execution Evidence Ratchet | Enforced `vdb` execution evidence for ExecBatch partial failure, Upsert dialect behavior, Tx rollback/commit errors, scan edge cases, and identifier safety. |
| 29 | Completed | vdb Example Depth Ratchet | Raised `vdb` example depth from 10 toward 20+ with reader-facing examples for execution, `ScanRows`, `ScanOne`, pagination, dialect, `WrapperForDialect`, and raw SQL boundaries. |
| 30 | Completed | Safe Crypto Advanced Backlog Governance | Defined machine-checked boundaries for TOTP/HOTP, password hashing, JWK/JWKS, secret handling, interoperability, and benchmark scope before adding more crypto APIs. |
| 31 | Completed | Safe Crypto OTP Governance | Added RFC-compatible HOTP/TOTP helpers with Base32 secrets, otpauth URLs, injected clock/window policy, RFC vectors, examples, and governance evidence. |
| 32 | Completed | Safe Crypto Password Hashing Governance | Fixed machine-checked password hashing boundaries for Argon2id-style encoded hashes, malformed-hash errors, mismatch verification, bounded test costs, and non-goals before implementation. |
| 33 | Completed | Safe Crypto Argon2id Password Hashing | Added Argon2id encoded password hashes with parameter envelopes, explicit salt source, mismatch verification, malformed-hash errors, bounded test costs, examples, and governance evidence. |
| 34 | Completed | Safe Crypto JWK/JWKS Governance | Fixed machine-checked JWK/JWKS boundaries for local key material helpers, RSA-first support, optional EC/OKP deferral, unknown-`kid` behavior, malformed-key errors, and no network discovery. |
| 35 | Completed | Safe Crypto RSA JWK/JWKS Helpers | Added local RSA JWK/JWKS key material helpers with public/private round trips, `kid` selection, malformed-key errors, no network discovery, examples, and governance evidence. |
| 36 | Completed | Safe Crypto Secret Handling Governance | Fixed machine-checked boundaries for demo secrets, deterministic fixtures, random-source injection, no production-looking fixed secrets, and secret-handling documentation. |
| 37 | Completed | Safe Crypto Interoperability Governance | Fixed machine-checked boundaries for interoperability-only helpers, SM4-ECB legacy warnings, SM2 UID policy, RSA option choices, PEM/JWK exchange, and non-default algorithm guidance. |
| 38 | Completed | Safe Crypto Benchmark Scope Governance | Fixed deterministic crypto benchmark scope for digest, HMAC, AES-GCM, AES seal/open, and secure-random smoke paths while excluding production-strength password hashing from quick gates. |
| 39 | Completed | Competitive Positioning Governance | Added utility-library comparison governance covering `samber/lo`, `duke-git/lancet`, `thoas/go-funk`, `gookit/goutil`, and `spf13/cast` with README and documentation entry points. |
| 40 | Completed | Safe Crypto Advanced Closeout | Closed the Crypto capability row around completed advanced crypto governance and enforced that all advanced backlog lanes have landed evidence. |
| 41 | Completed | Go Version Adoption Policy | Documented the Go 1.25 minimum-version rationale, CI matrix, release toolchain pin, and downgrade requirements in a machine-checked policy. |
| 42 | Completed | Collection Parity Matrix | Added collection comparison governance for `vslice`, `vmap`, `vset`, `samber/lo`, `duke-git/lancet`, and standard library `slices` / `maps` workflows. |
| 43 | Completed | vconv/vbean Migration Matrix | Added migration governance for strict conversion, weak conversion, copy, decode, merge, and unused metadata across `vconv`, `vbean`, `vconf`, and common specialist libraries. |

## Active workflow

Sprint 22 completed large-facade example-depth governance: `example_depth_governance` records non-regression baselines for `vhttp`, `vnet`, `vnum`, `vresty`, and `vzip`, keeps the roadmap Examples lane aligned with those targets, and ratcheted `vnum` examples from 23 to 53, `vzip` examples from 22 to 68, `vnet` examples from 47 to 97, `vhttp` examples from 52 to 146, and `vresty` examples from 74 to 120 in the generated catalog.

Sprint 23 completed Safe HTTP cookbook depth: `safe_http_cookbook_governance` records the governed cookbook path, required scenarios, required checks, and scorecard status for `vhttp`, `vresty`, and `vurl`.

Sprint 24 completed Safe Crypto cookbook depth: `safe_crypto_cookbook_governance` records the governed cookbook path, required scenarios, required checks, and scorecard status for `vcrypto`, `vrand`, and `vjwt`.

Sprint 25 completed Daily JSON/file FAQ depth: `daily_json_file_faq_governance` records the governed FAQ path, required questions, required checks, and scorecard status for `vjson` and `vfile`.

Sprint 26 completed star-domain no-missing status: `star_domain_no_missing_governance` keeps Recommended API docs, FAQ, comparison page, and cookbook status cells from regressing to `Missing` after the star-domain lanes have governance evidence.

Sprint 27 completed `vdb` deepening backlog governance: `vdb_deepening_governance` keeps the database lane focused on context-first execution, dialect depth, batch/upsert behavior, scan helpers, transaction contracts, identifier safety, and benchmark scope before adding more SQL helper APIs.

Sprint 28 completed `vdb` execution evidence governance: `vdb_execution_evidence_governance` locks the key behavior contracts to named tests instead of file presence alone.

Sprint 29 completed `vdb` example-depth governance: `vdb_example_depth_governance` ratchets reader-facing examples and keeps the generated tool catalog aligned with the new coverage.

Sprint 30 completed advanced safe crypto backlog governance: `safe_crypto_advanced_backlog_governance` keeps TOTP/HOTP, password hashing, JWK/JWKS, secret handling, interoperability, and benchmark scope explicit before any new public crypto APIs are added.

Sprint 31 completed safe crypto OTP governance: `safe_crypto_otp_governance` records HOTP/TOTP facade APIs, RFC vectors, deterministic clock/window tests, Base32 secret helpers, otpauth URL examples, and generated catalog coverage.

Sprint 32 completed safe crypto password hashing governance: `safe_crypto_password_hashing_governance` records the encoded-hash envelope, parameter bounds, salt source policy, mismatch behavior, malformed-hash errors, cost-bound fixtures, and non-goals before Argon2id implementation.

Sprint 33 completed safe crypto Argon2id password hashing: `safe_crypto_argon2id_governance` records the facade APIs, encoded hash round trips, mismatch behavior, malformed envelope errors, deterministic salt fixtures, generated catalog coverage, and security validation.

Sprint 34 completed safe crypto JWK/JWKS governance: `safe_crypto_jwk_jwks_governance` records local key material scope, RSA-first JWK/JWKS implementation boundaries, optional EC/OKP deferral, unknown-`kid` behavior, malformed-key errors, and no remote discovery or rotation daemon.

Sprint 35 completed safe crypto RSA JWK/JWKS helpers: `safe_crypto_jwk_jwks_implementation_governance` records RSA JWK/JWKS facade APIs, local key material round trips, `kid` selection, unknown-`kid` behavior, malformed-key errors, generated catalog coverage, and security validation.

Sprint 36 completed safe crypto secret handling governance: `safe_crypto_secret_handling_governance` records demo-secret labeling, deterministic fixture boundaries, random-source injection requirements, no production-looking fixed secrets, and documentation coverage.

Sprint 37 completed safe crypto interoperability governance: `safe_crypto_interoperability_governance` records interoperability-only helper boundaries, SM4-ECB legacy warnings, SM2 UID policy, RSA option choices, PEM/JWK key-material exchange, and non-default algorithm guidance.

Sprint 38 completed safe crypto benchmark scope governance: `safe_crypto_benchmark_scope_governance` records deterministic quick benchmark allowlists, password-hashing exclusions, and bounded runtime-evidence rules for future crypto benchmark additions.

Sprint 39 completed competitive positioning governance: `utility_library_comparison_governance` records the comparison page, README entry, competitor coverage, and boundary rules for common Go utility library choices.

Sprint 40 completed safe crypto advanced closeout: `safe_crypto_advanced_closeout_governance` keeps the Crypto capability row aligned with the landed advanced backlog lanes and prevents stale "remaining depth" wording from returning.

Sprint 41 completed Go version adoption policy: `go_version_adoption_governance` records the Go 1.25 minimum, Go 1.25.11/1.26 CI coverage, Go 1.25.11 release toolchain, and the requirements for any future downgrade proposal.

Sprint 42 completed collection parity matrix: `collections_comparison_governance` records map/filter/reduce/group/partition/window/chunk/set-like workflow boundaries across `vslice`, `vmap`, `vset`, `samber/lo`, `duke-git/lancet`, and standard library `slices` / `maps`.

Sprint 43 completed conversion and bean migration governance: `vconv_vbean_migration_governance` records strict conversion, weak conversion, copy, decode, merge, and unused metadata boundaries across `vconv`, `vbean`, `vconf`, `spf13/cast`, `jinzhu/copier`, `mitchellh/mapstructure`, and `mergo`.

Recommended roadmap loop:

1. Inventory current deep-business package APIs, examples, quickstart docs, benchmark coverage, and generated catalog entries.
2. Define the selected contract explicitly, including security defaults, sentinel error behavior, defensive copying, benchmark scope, and trusted escape hatches.
3. Add focused internal and facade tests for the selected contract before implementation changes.
4. Add deterministic `ExampleXxx` coverage for reader-facing behavior and benchmark baselines only for stable hot paths.
5. Update quickstart docs, `docs/api/exports.txt`, `docs/api/tools.json`, and `docs/api/tools.md` when public facade APIs or examples change.
6. Validate with focused package tests first, then the governance gates listed below.

## Scenario guidance

| Scenario | Use now | Planned lane |
| --- | --- | --- |
| Transform slices/maps with type-safe helpers | `vslice`, `vmap` | Error-aware transforms and window/zip helpers are available; remaining lane is advanced grouping/partitioning and benchmark comparison narrative |
| Copy, decode, or merge struct/map data | `vbean`, `vmap` | copy/decode/merge semantic split with deep-copy and metadata options |
| Validate common strings and identity formats | `vform`, `vident` | credit code validation is available; broader struct-tag validation remains delegated to `go-playground/validator` |
| Build safe HTTP clients or open untrusted URLs | `vhttp`, `vresty`, `vurl` | more examples and benchmarked helper paths |
| Build CLI or terminal utilities | `vcli` plus the standard library | deeper `venv`, `vdump`, and `vtest` lanes remain possible |
| Call AI model providers | `vai` for provider-injected contracts; provider SDKs for production transport | deeper streaming and tool-call adapter examples remain planned |
| Transfer files over FTP or SSH/SFTP | `vftp` and `vssh` for provider-injected contracts; dedicated clients for real network transfers | planned deeper connection/provider examples |
| Convert Chinese text to pinyin or tokenize text | `vhan` and `vtok` for provider-injected NLP contracts; dedicated NLP libraries for real dictionaries and segmentation | planned deeper provider examples |
| Render templates beyond `html/template` | `vtpl` for standard templates | planned multi-template adapter lane |

## Engineering constraints

- Planned lanes are not public APIs until packages are implemented, tested, documented, and added to the API snapshot.
- New public facade packages must follow the existing `internal/<domain>` plus `v<domain>` architecture.
- New public APIs must include godoc comments, deterministic tests, examples where reader-facing, API snapshot updates, and generated tool catalog updates.
- Security-sensitive lanes such as AI, FTP, SSH/SFTP, HTTP, crypto, JWT, archive, file, URL, config, random, ID, and DB helpers must use explicit errors and provider injection for tests.
- Ecosystem adapters should isolate optional dependencies behind narrow interfaces so unrelated users do not pay dependency or attack-surface costs.
- Benchmark documentation must describe baselines and commands, not claim universal performance wins.

## Validation gates for roadmap-driven work

Run focused tests first, then the standard governance gates for the touched area:

```bash
go test ./internal/<domain> ./v<domain>
UPDATE_API=1 make api-check
make docs-gen
make docs-check
make tools-check
make agent-check
make agent-security-check
```

Use `make bench-smoke` for benchmark-suite health and package-specific `go test -bench=. -benchmem -run=^$ ./<packages>` for benchmark baselines.
