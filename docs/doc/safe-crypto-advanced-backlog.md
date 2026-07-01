# Safe Crypto Advanced Backlog

`vcrypto`, `vjwt`, `vrand`, and `vpass` already cover digest, HMAC, AES-GCM, RSA, SM2/SM3/SM4, secure random material, JWT signing, and password-strength analysis. This backlog defines the next crypto depth lanes without turning the package set into an identity provider, password vault, or key-management service.

## Scope

| Lane | Current baseline | Next hardening target |
| --- | --- | --- |
| TOTP and HOTP | RFC-compatible helpers are available with deterministic clock/counter injection, explicit issuer/account formatting, and constant-time verification. | Keep window policy, Base32 secret handling, and provisioning URL behavior covered by named tests and examples. |
| Password hashing | `vcrypto` has PBKDF2 key derivation and `vpass` has strength analysis, but no password storage API. | Prefer Argon2id or bcrypt-style password hashing through explicit, versioned parameter structs and encoded hash envelopes. |
| JWK and JWKS | `vjwt` signs and validates tokens; key discovery and rotation stay external. | Add JWK/JWKS parsing or publishing only as key material helpers, not OAuth/OIDC discovery. |
| Secret handling | `vrand` and `vcrypto` expose secure random bytes and option-injected readers. | Keep salts, nonces, OTP secrets, private keys, and encoded password hashes out of examples that look production-ready with fixed secrets. |
| Interoperability boundaries | SM and RSA helpers already expose interoperability-focused APIs. | Put legacy, optional, or externally mandated algorithms behind explicit names and docs that state why they exist. |
| Benchmark scope | Existing crypto benchmarks cover stable digest, HMAC, and authenticated-encryption paths. | Benchmark deterministic hot paths only; do not benchmark password hashing with production-strength cost in quick gates. |

## Non-Goals

- No OAuth or OIDC provider implementation.
- No password storage service.
- No key management service.
- No custom cryptographic primitive.
- No long-lived secret registry or process-global key rotation state.

## Required Evidence

- TOTP and HOTP work must include fixed RFC vectors, injected clock or counter sources, window verification tests, and invalid-secret/error-contract tests.
- Password hashing work must include encoded-parameter round trips, mismatch verification, malformed-hash errors, and cost-bound test fixtures.
- JWK and JWKS work must include RSA/EC/OKP key fixtures where supported, unknown-kid behavior, malformed-key errors, and no network discovery.
- Every lane must update facade examples, `docs/api/tools.json`, `docs/api/tools.md`, and `ai-context.json` before it is considered closed.

## Landed Evidence

| Lane | Status | Evidence |
| --- | --- | --- |
| TOTP and HOTP | Completed | `safe_crypto_otp_governance` records RFC vectors, clock/window tests, invalid-input tests, facade examples, generated catalog coverage, and Sprint 31 roadmap state. |

## Validation

Run focused crypto tests before changing crypto behavior:

```bash
go test ./internal/crypto ./vcrypto ./internal/jwt ./vjwt ./internal/rand ./vrand ./internal/pass ./vpass
```

Run governance and security gates after docs, examples, metadata, or public API changes:

```bash
make docs-check
make ai-context-check
make governance-maturity-check
make agent-security-check
```
