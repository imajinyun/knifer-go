# vpass Quickstart

`vpass` provides password strength analysis, scoring, strength levels, and shortcut checks for strong or weak passwords.

## Analyze password strength

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vpass"
)

func main() {
	analysis := vpass.Analyze("G0-Knifer#Pass2026")
	fmt.Println(analysis.Score)
	fmt.Println(analysis.Strength == vpass.StrengthVeryStrong)
}
```

## Get scores and strength levels

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vpass"
)

func main() {
	fmt.Println(vpass.Score("password"))
	fmt.Println(vpass.StrengthOf("password") == vpass.StrengthVeryWeak)
	fmt.Println(vpass.StrengthUnknown.String())
}
```

## Quickly detect strong and weak passwords

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vpass"
)

func main() {
	fmt.Println(vpass.IsStrong("G0-Knifer#Pass2026"))
	fmt.Println(vpass.IsWeak("12345"))
}
```
