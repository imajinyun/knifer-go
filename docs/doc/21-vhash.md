# vhash Quickstart

`vhash` provides multiple non-cryptographic string hash algorithms, including FNV, BKDR, DJB, SDBM, and Java String hashCode, for bucketing, legacy compatibility, or hash behavior tests.

## Which helper should I use?

Choose a hash by compatibility requirements first, then by output width and collision tolerance.

| Need | Use | Notes |
| --- | --- | --- |
| General non-cryptographic string hash | `FnvHash`, `Hash32` | FNV is a familiar default for bucketing and tests; `Hash32` accepts a standard-library constructor. |
| Compatibility with Java hashCode | `JavaDefaultHash` | Use when matching existing Java systems, persisted buckets, or test vectors. |
| Classic 32-bit algorithms | `BkdrHash`, `DjbHash`, `SdbmHash`, `RsHash`, `JsHash`, `PjwHash`, `ElfHash`, `ApHash` | Pick only when an existing system expects that exact algorithm. |
| 64-bit-style legacy hashes | `HfHash`, `HfIpHash`, `TianlHash` | Use for compatibility with code that already stores or routes by these outputs. |
| Custom `hash.Hash32` provider | `Hash32` | Good for testing or wiring a standard `hash` implementation without duplicating write/sum boilerplate. |

## Hash correctness checklist

- Do not use these helpers for passwords, tokens, signatures, MACs, or any cryptographic purpose.
- Choose the algorithm based on compatibility and document why that exact hash is required.
- Keep input normalization stable: case folding, trimming, encoding, and separators must match every producer and consumer.
- Treat collisions as expected. Bucketing code must tolerate two different inputs producing the same value.
- Be explicit about signed versus unsigned return types when persisting or exchanging hash values.
- Add test vectors before changing a hash algorithm used for persisted keys, sharding, or rollout decisions.

## When not to use vhash

- Use `crypto/sha256`, `crypto/hmac`, `bcrypt`, or another vetted primitive for security-sensitive hashing.
- Use a consistent hashing library when you need ring movement guarantees for distributed caches or shards.
- Use `hash/maphash` for process-local randomized hashing where persistence and cross-process stability are not required.
- Avoid changing algorithms for data already partitioned, stored, or routed by old hash values unless you have a migration plan.

## Related packages

- Use `vcrypto` when the hash must be cryptographic or used for integrity, signatures, or authentication.
- Use `vblf` when hash functions are part of probabilistic set membership checks.
- Use `vcodec` when hash bytes need hex, Base64, or URL-safe representation.

## Benchmarks and trade-offs

Benchmark with representative key lengths and distributions before choosing a hash for hot bucketing paths:

```bash
go test -bench=. -benchmem -run=^$ ./internal/hash ./vhash
```

Non-cryptographic hashes are fast and deterministic, but they do not protect integrity or confidentiality. Wider outputs can reduce accidental collisions, but compatibility with existing stored values often matters more than raw speed.

## FAQ

### Are these hashes secure?

No. They are non-cryptographic helpers for bucketing, compatibility, and tests. Use Go's `crypto/*` packages for security-sensitive work.

### Which hash should I choose for new code?

Use `FnvHash` or `Hash32` for simple non-security bucketing unless you must match a legacy algorithm. Add collision handling regardless of the choice.

### Can I persist hash outputs?

Yes, but treat the algorithm and input normalization as part of the data contract. Persisted outputs make future algorithm changes a migration problem.

## Use FNV and generic Hash32

```go
package main

import (
	"fmt"
	"hash/fnv"

	"github.com/imajinyun/knifer-go/vhash"
)

func main() {
	fmt.Println(vhash.FnvHash("knifer-go"))
	fmt.Println(vhash.Hash32("knifer-go", fnv.New32))
	fmt.Println(vhash.Hash32("knifer-go", nil))
}
```

## Use classic 32-bit string hashes

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vhash"
)

func main() {
	s := "knifer-go"
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

	"github.com/imajinyun/knifer-go/vhash"
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

	"github.com/imajinyun/knifer-go/vhash"
)

func main() {
	s := "bucket-key"
	fmt.Println(vhash.HfHash(s))
	fmt.Println(vhash.HfIpHash("10.0.0.1"))
	fmt.Println(vhash.TianlHash(s))
}
```
