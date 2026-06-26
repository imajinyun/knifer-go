# vform Quickstart

`vform` validates common form fields such as email, mobile numbers, URLs, IPs, ID cards, Chinese text, and numeric strings, and also supports injecting matchers for selected rules.

## Which helper should I use?

| Goal | Start with | Notes |
| --- | --- | --- |
| Validate one email string | `IsEmail` | Lightweight predicate for form/query/CSV values. |
| Validate one mainland China mobile number | `IsMobile` | Use a custom matcher when product rules differ from the built-in pattern. |
| Validate a URL-shaped string | `IsURL` | Checks shape, not network reachability or SSRF safety. |
| Validate IP strings | `IsIPv4`, `IsIPv6` | Use when the value must be specifically v4 or v6. |
| Validate Chinese text | `IsChinese` | Predicate-style check for string content. |
| Validate numeric strings | `IsNumberStr` | Use for string-level checks before parsing or conversion. |
| Validate ID cards | `IsIDCard`, `IsIDCardWithOptions` | Built-in rule is for common ID-card format checks; inject custom matcher for product-specific identifiers. |
| Override selected rules | `WithEmailMatcher`, `WithMobileMatcher`, `WithIDCardMatcher`, `WithChineseMatcher`, `WithNumberMatcher` | Options are per call and keep tests deterministic. |

## Form validation checklist

- Treat `vform` predicates as input-shape checks, not full business authorization or identity verification.
- Use struct validation libraries for DTO-wide requirements, cross-field checks, localization, and nested data.
- Use custom matchers for product-specific mobile, ID, or internal identifier rules instead of widening global assumptions.
- Validate URLs for scheme, host allowlist, DNS/IP policy, and reachability separately when they will be fetched by the server.
- Parse numeric strings after `IsNumberStr` when range, precision, or overflow matters.
- Keep matcher functions pure and fast; predicates may be called in request validation loops.

## Scope and struct validation direction

`vform` is intentionally a string-level validation facade. It validates individual values that commonly arrive from forms, query parameters, CSV files, and decoded maps. It does not implement struct-tag validation.

For struct validation, prefer the ecosystem standard [`github.com/go-playground/validator/v10`](https://github.com/go-playground/validator). That package already covers the broad surface users expect from tag-based validation: nested structs, slices and maps, cross-field rules, translations, custom validators, and mature production behavior. Duplicating that surface in a lightweight `vvalidate` package would add a long-term maintenance burden and risk subtle behavioral gaps.

Recommended split:

- Use `vform` when code needs a small predicate for one value, or when a `vbean.Decode`/`vbean.DecodeResult` flow wants to validate selected fields explicitly after conversion.
- Use `go-playground/validator/v10` when the validation contract belongs on struct tags and must traverse nested request/config DTOs.
- Do not add `go-playground/validator` to `knifer-go` itself for this direction decision; keeping it as an application dependency avoids increasing the base module dependency graph for users who only need string predicates.

```go
package main

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type Signup struct {
	Email string `validate:"required,email"`
	Age   int    `validate:"gte=18"`
}

func main() {
	validate := validator.New()
	err := validate.Struct(Signup{Email: "alice@example.com", Age: 20})
	fmt.Println(err == nil)
}
```

## Validate email, mobile, and URL values

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vform"
)

func main() {
	fmt.Println(vform.IsEmail("alice@example.com"))
	fmt.Println(vform.IsMobile("13800138000"))
	fmt.Println(vform.IsURL("https://example.com/path"))
}
```

## Validate IP, Chinese text, and numeric strings

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vform"
)

func main() {
	fmt.Println(vform.IsIPv4("127.0.0.1"))
	fmt.Println(vform.IsIPv6("::1"))
	fmt.Println(vform.IsChinese("\u4e2d\u6587"))
	fmt.Println(vform.IsNumberStr("-12.34"))
}
```

## Validate ID card numbers

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vform"
)

func main() {
	fmt.Println(vform.IsIDCard("11010519491231002X"))
	fmt.Println(vform.IsIDCardWithOptions(
		"custom-id",
		vform.WithIDCardMatcher(func(s string) bool { return s == "custom-id" }),
	))
}
```

## When not to use vform

- Use `go-playground/validator/v10` or similar libraries for struct tags, translations, nested structs, and cross-field validation.
- Use `net/url`, URL policies, DNS/IP checks, and SSRF guards when a URL will be requested by backend code.
- Use authoritative external services when email, phone, or identity ownership must be verified.
- Use numeric parsers such as `vnum` when the task is converting values and reporting parse errors rather than checking string shape.

## Related packages

- Use `vbean` when validated inputs need to be bound into typed structs or copied between DTOs.
- Use `vnum` when numeric fields need parsing, decimal precision, rounding, or range calculations.
- Use `vregex` when custom validation predicates rely on reviewed pattern matching.

## Benchmarks and trade-offs

- Predicate helpers are cheap and dependency-light, which fits per-field validation in handlers and import pipelines.
- Regex-style shape checks can reject obvious bad input quickly but cannot prove that an email, phone, URL, or ID belongs to a real user.
- Custom matcher injection improves product fit and testability, but matcher behavior is caller-owned and should be documented near the call site.
- Keeping struct validation out of `vform` avoids pulling a large dependency graph into users who only need simple predicates.

## FAQ

### Does `IsURL` make a URL safe to fetch?

No. It only validates URL shape. Server-side fetches still need allowed schemes, host allowlists, private-IP rejection, redirect policy, and timeouts.

### Why does `vform` not support struct tags?

Struct-tag validation is already well served by `go-playground/validator/v10`. `vform` intentionally stays focused on lightweight string predicates.

### When should I inject a matcher?

Inject a matcher when the built-in shape differs from product policy, such as internal email domains, nonstandard IDs, or test-only values.

### Is ID-card validation proof of identity?

No. It is a format/checksum-style predicate. Ownership and authenticity require separate authoritative verification.

## Inject custom matchers

```go
package main

import (
	"fmt"
	"strings"

	"github.com/imajinyun/knifer-go/vform"
)

func main() {
	fmt.Println(vform.IsEmailWithOptions(
		"user@internal",
		vform.WithEmailMatcher(func(s string) bool { return strings.HasSuffix(s, "@internal") }),
	))

	fmt.Println(vform.IsNumberStrWithOptions(
		"N/A",
		vform.WithNumberMatcher(func(s string) bool { return s == "N/A" }),
	))
}
```
