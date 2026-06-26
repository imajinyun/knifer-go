# vrand Quickstart

`vrand` provides random integer, float, bool, string, slice element, and byte generators, with support for per-call random source injection.

## Security boundary

Use `SecureBytes` or `SecureBytesWithOptions` for tokens, keys, salts, and nonces. Encode those bytes with `hex`, `base64`, or an application-specific alphabet when a string representation is required. The string and number helpers are pseudo-random convenience APIs and must not be used for secrets.

## Which helper should I use?

Choose helpers by whether the value is security-sensitive or only needs pseudo-random convenience behavior.

| Need | Use | Notes |
| --- | --- | --- |
| Secret bytes, tokens, keys, salts, or nonces | `SecureBytes`, `SecureBytesWithOptions` | Encode bytes separately when a string representation is required. |
| Pseudo-random integers, ranges, floats, or booleans | `Int`, `IntRange`, `Long`, `Float`, `Bool` | Good for simulations, sampling, UI variation, and tests; not for secrets. |
| Pseudo-random strings from built-in alphabets | `String`, `Numbers`, `StringUpper` | Good for non-secret placeholders and examples. |
| Pseudo-random strings from a custom alphabet | `StringFrom` | Ensure the alphabet is non-empty and suitable for the downstream format. |
| Select an element from a slice | `Ele` | Decide how empty slices should be handled before calling. |
| Deterministic tests | `WithRandomSource` | Inject `math/rand` sources only for reproducible non-security helpers. |

## Randomness safety checklist

- Use secure byte helpers for anything that authenticates, authorizes, encrypts, salts, or protects access.
- Do not turn pseudo-random strings or numbers into passwords, reset tokens, API keys, nonces, or session identifiers.
- Encode secure bytes with `hex`, Base64, or URL-safe Base64 rather than reducing entropy with a small ad-hoc alphabet.
- Keep deterministic random sources limited to examples and tests; never inject them into secret generation paths.
- Check secure-random errors and fail closed if the operating system randomness source is unavailable.
- Consider entropy length explicitly: 16 bytes can be enough for many identifiers, while long-lived keys or tokens often need 32 bytes or more.

## Related packages

- Use `vcrypto` when random bytes feed keys, salts, nonces, signatures, or encryption workflows.
- Use `vcodec` when random bytes need hex, Base64, or URL-safe representation.
- Use `vid` when random material should become application identifiers rather than raw bytes.

## When not to use vrand

- Use `SecureBytes` rather than pseudo-random string or number helpers for secrets, credentials, bearer tokens, nonces, salts, and keys.
- Use a domain-specific statistical or simulation library when the distribution, seeding, reproducibility, or sampling method must be controlled precisely.
- Use `crypto/rand` or cryptographic protocols directly when interoperability requires exact byte generation or encoding behavior.
- Avoid deterministic `WithRandomSource` outside tests, examples, simulations, and non-security reproducibility workflows.
- Avoid shrinking secure random bytes into a small custom alphabet unless the entropy budget has been reviewed.

## Benchmarks and trade-offs

Measure secure random generation locally with the focused benchmark suite:

```bash
go test -bench=. -benchmem -run=^$ ./vrand ./internal/rand
```

The suite covers secure byte generation. Treat benchmark output as a local baseline rather than a universal performance claim. For non-security helpers, benchmark the specific distribution and source injection you plan to use.

## FAQ

### Can I use String or Numbers for reset tokens?

No. Use `SecureBytes` and encode the result. The string and number helpers are pseudo-random convenience APIs and are not designed for secrets.

### Why does WithRandomSource accept math/rand?

It exists for deterministic tests, examples, simulations, and compatibility behavior. It must not be used to generate keys, tokens, nonces, salts, or credentials.

### How should I produce a URL-safe secret token?

Generate secure bytes first, then encode them with URL-safe Base64 or another audited encoding. Do not build a secret token by repeatedly choosing pseudo-random characters.

## Generate numbers and booleans

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vrand"
)

func main() {
	fmt.Println(vrand.Int(10))
	fmt.Println(vrand.IntRange(10, 20))
	fmt.Println(vrand.Long())
	fmt.Println(vrand.Float())
	fmt.Println(vrand.Bool())
}
```

## Generate random strings

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vrand"
)

func main() {
	fmt.Println(vrand.String(8))
	fmt.Println(vrand.Numbers(6))
	fmt.Println(vrand.StringUpper(8))
	fmt.Println(vrand.StringFrom("ABC", 4))
}
```

## Choose random elements from slices

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vrand"
)

func main() {
	items := []string{"go", "knifer", "tool"}
	fmt.Println(vrand.Ele(items))
}
```

## Generate secure random bytes and reproducible pseudo-random results

Use `SecureBytes` for secrets, tokens, keys, and nonces. `WithRandomSource`
accepts `math/rand` only for reproducible non-security helpers such as examples,
tests, random selection, and compatibility fallback behavior.

```go
package main

import (
	"fmt"
	mathrand "math/rand"

	"github.com/imajinyun/knifer-go/vrand"
)

func main() {
	b, err := vrand.SecureBytes(16)
	if err != nil {
		panic(err)
	}
	fmt.Println(len(b))

	source := mathrand.New(mathrand.NewSource(1))
	fmt.Println(vrand.IntWithOptions(100, vrand.WithRandomSource(source)))
}
```
