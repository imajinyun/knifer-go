# Security Policy

## Supported versions

`go-knifer` follows semantic versioning. Security fixes are provided for the
latest minor release on the current major version.

Weakening a documented security default in a public `v*` facade is treated as a
breaking change. Security deprecations must name the safer replacement, stay
available for at least two minor releases when safe to do so, and be recorded in
release notes before removal.

| Version | Supported |
| --- | --- |
| `v0.x` latest | Yes |
| Older `v0.x` releases | Best effort |

## Reporting a vulnerability

Please do not open a public issue for suspected vulnerabilities.

Report security issues by sending a private advisory through GitHub Security
Advisories, or by contacting a maintainer through a private channel listed on
the repository profile.

Include:

- Affected package and version.
- Minimal reproduction code or input.
- Expected behavior and observed behavior.
- Impact assessment, especially for SSRF, path traversal, cryptography, JWT,
  archive extraction, configuration loading, file IO, or network helpers.

## Response process

The maintainers aim to:

- Acknowledge the report within 3 business days.
- Confirm affected versions and impact before publishing details.
- Prepare a fix, regression test, and release notes entry.
- Publish a GitHub Security Advisory when the issue is confirmed.

## Security-sensitive areas

Changes touching these packages require extra review:

- `vhttp`, `vresty`, `vurl`, `vconf`: SSRF, redirects, remote reads, TLS, and
  request construction.
- `vzip`, `vfile`: path traversal, symlink escape, file permissions, and
  decompression limits.
- `vcrypto`, `vjwt`, `vrand`, `vid`: cryptography, token verification,
  randomness, signatures, and key handling.
- `vdb`: SQL construction, named arguments, transactions, and resource cleanup.
- `vcli`: external command execution, argument separation, shell boundaries,
  timeout policy, and output limits.
- `vai`: provider requests, prompt or embedding payload sensitivity, credential
  handling, redaction-safe diagnostics, and defensive request copying.
- `vftp`, `vssh`: remote command or transfer boundaries, credentials, SFTP/FTP
  path handling, transfer limits, and output-size limits.

Security linter suppressions in `.golangci.yml` must stay narrow and documented.
Prefer adding a regression test over broadening a suppression. New `#nosec` or
`//nolint:gosec` comments must name the specific operation, explain why it is
safe for that trust boundary, and stay next to the suppressed line.

## Secret handling

Security-sensitive helpers must not log raw secrets, private keys, tokens,
nonces, salts, or derived credentials. Use `vrand.SecureBytes` or
`vcrypto.RandomBytes` for keys, tokens, salts, and nonces, and treat unexpected
crypto errors as fail-closed security failures rather than recoverable defaults.
