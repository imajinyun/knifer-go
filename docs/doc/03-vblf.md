# vblf Quickstart

`vblf` provides Bloom filter facade APIs, including function-hash filters, bitmap/bitset filters, hash functions, and initialization from files or readers.

## Default function Bloom filter

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vblf"
)

func main() {
	f := vblf.NewDefaultBloomFilter(1000)
	f.Add("user:1")

	fmt.Println(f.Contains("user:1"))
}
```

## Create a function filter with options

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vblf"
)

func main() {
	f := vblf.NewFuncFilterWithOptions(
		vblf.WithMaxValue(1000),
		vblf.WithMachineNum(vblf.BloomMachine64),
		vblf.WithHashFunc(func(s string) int64 {
			return int64(vblf.JavaDefaultHash(s))
		}),
	)

	f.Add("order:42")
	fmt.Println(f.Contains("order:42"))
}
```

## BitSet Bloom Filter

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vblf"
)

func main() {
	f := vblf.NewBitSetBloomFilter(1000, 5, 3)
	f.Add("hello")
	f.Add("world")

	fmt.Println(f.Contains("hello"), f.Contains("world"))
}
```

## Initialize from a reader

```go
package main

import (
	"fmt"
	"strings"

	"github.com/imajinyun/go-knifer/vblf"
)

func main() {
	f := vblf.NewBitSetBloomFilter(1000, 5, 3)
	if err := f.InitFromReader(strings.NewReader("alice\nbob\n")); err != nil {
		panic(err)
	}

	fmt.Println(f.Contains("alice"))
}
```
