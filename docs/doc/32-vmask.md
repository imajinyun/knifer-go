# vmask Quickstart

`vmask` provides masking helpers for common sensitive data, including names, ID numbers, phones, addresses, email, passwords, bank cards, IPs, license plates, and custom ranges.

## Golden path APIs

The first-choice API set for this facade is kept in sync with `ai-context.json` and the generated tools catalog.

- `BankCard`
- `Address`
- `CarLicense`
- `ChineseName`
- `Clear`

## Which helper should I use?

| Goal | Start with | Notes |
| --- | --- | --- |
| Mask with a known data type | `ChineseName`, `MobilePhone`, `Email`, `Password`, `BankCard`, `IPv4`, `IPv6` | Prefer explicit helpers when the data category is known at compile time. |
| Dispatch by runtime type | `Masked` | Use when masking type comes from configuration or metadata. |
| Return nil for cleared values | `MaskedPtr` with `ClearToNullType`, or `ClearToNil` | Useful for pointer-oriented DTOs that distinguish empty from absent. |
| Clear to empty string | `Clear` or `Masked(..., ClearToEmptyType)` | Use when downstream systems require a string value. |
| Keep only the first character | `FirstMask` | Works on rune-based strings. |
| Mask a custom range | `Hide` | Uses rune indexes in `[start, end)` and replaces with `*`. |
| Mask IDs with visible edges | `IDCardNum`, `Passport`, `CreditCode` | Choose visible prefix/suffix based on product and compliance requirements. |

## Masking safety checklist

- Mask as close to the display/logging boundary as possible; do not replace encrypted or hashed storage with masked plaintext.
- Treat masking as presentation redaction, not anonymization. Masked values may still be linkable or reversible with context.
- Prefer explicit helpers over runtime dispatch when the data category is known, so code review can verify the policy.
- Use `Hide` carefully with rune indexes, especially for mixed-width Unicode strings.
- Avoid logging original values before masking, including error messages and structured fields.
- Document how many leading/trailing characters remain visible for identifiers, cards, and addresses.

## Mask by specific data type

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vmask"
)

func main() {
	fmt.Println(vmask.ChineseName("\u5f20\u4e09\u4e30"))
	fmt.Println(vmask.MobilePhone("13800138000"))
	fmt.Println(vmask.Email("alice@example.com"))
	fmt.Println(vmask.Password("secret"))
}
```

## Dispatch with built-in types

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vmask"
)

func main() {
	fmt.Println(vmask.Masked("13800138000", vmask.MobilePhoneType))
	fmt.Println(vmask.Masked("alice@example.com", vmask.EmailType))
	fmt.Println(vmask.Masked("192.168.1.10", vmask.IPv4Type))

	cleared := vmask.MaskedPtr("secret", vmask.ClearToNullType)
	fmt.Println(cleared == nil)
}
```

## Mask IDs, bank cards, and addresses

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vmask"
)

func main() {
	fmt.Println(vmask.IDCardNum("110101199001011234", 6, 4))
	fmt.Println(vmask.BankCard("6222020202020202020"))
	fmt.Println(vmask.Address("\u5317\u4eac\u5e02\u671d\u9633\u533a\u793a\u4f8b\u8def 100 \u53f7", 6))
	fmt.Println(vmask.CreditCode("91310000123456789X"))
}
```

## Customize hidden ranges and empty values

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vmask"
)

func main() {
	fmt.Println(vmask.Hide("abcdef", 1, 4))
	fmt.Println(vmask.FirstMask("sensitive-info"))
	fmt.Println(vmask.Clear())
	fmt.Println(vmask.ClearToNil() == nil)
}
```

## When not to use vmask

- Use encryption, hashing, tokenization, or access-control systems when sensitive data must be protected at rest or across services.
- Use irreversible anonymization or aggregation when privacy requirements prohibit re-identification.
- Avoid generic runtime masking when a field has a regulated format with a specific redaction policy.
- Do not use masked values as stable security identifiers unless collision and linkability risks are acceptable.

## Related packages

- Use `vstr` when masking depends on trimming, substring handling, or other text normalization.
- Use `vident` when identity documents should be validated or parsed before redaction.
- Use `vlog` when redacted values are emitted through application diagnostics.

## Benchmarks and trade-offs

- Explicit helpers are simple and cheap string transformations suitable for logs and UI rendering.
- Runtime `Masked` dispatch is flexible but hides policy selection behind a `Type` value, so it is less transparent in reviews.
- Rune-based masking handles Unicode better than byte offsets, with small extra iteration cost.
- Keeping visible prefixes/suffixes improves usability but increases re-identification risk.
- `MaskedPtr` and `ClearToNil` make absent-value semantics explicit, but callers must handle nil pointers.

## FAQ

### Is masking the same as anonymization?

No. Masking hides part of a value for display. It may still reveal enough information to identify a person or account when combined with other data.

### Should I store masked values?

Usually no. Store protected canonical data using encryption, hashing, or tokenization as appropriate, then mask only for presentation or logs.

### Are `Hide` indexes byte-based?

No. `Hide` uses rune indexes, so it is safer for Unicode text than byte slicing.

### When should I use `MaskedPtr`?

Use it when a downstream DTO needs `nil` to represent cleared data, especially with `ClearToNullType`.
