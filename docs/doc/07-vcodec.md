# vcodec Quickstart

`vcodec` provides common encoding and decoding helpers for Base64, URL-safe Base64, raw URL Base64, and Hex.

## Encode and decode Base64 strings

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vcodec"
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

	"github.com/imajinyun/go-knifer/vcodec"
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

	"github.com/imajinyun/go-knifer/vcodec"
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

	"github.com/imajinyun/go-knifer/vcodec"
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
