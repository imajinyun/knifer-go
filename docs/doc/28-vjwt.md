# vjwt Quickstart

`vjwt` provides JWT creation, parsing, signature verification, date-claim validation, and multiple signers including HMAC, RSA-PSS, and ECDSA.

## Golden path APIs

The first-choice API set for this facade is kept in sync with `ai-context.json` and the generated tools catalog.

- `CreateJWTToken`
- `ES256WithOptions`
- `AlgorithmName`
- `MustHMACSigner`
- `CreateJWTTokenWithSigner`

## Which helper should I use?

Choose helpers by the trust boundary: token creation, signature verification, claim validation, or key/algorithm selection.

| Need | Use | Notes |
| --- | --- | --- |
| Create a simple HMAC token | `CreateToken` | Good for trusted internal tokens when the shared secret is managed outside source code. |
| Create a token with explicit headers, payload, algorithm, and key | `CreateTokenWithOptions` | Prefer this when `kid`, issuer, audience, or algorithm policy must be visible at the call site. |
| Use asymmetric signing | `CreateTokenWithSigner`, `PS256`, ECDSA/RSA signer helpers | Prefer asymmetric signers when verifiers should not hold signing keys. |
| Parse token structure without trusting it | `ParseToken`, `JWTOf` | Parsing exposes headers and claims; it does not by itself make the token trustworthy. |
| Verify a token signature | `Verify`, `VerifyWithSigner` | Verify before authorizing a request or trusting claims. |
| Validate time-based claims | `ValidateWithOptions`, `WithValidateTime`, `WithValidateLeeway` | Use deterministic validation time in tests and small leeway for clock skew. |

## JWT safety checklist

- Verify signatures before trusting any header or payload claim for authorization decisions.
- Keep accepted algorithms explicit. Do not let untrusted token headers silently choose an unexpected signing method.
- Validate `exp`, `nbf`, and `iat` where applicable, and keep clock-skew leeway small and documented.
- Validate application claims such as `iss`, `aud`, `sub`, tenant, scope, and key id against your own policy.
- Treat JWT payloads as readable metadata, not encrypted secrets. Do not store credentials or sensitive personal data in plain JWT claims.
- Rotate and scope signing keys outside source code. Prefer asymmetric signers when many services only need verification.

## When not to use vjwt

- Use an opaque server-side session token when claims should not be readable by clients or when immediate revocation is required.
- Use a full identity provider or OAuth/OIDC library when you need discovery documents, JWKS rotation, authorization-code flows, refresh tokens, or token introspection.
- Use `CreateSignerStrict` or key-management code outside this facade when HMAC key length, rotation, and policy enforcement need to be centralized.
- Do not use JWT payloads as encrypted storage. Add encryption separately or keep sensitive state server-side.
- Do not accept tokens from unknown issuers just because the signature verifies; application claim policy still belongs at the boundary.

## Must API compatibility

`MustHMACSigner` is a compatibility helper for trusted startup constants and tests where an invalid algorithm should fail fast. New code should prefer `vjwt.NewHMACSigner` when it needs ordinary algorithm validation errors, or `vjwt.NewHMACSignerStrict` when HMAC key-length policy must be enforced at construction time.

## Related packages

- Use `vcrypto` when JWT keys, signatures, or hashing need lower-level cryptographic helpers.
- Use `vrand` when generating secure token IDs, nonces, or secret material.
- Use `vjson` when custom claims need JSON fixture generation or inspection in tests.

## Benchmarks and trade-offs

Measure signing, parsing, verification, and date validation with the JWT benchmarks before choosing an algorithm for a hot path:

```bash
go test -bench=. -benchmem -run=^$ ./internal/jwt ./vjwt
```

HMAC signers are usually cheaper and easier to operate for a small trusted service set, but every verifier that can validate can also sign. RSA-PSS and ECDSA separate signing from verification, at the cost of larger keys, more CPU, and more key-distribution work.

Validation options such as deterministic clocks and leeway improve correctness and testability. Keep the leeway small: broad windows make replay and stale-token behavior harder to reason about.

## FAQ

### Is a JWT encrypted?

No. A signed JWT protects integrity, not confidentiality. Anyone who can read the token can decode its header and payload unless you use a separate encryption scheme.

### How should I choose HMAC vs RSA/ECDSA signers?

Use HMAC when a small trusted set of services can safely share one secret. Use asymmetric signers when signing and verification responsibilities should be separated, such as many verifiers and one issuer.

### Is parsing enough before using claims?

No. Parse, verify the signature, then validate time and application claims before authorizing access.

## Create and verify tokens with HMAC

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vjwt"
)

func main() {
	key := []byte("secret")
	token, err := vjwt.CreateToken(map[string]any{vjwt.JWTPayloadSubject: "alice"}, key)
	if err != nil {
		panic(err)
	}

	parsed, err := vjwt.ParseToken(token)
	if err != nil {
		panic(err)
	}
	fmt.Println(parsed.Payload(vjwt.JWTPayloadSubject))
	fmt.Println(vjwt.Verify(token, key))
}
```

## Set headers, payloads, and algorithms with options

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vjwt"
)

func main() {
	token, err := vjwt.CreateTokenWithOptions(
		vjwt.WithTokenHeaders(map[string]any{vjwt.JWTHeaderKeyID: "key-1"}),
		vjwt.WithTokenPayload(map[string]any{vjwt.JWTPayloadIssuer: "issuer", vjwt.JWTPayloadSubject: "alice"}),
		vjwt.WithTokenAlgorithm(vjwt.JWTAlgHS384),
		vjwt.WithTokenKey([]byte("secret")),
	)
	if err != nil {
		panic(err)
	}

	parsed, err := vjwt.JWTOf(token)
	if err != nil {
		panic(err)
	}
	fmt.Println(parsed.Header(vjwt.JWTHeaderKeyID), parsed.Algorithm())
}
```

## Build JWTs fluently and validate date claims

```go
package main

import (
	"fmt"
	"time"

	"github.com/imajinyun/knifer-go/vjwt"
)

func main() {
	now := time.Now()
	j := vjwt.New().
		SetKey([]byte("secret")).
		SetIssuer("knifer-go").
		SetSubject("alice").
		SetIssuedAt(now).
		SetNotBefore(now.Add(-time.Minute)).
		SetExpiresAt(now.Add(time.Hour))

	token, err := j.Sign()
	if err != nil {
		panic(err)
	}

	parsed, err := vjwt.ParseToken(token)
	if err != nil {
		panic(err)
	}
	fmt.Println(parsed.ValidateWithOptions(vjwt.WithValidateTime(now), vjwt.WithValidateLeeway(30)))
}
```

## Use an RSA-PSS signer

```go
package main

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"

	"github.com/imajinyun/knifer-go/vjwt"
)

func main() {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	signer := vjwt.PS256(priv, &priv.PublicKey)

	token, err := vjwt.CreateTokenWithSigner(map[string]any{vjwt.JWTPayloadSubject: "alice"}, signer)
	if err != nil {
		panic(err)
	}

	fmt.Println(vjwt.VerifyWithSigner(token, signer))
}
```
