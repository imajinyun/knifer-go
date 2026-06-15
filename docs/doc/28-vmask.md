# vmask Quickstart

`vmask` provides masking helpers for common sensitive data, including names, ID numbers, phones, addresses, email, passwords, bank cards, IPs, license plates, and custom ranges.

## Mask by specific data type

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vmask"
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

	"github.com/imajinyun/go-knifer/vmask"
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

	"github.com/imajinyun/go-knifer/vmask"
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

	"github.com/imajinyun/go-knifer/vmask"
)

func main() {
	fmt.Println(vmask.Hide("abcdef", 1, 4))
	fmt.Println(vmask.FirstMask("sensitive-info"))
	fmt.Println(vmask.Clear())
	fmt.Println(vmask.ClearToNil() == nil)
}
```
