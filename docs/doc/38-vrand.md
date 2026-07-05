# vrand Quickstart

`vrand` provides random integer, float, bool, string, slice element, and byte generators, with support for per-call random source injection.

## Security boundary

Use `SecureBytes` or `SecureBytesWithOptions` for tokens, keys, salts, and nonces. Encode those bytes with `hex`, `base64`, or an application-specific alphabet when a string representation is required. The string and number helpers are pseudo-random convenience APIs and must not be used for secrets.

## Golden path APIs

The first-choice API set for this facade is kept in sync with `ai-context.json` and the generated tools catalog.

- `BytesWithOptions`
- `BoolWithOptions`
- `Bool`
- `WithRandomSource`
- `ConfigureDefaultRandomSourceProvider`

## Which helper should I use?

Choose helpers by whether the value is security-sensitive or only needs pseudo-random convenience behavior.

| Need | Use | Notes |
| --- | --- | --- |
| Secret bytes, tokens, keys, salts, or nonces | `SecureBytes`, `SecureBytesWithOptions` | Encode bytes separately when a string representation is required. |
| Pseudo-random integers, ranges, floats, or booleans | `Int`, `IntRange`, `Long`, `Float`, `Bool` | Good for simulations, sampling, UI variation, and tests; not for secrets. |
| Pseudo-random strings from built-in alphabets | `String`, `Numbers`, `StringUpper` | Good for non-secret placeholders and examples. |
| Pseudo-random strings from a custom alphabet | `StringFrom` | Ensure the alphabet is non-empty and suitable for the downstream format. |
| Select an element from a slice | `Ele` | Decide how empty slices should be handled before calling. |
| Weighted pseudo-random selection | `WeightedPick`, `WeightedPickN`, `WeightedPickUniqueN` | Use for simulations, scheduling, experiments, and sampling where weights are non-secret business inputs. |
| Deterministic tests | `WithRandomSource` | Inject `math/rand` sources only for reproducible non-security helpers. |

## Randomness safety checklist

- Use secure byte helpers for anything that authenticates, authorizes, encrypts, salts, or protects access.
- Do not turn pseudo-random strings or numbers into passwords, reset tokens, API keys, nonces, or session identifiers.
- Encode secure bytes with `hex`, Base64, or URL-safe Base64 rather than reducing entropy with a small ad-hoc alphabet.
- Keep deterministic random sources limited to examples and tests; never inject them into secret generation paths.
- Check secure-random errors and fail closed if the operating system randomness source is unavailable.
- Consider entropy length explicitly: 16 bytes can be enough for many identifiers, while long-lived keys or tokens often need 32 bytes or more.
- Treat weighted helpers as pseudo-random sampling APIs. They are not a way to generate secrets, hide probabilities, or make security decisions.
- Validate weighted inputs: item and weight slices must have matching lengths, weights must be finite and non-negative, and the total weight must be positive.

## Related packages

- Use `vcrypto` when random bytes feed keys, salts, nonces, signatures, or encryption workflows.
- Use `vcodec` when random bytes need hex, Base64, or URL-safe representation.
- Use `vid` when random material should become application identifiers rather than raw bytes.

## When not to use vrand

- Use `SecureBytes` rather than pseudo-random string or number helpers for secrets, credentials, bearer tokens, nonces, salts, and keys.
- Use a domain-specific statistical or simulation library when the distribution, seeding, reproducibility, or sampling method must be controlled precisely.
- Use `crypto/rand` or cryptographic protocols directly when interoperability requires exact byte generation or encoding behavior.
- Avoid deterministic `WithRandomSource` outside tests, examples, simulations, and non-security reproducibility workflows.
- Avoid `WeightedPick*` for lotteries, security controls, access decisions, or regulated randomness unless your domain requirements explicitly accept pseudo-random sampling and the implementation has been reviewed for that use.
- Avoid shrinking secure random bytes into a small custom alphabet unless the entropy budget has been reviewed.

## Benchmarks and trade-offs

Measure secure random generation locally with the focused benchmark suite:

```bash
go test -bench=. -benchmem -run=^$ ./vrand ./internal/rand
```

The suite covers secure byte generation and weighted selection paths. Treat benchmark output as a local baseline rather than a universal performance claim. For non-security helpers, benchmark the specific distribution and source injection you plan to use.

## FAQ

### Can I use String or Numbers for reset tokens?

No. Use `SecureBytes` and encode the result. The string and number helpers are pseudo-random convenience APIs and are not designed for secrets.

### Why does WithRandomSource accept math/rand?

It exists for deterministic tests, examples, simulations, and compatibility behavior. It must not be used to generate keys, tokens, nonces, salts, or credentials.

### How should I produce a URL-safe secret token?

Generate secure bytes first, then encode them with URL-safe Base64 or another audited encoding. Do not build a secret token by repeatedly choosing pseudo-random characters.

### Can I use weighted helpers for security-sensitive selection?

No. Weighted helpers use pseudo-random selection semantics and are intended for sampling, simulations, routing, experiments, and scheduling. Use a reviewed cryptographic or domain-specific mechanism when the selection affects authentication, authorization, money movement, or other security-sensitive decisions.

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

## Choose weighted random elements

Use weighted helpers when business inputs define relative likelihood. `WeightedPick`
returns one item, `WeightedPickN` samples with replacement, and
`WeightedPickUniqueN` samples without replacement. Inject a weighted random
source only for deterministic tests and examples.

```go
package main

import (
	"fmt"
	mathrand "math/rand"
	"strings"

	"github.com/imajinyun/knifer-go/vrand"
)

func main() {
	source := mathrand.New(mathrand.NewSource(1))
	item, err := vrand.WeightedPick(
		[]string{"cold", "hot"},
		[]float64{0, 10},
		vrand.WithWeightedRandSource(source),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(item)
}
```

When very small floating-point weights should be treated as absent, configure a
positive precision threshold and handle the resulting validation error.

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vrand"
)

func main() {
	_, err := vrand.WeightedPick(
		[]string{"tiny"},
		[]float64{1e-15},
		vrand.WithWeightedPrecision(1e-12),
	)
	fmt.Println(err != nil)
}
```

## Sample multiple weighted values

```go
package main

import (
	"fmt"
	mathrand "math/rand"

	"github.com/imajinyun/knifer-go/vrand"
)

func main() {
	withReplacement, err := vrand.WeightedPickN(
		[]string{"a", "b", "c"},
		[]float64{0, 1, 0},
		3,
		vrand.WithWeightedRandSource(mathrand.New(mathrand.NewSource(1))),
	)
	if err != nil {
		panic(err)
	}

	unique, err := vrand.WeightedPickUniqueN(
		[]string{"red", "green", "blue"},
		[]float64{1, 1, 1},
		2,
		vrand.WithWeightedRandSource(mathrand.New(mathrand.NewSource(2))),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(withReplacement)
	fmt.Println(len(unique))
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

	strict, err := vrand.BytesWithOptions(
		4,
		vrand.WithRandomReader(strings.NewReader("x")),
		vrand.WithStrictCryptoRandom(),
	)
	fmt.Println(len(strict), err != nil)
}
```
