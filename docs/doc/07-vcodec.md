# vcodec Quickstart

`vcodec` provides common encoding and decoding helpers for Base64, URL-safe Base64, raw URL Base64, Hex, Base32, Base58, Base62, Morse code, and ROT transforms.

## Which helper should I use?

Choose the codec based on the transport or representation requirement, not just convenience.

| Need | Use | Notes |
| --- | --- | --- |
| Standard Base64 text | `Base64Encode`, `Base64EncodeStr`, `Base64Decode`, `Base64DecodeStr` | Good for generic text/binary transport where `+`, `/`, and padding are acceptable. |
| URL-safe Base64 with padding | `Base64URLEncode`, `Base64URLEncodeStr`, `Base64URLDecode`, `Base64URLDecodeStr` | Prefer when encoded text appears in URLs, cookies, or filenames and downstream expects padded URL-safe Base64. |
| URL-safe Base64 without padding | `Base64RawURLEncode`, `Base64RawURLEncodeStr`, `Base64RawURLDecode`, `Base64RawURLDecodeStr` | Common for tokens and compact wire formats where trailing `=` should be omitted. |
| Hex text for debugging or fixed textual representation | `HexEncode`, `HexEncodeStr`, `HexDecode`, `HexDecodeStr` | Easier for humans to inspect, but larger than Base64. |
| RFC 4648 Base32 text | `Base32Encode`, `Base32Decode`, `Base32EncodeWithEncoding`, `Base32DecodeWithEncoding` | Use when downstream systems require Base32 or Base32Hex alphabets. |
| Bitcoin/Flickr Base58 text | `Base58Encode`, `Base58Decode`, `Base58EncodeWithAlphabet`, `Base58DecodeWithAlphabet` | Useful for compact human-facing identifiers that avoid visually ambiguous characters. |
| URL-friendly Base62 text | `Base62Encode`, `Base62Decode` | Useful for short IDs where digits and ASCII letters are accepted. |
| Morse code | `MorseEncode`, `MorseDecode` | Encodes supported ASCII letters, digits, and punctuation; unsupported characters return an input error. |
| Simple ROT transforms | `ROT13`, `ROT47`, `ROTN` | Obfuscation and compatibility helpers only; these are not security features. |

## Codec selection checklist

- Match the downstream protocol first: some systems require padded Base64, others require raw URL-safe Base64 or hex.
- Treat decode failures as input-validation failures and handle the returned error explicitly.
- Do not use encoding as a security boundary. Base64 and hex are reversible representations, not encryption.
- Prefer URL-safe variants when encoded data will be embedded in URLs, query strings, cookies, or filesystem-safe tokens.
- Prefer hex when readability and deterministic byte-to-text mapping matter more than compactness.
- Prefer Base58 when IDs are typed or read by humans and ambiguous glyphs should be avoided.
- Use `Base32HexEncoding` only when the downstream protocol explicitly requires the extended hex alphabet.
- Treat Morse and ROT helpers as reversible text transforms, not validation, escaping, or cryptography.

## When not to use vcodec

- Use standard library encoders directly when a hot path needs streaming encoders, preallocated destination buffers, or lower-level control.
- Use cryptographic helpers when confidentiality, integrity, signatures, or password hashing are required; encoding is reversible.
- Use URL escaping, JSON encoding, or form encoding when the problem is syntax escaping rather than byte-to-text representation.
- Avoid repeatedly encoding large payloads into strings when byte slices or streaming APIs would avoid extra allocations.

## Related packages

- Use `vcrypto` when encoded data also needs hashing, encryption, signing, or verification.
- Use `vhash` when the goal is a deterministic non-cryptographic hash rather than an encoding.
- Use `vurl` when URL escaping is part of URL parsing, query construction, or normalization.

## Benchmarks and trade-offs

Benchmark with representative payload sizes and output forms before choosing helpers in hot paths:

```bash
go test -bench=. -benchmem -run=^$ ./internal/codec ./vcodec
```

String helpers are convenient for examples and text protocols, but they allocate strings in addition to encoded bytes. Byte helpers are better when the next step already works with `[]byte`.

Base64 is more compact than hex, while hex is easier to inspect and compare manually. URL-safe variants avoid `+` and `/`, and raw URL variants omit padding for compact token formats. Base58 and Base62 use integer conversion internally, so benchmark them with representative payload sizes before putting them on hot paths.

## FAQ

### Is Base64 a security feature?

No. Base64 only changes representation. Anyone who can read the encoded value can decode it. Use real cryptography when confidentiality or integrity matters.

### How do I choose between Base64URL and Base64RawURL?

Use `Base64URL*` when the protocol expects URL-safe Base64 with padding. Use `Base64RawURL*` when the protocol expects the same alphabet without trailing `=` padding, which is common in tokens and compact identifiers.

### When should I use hex instead of Base64?

Use hex when operators or logs need byte-for-byte readability, or when an external format explicitly requires hexadecimal text. Use Base64 when compactness matters more.

### Is Base58 or Base62 encryption?

No. Base58 and Base62 are reversible encodings. Use them for transport and human-facing identifiers, not confidentiality or integrity.

## Encode and decode Base64 strings

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vcodec"
)

func main() {
	encoded := vcodec.Base64EncodeStr("hello")
	decoded, err := vcodec.Base64DecodeStr(encoded)
	if err != nil {
		panic(err)
	}

	fmt.Println(encoded, decoded)
}
```

## URL-safe Base64

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vcodec"
)

func main() {
	encoded := vcodec.Base64URLEncode([]byte("a/b?c=d"))
	decoded, err := vcodec.Base64URLDecode(encoded)
	if err != nil {
		panic(err)
	}

	fmt.Println(encoded, string(decoded))
}
```

## Raw URL Base64

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vcodec"
)

func main() {
	encoded := vcodec.Base64RawURLEncode([]byte("token"))
	decoded, err := vcodec.Base64RawURLDecode(encoded)
	if err != nil {
		panic(err)
	}

	fmt.Println(encoded, string(decoded))
}
```

## Encode and decode Hex

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vcodec"
)

func main() {
	hexText := vcodec.HexEncodeStr("go")
	plain, err := vcodec.HexDecodeStr(hexText)
	if err != nil {
		panic(err)
	}

	fmt.Println(hexText, plain)
}
```

## Encode and decode Base32

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vcodec"
)

func main() {
	encoded := vcodec.Base32Encode([]byte("go"))
	decoded, err := vcodec.Base32Decode(encoded)
	if err != nil {
		panic(err)
	}

	fmt.Println(encoded, string(decoded))
}
```

## Encode human-facing IDs with Base58

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vcodec"
)

func main() {
	encoded := vcodec.Base58Encode([]byte("hello world"))
	decoded, err := vcodec.Base58Decode(encoded)
	if err != nil {
		panic(err)
	}

	fmt.Println(encoded, string(decoded))
}
```

## Encode compact IDs with Base62

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vcodec"
)

func main() {
	encoded := vcodec.Base62Encode([]byte("hello"))
	decoded, err := vcodec.Base62Decode(encoded)
	if err != nil {
		panic(err)
	}

	fmt.Println(encoded, string(decoded))
}
```

## Encode and decode Morse code

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vcodec"
)

func main() {
	encoded, err := vcodec.MorseEncode("SOS 1")
	if err != nil {
		panic(err)
	}
	decoded, err := vcodec.MorseDecode(encoded)
	if err != nil {
		panic(err)
	}

	fmt.Println(encoded, decoded)
}
```

## Apply ROT transforms

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vcodec"
)

func main() {
	encoded := vcodec.ROT13("hello")

	fmt.Println(encoded)
	fmt.Println(vcodec.ROT13(encoded))
}
```
