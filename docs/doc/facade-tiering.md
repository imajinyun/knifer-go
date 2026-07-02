# Facade Tiering and Import Guide

Use this page when deciding which `v*` package belongs in new application code.
The tier inventory mirrors `ai-context.json` so humans and AI agents use the
same import boundaries as the governance checks.

Machine-readable sources: `dependency_tiers` and `security_sensitive_packages`.

## Day-One Defaults

Start here when a task maps directly to one of these common workflows.

| Task | Default facade | Related facades |
| --- | --- | --- |
| string cleanup | `vstr` | `vregex`, `vdfa` |
| slice transformation | `vslice` | `vmap`, `vset` |
| map transformation | `vmap` | `vslice`, `vset` |
| JSON path and formatting | `vjson` | `vxml`, `vfile` |
| file IO | `vfile` | `vzip`, `vurl` |
| safe HTTP | `vhttp` | `vresty`, `vurl`, `vnet` |
| crypto | `vcrypto` | `vrand`, `vjwt`, `vpass` |
| configuration | `vconf` | `vbean`, `vconv` |
| database | `vdb` | `vconf`, `vcli` |
| CLI command execution | `vcli` | `vsys`, `vlog` |

## Dependency Tiers

| Tier | Facades | Import rule |
| --- | --- | --- |
| core facades | `vbean`, `vblf`, `vbool`, `vcache`, `vcli`, `vcodec`, `vconf`, `vconv`, `vcron`, `vcrypto`, `vcsv`, `vdate`, `vdb`, `vdfa`, `vfile`, `vform`, `vgeo`, `vhash`, `vhttp`, `vid`, `vident`, `vjob`, `vjson`, `vjwt`, `vlog`, `vmail`, `vmap`, `vmask`, `vnet`, `vnum`, `vobj`, `vpass`, `vrand`, `vref`, `vregex`, `vsem`, `vset`, `vskt`, `vslice`, `vstr`, `vsys`, `vtpl`, `vurl`, `vver`, `vxml`, `vzip` | Standard-library-first; third-party imports require explicit allowlist review. |
| heavy extension facades | `verr`, `vimg`, `vpoi`, `vresty` | Optional integrations stay inside their owning facade and matching `internal/*` package family. |
| provider contract facades | `vai`, `vftp`, `vhan`, `vssh`, `vtok` | Public APIs expose provider interfaces and call contracts; concrete clients, credentials, dictionaries, and NLP engines stay outside core. |

## Security-Sensitive Overlay

Security-sensitive facades are not a separate dependency tier. They are packages
where untrusted input, secrets, network boundaries, filesystem boundaries, SQL,
or command execution can affect safety. Prefer Safe, E, context-aware, or
WithOptions flows in these packages.

| Category | Facades |
| --- | --- |
| Network and URL boundaries | `vhttp`, `vresty`, `vurl`, `vnet` |
| File, archive, and config boundaries | `vfile`, `vzip`, `vconf` |
| Crypto, token, random, and identity boundaries | `vcrypto`, `vjwt`, `vrand`, `vid` |
| SQL and command boundaries | `vdb`, `vcli` |
| Provider contract boundaries | `vai`, `vftp`, `vssh` |

## Import Rules

- Import public `v*` facade packages, not `internal/*`.
- Choose one default facade first.
- Use related facades only when the workflow crosses package boundaries.
- Use heavy extension facades only when their dependency or integration is
  already part of the application decision.
- Keep provider contract facades provider-neutral; concrete clients belong in
  application code or separate adapters.
- Treat the security-sensitive overlay as a review signal, not as a reason to
  avoid the facade.
- Run `make arch`, `make ai-context-check`, and `make governance-maturity-check`
  after changing tier metadata or facade imports.

## Machine-Readable Boundaries

- day-one defaults
- core facades
- heavy extension facades
- provider contract facades
- security-sensitive overlay
- import public v* facade packages
- do not import internal packages
- heavy dependencies require allowlist review
- provider contracts stay provider-neutral
- Safe/E/WithOptions flows in security-sensitive facades
