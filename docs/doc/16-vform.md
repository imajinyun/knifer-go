# vform Quickstart

`vform` validates common form fields such as email, mobile numbers, URLs, IPs, ID cards, Chinese text, and numeric strings, and also supports injecting matchers for selected rules.

## Scope and struct validation direction

`vform` is intentionally a string-level validation facade. It validates individual values that commonly arrive from forms, query parameters, CSV files, and decoded maps. It does not implement struct-tag validation.

For struct validation, prefer the ecosystem standard [`github.com/go-playground/validator/v10`](https://github.com/go-playground/validator). That package already covers the broad surface users expect from tag-based validation: nested structs, slices and maps, cross-field rules, translations, custom validators, and mature production behavior. Duplicating that surface in a lightweight `vvalidate` package would add a long-term maintenance burden and risk subtle behavioral gaps.

Recommended split:

- Use `vform` when code needs a small predicate for one value, or when a `vbean.Decode`/`vbean.DecodeResult` flow wants to validate selected fields explicitly after conversion.
- Use `go-playground/validator/v10` when the validation contract belongs on struct tags and must traverse nested request/config DTOs.
- Do not add `go-playground/validator` to `go-knifer` itself for this direction decision; keeping it as an application dependency avoids increasing the base module dependency graph for users who only need string predicates.

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

	"github.com/imajinyun/go-knifer/vform"
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

	"github.com/imajinyun/go-knifer/vform"
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

	"github.com/imajinyun/go-knifer/vform"
)

func main() {
	fmt.Println(vform.IsIDCard("11010519491231002X"))
	fmt.Println(vform.IsIDCardWithOptions(
		"custom-id",
		vform.WithIDCardMatcher(func(s string) bool { return s == "custom-id" }),
	))
}
```

## Inject custom matchers

```go
package main

import (
	"fmt"
	"strings"

	"github.com/imajinyun/go-knifer/vform"
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
