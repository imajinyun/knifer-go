# vcrypto Quickstart

`vcrypto` provides common cryptographic helpers, including digests, HMAC, AES-GCM, random bytes, PBKDF2, RSA encryption/decryption/signing, SM2/SM3/SM4 national-crypto helpers, and PEM conversion.

## Golden path APIs

The first-choice API set for this facade is kept in sync with `ai-context.json` and the generated tools catalog.

- `AESDecryptGCM`
- `WithGCMRandomOptions`
- `ConstantTimeEqual`
- `AESDecryptGCMWithOptions`
- `AESEncryptGCM`

## Which helper should I use?

| Need | Use | Notes |
| --- | --- | --- |
| Hash non-secret data | SHA-256/SHA-512 helpers | Use for checksums and fingerprints. Do not use plain hashes as authentication codes. |
| Authenticate a message | HMAC helpers | Use when both sides share a secret key and the message must be tamper-evident. |
| Encrypt bytes symmetrically | AES-GCM helpers | Use authenticated encryption with a fresh nonce per key/message pair. |
| Use Chinese national cryptography | SM2/SM3/SM4 helpers | Use for interoperability with GB/T SM workflows; prefer SM4-GCM when both sides can support authenticated encryption. |
| Derive keys from passwords | `PBKDF2` | Use only for key derivation workflows where both sides already agree on KDF parameters. |
| Hash passwords for storage | `HashPasswordArgon2id`, `VerifyPasswordArgon2id`, `ParsePasswordHash` | Use encoded Argon2id hashes with explicit parameters and unique salts; this is not an account or password-storage service. |
| Generate random keys, salts, nonces, or tokens | `GenAESKey`, random byte helpers, or `vrand.SecureBytes` | Do not use pseudo-random helpers for secrets. |
| Sign or verify with RSA | RSA signing/verification helpers | Use when interoperability requires RSA keys and signatures. |
| Marshal keys | PEM conversion helpers | Treat private-key bytes as secrets and avoid logging them. |
| Exchange RSA key material as JWK/JWKS | `RSAPublicKeyToJWK`, `JWKToRSAPublicKey`, `MarshalJWKS`, `SelectJWKByKeyID` | Use for local key material parse/export and `kid` selection; network discovery and token validation policy stay outside `vcrypto`. |
| Generate or verify one-time passwords | `HOTP`, `TOTP`, `TOTPVerify`, `OTPAuthURL` | Use for RFC-compatible authenticator workflows with explicit clock, step, digits, and window policy. |

## Crypto safety checklist

- Do not use MD5 or SHA-1 for new security-sensitive designs.
- Do not reuse AES-GCM nonces with the same key.
- Do not reuse SM4-GCM nonces with the same key.
- Avoid SM4-ECB for new designs; it is provided only for interoperability with legacy systems.
- Keep SM2 signing UID policy explicit when interoperating with external systems.
- Do not generate keys, tokens, nonces, or salts with `math/rand` or deterministic test sources.
- Do not log secret bytes, private keys, raw tokens, or derived credentials.
- Do not ignore crypto errors; treat them as security-relevant failures.
- Keep keys outside source code and rotate them according to application policy.
- Separate hashing, message authentication, encryption, password derivation, and signing decisions; they solve different problems.
- Treat fixed keys, fixed salts, fixed nonces, and fixed private keys in examples as demo-only fixtures. Production code should use `vrand.SecureBytes`, `vcrypto.RandomBytes`, or injected random readers at controlled test boundaries.

## When not to use vcrypto

- Use Go's `crypto/*` packages directly when you need low-level control, streaming primitives, custom cipher modes, protocol-specific parameters, or audited interoperability code.
- Use dedicated password-hashing packages such as bcrypt, scrypt, or Argon2id for password storage instead of plain digests or generic key derivation helpers.
- Avoid legacy primitives such as MD5, SHA-1, or unauthenticated encryption for new security-sensitive designs.
- Use a key-management service, HSM, or secret manager when key generation, storage, rotation, and audit must be centrally controlled.
- Do not use encryption helpers as a substitute for authentication, authorization, replay protection, or access-control policy.

## Related packages

- Use `vrand` to generate secure bytes for keys, salts, nonces, and tokens before encoding them.
- Use `vcodec` when encrypted, signed, or hashed bytes need Base64, hex, or URL-safe representation.
- Use `vhash` only for non-cryptographic hash use cases where collision resistance is not required.

## Demo fixture policy

The quickstart and executable examples sometimes use fixed byte slices, fixed
salts, fixed nonces, or generated test keys so that examples stay deterministic.
Those values are demo fixtures, not production defaults. Production callers
should generate secrets with `vrand.SecureBytes`, `vcrypto.RandomBytes`, or
policy-controlled key-management systems. Tests may use deterministic readers,
but only at the call site under review.

## Benchmarks and trade-offs

Measure crypto helper overhead locally with the focused benchmark suite:

```bash
go test -bench=. -benchmem -run=^$ ./internal/crypto ./vcrypto ./internal/rand ./vrand
```

The suite covers SHA-256 digest, HMAC-SHA256 signing, AES-GCM encrypt/decrypt, SM3/SM4/SM2 regression paths, and secure random byte generation. Treat benchmark output as a local baseline, not a universal performance claim.

Quick benchmark gates should stay on deterministic, bounded hot paths such as
digest, HMAC, AES-GCM, AES seal/open, and secure-random smoke checks. Do not add
production-strength password hashing, remote key discovery, or key rotation work
to quick benchmark gates; run those as explicit opt-in evidence when needed.

## FAQ

### Does knifer-go replace Go's `crypto/*` standard library packages?

No. It provides focused helper entry points and documents safe defaults for common workflows. Use the standard library directly when you need low-level control.

### Which APIs should I avoid for security-sensitive code?

Avoid legacy digest algorithms such as MD5 or SHA-1 for new security-sensitive designs. Prefer documented recommended APIs.

### Are hashes the same as encryption?

No. Hashes are one-way fingerprints, HMAC authenticates messages with a shared secret, and encryption protects confidentiality. Choose the primitive that matches the threat model.

### Should I choose AES or SM4?

Use AES-GCM for general Go-to-Go or cross-platform encryption unless national-crypto interoperability is required. Use SM4 helpers when counterparties, compliance profiles, or existing payload formats require SM algorithms.

### Which helpers are interoperability-only?

Helpers for SM4-ECB, externally mandated SM2 UID policy, RSA-OAEP/PSS option
choices, and PEM/JWK key-material exchange exist for explicit interoperability
contracts. They are not the default recommendation for new designs. Prefer
authenticated encryption such as AES-GCM or SM4-GCM, keep algorithm policy at
the call site, and add tests that document the external format being matched.

### Are secrets ever logged?

Security-sensitive helpers must not log raw secrets, tokens, keys, nonces, or salts. Treat any such behavior as a security bug.

## SHA and HMAC

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vcrypto"
)

func main() {
	digest := vcrypto.SHA256Hex("hello")
	mac := vcrypto.HMACSHA256Hex([]byte("secret"), []byte("hello"))

	fmt.Println(digest)
	fmt.Println(mac)
}
```

When the hash factory is `nil`, `HMACHex` and `HMACBytes` fall back to SHA-256 instead of panicking. Prefer passing an explicit hash when interoperability matters.

## AES-GCM encryption and decryption

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vcrypto"
)

func main() {
	key, err := vcrypto.GenAESKey(32)
	if err != nil {
		panic(err)
	}

	nonce, cipherText, err := vcrypto.AESSealGCM([]byte("secret data"), key, []byte("aad"))
	if err != nil {
		panic(err)
	}

	plain, err := vcrypto.AESOpenGCM(cipherText, key, nonce, []byte("aad"))
	if err != nil {
		panic(err)
	}
	fmt.Println(string(plain))
}
```

Authentication failures from `AESOpenGCM` / `AESDecryptGCM` match `vcrypto.ErrInvalidCipherText`, so callers can distinguish tampering or wrong AAD from nonce-length validation errors.

## SM2, SM3, and SM4 helpers

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vcrypto"
)

func main() {
	digest := vcrypto.SM3Hex([]byte("abc"))

	key, err := vcrypto.GenSM4Key()
	if err != nil {
		panic(err)
	}
	nonce, cipherText, err := vcrypto.SM4SealGCM([]byte("secret data"), key, []byte("aad"))
	if err != nil {
		panic(err)
	}
	plain, err := vcrypto.SM4DecryptGCM(cipherText, key, nonce, []byte("aad"))
	if err != nil {
		panic(err)
	}

	priv, err := vcrypto.GenSM2Key()
	if err != nil {
		panic(err)
	}
	sig, err := vcrypto.SM2Sign([]byte("message"), priv)
	if err != nil {
		panic(err)
	}
	verifyErr := vcrypto.SM2Verify([]byte("message"), sig, &priv.PublicKey)

	fmt.Println(len(digest), string(plain), verifyErr == nil)
}
```

SM4-GCM authentication failures match `vcrypto.ErrInvalidCipherText`. SM2 verification failures match `vcrypto.ErrInvalidSM2Signature`. SM2 private keys are encoded as PKCS#8 PEM by `SM2PrivateKeyToPEM`; public keys use PKIX PEM.

## Derive keys with PBKDF2

```go
package main

import (
	"crypto/sha256"
	"fmt"

	"github.com/imajinyun/knifer-go/vcrypto"
)

func main() {
	key, err := vcrypto.PBKDF2([]byte("password"), []byte("salt"), 10000, 32, sha256.New)
	if err != nil {
		panic(err)
	}
	fmt.Println(len(key))
}
```

## Hash and verify passwords with Argon2id

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vcrypto"
)

func main() {
	encoded, err := vcrypto.HashPasswordArgon2id([]byte("correct horse battery staple"))
	if err != nil {
		panic(err)
	}

	ok, err := vcrypto.VerifyPasswordArgon2id(encoded, []byte("correct horse battery staple"))
	if err != nil {
		panic(err)
	}

	info, err := vcrypto.ParsePasswordHash(encoded)
	if err != nil {
		panic(err)
	}

	fmt.Println(ok, info.Algorithm)
}
```

Use option overrides such as `WithArgon2idMemory`, `WithArgon2idIterations`, and `WithPasswordHashRandomOptions` only when the deployment policy or deterministic tests require explicit parameters. A password mismatch returns `false, nil`; malformed encoded hashes return errors matching `vcrypto.ErrInvalidPasswordHash`.

## Export and select RSA keys with JWK/JWKS

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vcrypto"
)

func main() {
	priv, err := vcrypto.GenRSAKey(2048)
	if err != nil {
		panic(err)
	}

	jwk, err := vcrypto.RSAPublicKeyToJWK(&priv.PublicKey, "kid-1")
	if err != nil {
		panic(err)
	}

	data, err := vcrypto.MarshalJWKS([]vcrypto.JWK{jwk})
	if err != nil {
		panic(err)
	}

	set, err := vcrypto.ParseJWKS(data)
	if err != nil {
		panic(err)
	}
	selected, err := vcrypto.SelectJWKByKeyID(set, "kid-1")
	if err != nil {
		panic(err)
	}
	fmt.Println(selected.KeyID)
}
```

`vcrypto` only handles local JWK/JWKS key material. It does not fetch remote JWKS URLs, refresh key sets, rotate keys, or decide whether a JWT should be trusted; those policies stay at the application or identity-provider boundary.

## Generate and verify one-time passwords

```go
package main

import (
	"fmt"
	"time"

	"github.com/imajinyun/knifer-go/vcrypto"
)

func main() {
	secret := []byte("12345678901234567890")
	now := time.Unix(59, 0).UTC()

	code, err := vcrypto.TOTP(secret, now, vcrypto.WithOTPDigits(8))
	if err != nil {
		panic(err)
	}

	ok, err := vcrypto.TOTPVerify(code, secret, now, vcrypto.WithOTPDigits(8))
	if err != nil {
		panic(err)
	}
	fmt.Println(ok, code)
}
```

Use `WithOTPClock` in tests and `WithTOTPWindow` only when the login policy intentionally accepts adjacent time steps. `OTPAuthURL` formats an `otpauth://` URL for provisioning authenticators; it does not implement account enrollment, MFA recovery, or identity-provider policy.

## RSA signing and verification

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vcrypto"
)

func main() {
	priv, err := vcrypto.GenRSAKey(2048)
	if err != nil {
		panic(err)
	}

	data := []byte("message")
	sig, err := vcrypto.SignSHA256WithRSA(data, priv)
	if err != nil {
		panic(err)
	}

	err = vcrypto.VerifySHA256WithRSA(data, sig, &priv.PublicKey)
	fmt.Println(err == nil)
}
```
