# vhash Quickstart

`vhash` provides multiple non-cryptographic string hash algorithms, including FNV, BKDR, DJB, SDBM, and Java String hashCode, for bucketing, legacy compatibility, or hash behavior tests.

## Use FNV and generic Hash32

```go
package main

import (
	"fmt"
	"hash/fnv"

	"github.com/imajinyun/go-knifer/vhash"
)

func main() {
	fmt.Println(vhash.FnvHash("go-knifer"))
	fmt.Println(vhash.Hash32("go-knifer", fnv.New32))
	fmt.Println(vhash.Hash32("go-knifer", nil))
}
```

## Use classic 32-bit string hashes

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vhash"
)

func main() {
	s := "go-knifer"
	fmt.Println(vhash.BkdrHash(s))
	fmt.Println(vhash.DjbHash(s))
	fmt.Println(vhash.SdbmHash(s))
	fmt.Println(vhash.JavaDefaultHash(s))
}
```

## Choose other algorithms for compatibility

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vhash"
)

func main() {
	s := "192.168.0.1"
	fmt.Println(vhash.RsHash(s))
	fmt.Println(vhash.JsHash(s))
	fmt.Println(vhash.PjwHash(s))
	fmt.Println(vhash.ElfHash(s))
}
```

## Use 64-bit algorithms

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vhash"
)

func main() {
	s := "bucket-key"
	fmt.Println(vhash.HfHash(s))
	fmt.Println(vhash.HfIpHash("10.0.0.1"))
	fmt.Println(vhash.TianlHash(s))
}
```
