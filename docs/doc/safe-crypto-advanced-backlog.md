# Safe Crypto Advanced Backlog

`vcrypto`, `vjwt`, `vrand`, and `vpass` already cover digest, HMAC, AES-GCM, RSA, SM2/SM3/SM4, secure random material, JWT signing, and password-strength analysis. This backlog defines the next crypto depth lanes without turning the package set into an identity provider, password vault, or key-management service.

## Scope

| Lane | Current baseline | Next hardening target |
| --- | --- | --- |
| TOTP and HOTP | RFC-compatible helpers are available with deterministic clock/counter injection, explicit issuer/account formatting, and constant-time verification. | Keep window policy, Base32 secret handling, and provisioning URL behavior covered by named tests and examples. |
| Password hashing | Governance is fixed for Argon2id-style encoded password hashes, malformed-hash errors, mismatch verification, bounded test costs, and non-goals. | Implement helpers only after the encoded envelope, parameter bounds, salt source, and verification semantics are wired to named tests. |
| JWK and JWKS | Governance is fixed for local JWK/JWKS key material helpers, `kid` selection, malformed-key errors, no network discovery, and no key rotation daemon. | Implement parsing or publishing only after RSA/EC/OKP support boundaries and unknown-`kid` behavior are wired to named tests. |
| Secret handling | Governance is fixed for separating deterministic fixtures from production secret handling, documenting fake/demo secrets, and requiring random-source injection for tests. | Keep salts, nonces, OTP secrets, private keys, and encoded password hashes out of examples that look production-ready with fixed secrets. |
| Interoperability boundaries | Governance is fixed for legacy or externally mandated algorithms, including SM4-ECB, RSA-OAEP/PSS option choices, SM2 UID policy, and PEM/JWK key-material exchange. | Keep interoperability-only helpers explicit, documented, and covered by tests before adding more algorithms or modes. |
| Benchmark scope | Existing crypto benchmarks cover stable digest, HMAC, and authenticated-encryption paths. | Benchmark deterministic hot paths only; do not benchmark password hashing with production-strength cost in quick gates. |

## Non-Goals

- No OAuth or OIDC provider implementation.
- No password storage service.
- No account lifecycle or reset flow.
- No breached-password corpus check.
- No MFA or recovery policy.
- No key management service.
- No custom cryptographic primitive.
- No long-lived secret registry or process-global key rotation state.
- No remote JWKS discovery, cache, refresh, or rotation daemon.
- No OAuth or OIDC discovery.
- No remote JWKS fetch.
- No JWKS cache or refresh loop.
- No key rotation daemon.
- No token validation policy inside vcrypto.

## Required Evidence

- TOTP and HOTP work must include fixed RFC vectors, injected clock or counter sources, window verification tests, and invalid-secret/error-contract tests.
- Password hashing work must include encoded-parameter round trips, mismatch verification, malformed-hash errors, and cost-bound test fixtures.
- JWK and JWKS work must include RSA/EC/OKP key fixtures where supported, unknown-kid behavior, malformed-key errors, and no network discovery.
- Every lane must update facade examples, `docs/api/tools.json`, `docs/api/tools.md`, and `ai-context.json` before it is considered closed.

## Landed Evidence

| Lane | Status | Evidence |
| --- | --- | --- |
| TOTP and HOTP | Completed | `safe_crypto_otp_governance` records RFC vectors, clock/window tests, invalid-input tests, facade examples, generated catalog coverage, and Sprint 31 roadmap state. |
| Password hashing | Completed | `safe_crypto_password_hashing_governance` records Argon2id-style parameter envelopes and non-goals; `safe_crypto_argon2id_governance` records the encoded hash implementation, mismatch behavior, malformed-hash errors, bounded-cost tests, facade examples, and Sprint 33 roadmap state. |
| JWK and JWKS | Completed | `safe_crypto_jwk_jwks_governance` records local key material scope and non-goals; `safe_crypto_jwk_jwks_implementation_governance` records RSA public/private JWK round trips, JWKS `kid` selection, unknown-`kid` behavior, malformed-key errors, no network discovery, facade examples, and Sprint 35 roadmap state. |
| Secret handling | Governance completed | `safe_crypto_secret_handling_governance` records demo-secret labeling, deterministic fixture boundaries, random-source injection requirements, no production-looking fixed secrets, and Sprint 36 roadmap state. |
| Interoperability boundaries | Governance completed | `safe_crypto_interoperability_governance` records explicit interoperability-only helpers, legacy-mode warnings, SM4-ECB non-default guidance, SM2 UID policy, RSA option boundaries, PEM/JWK key-material exchange, and Sprint 37 roadmap state. |

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
