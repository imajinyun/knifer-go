# vblf Quickstart

`vblf` provides Bloom filter facade APIs, including function-hash filters, bitmap/bitset filters, hash functions, and initialization from files or readers.

## Which helper should I use?

Choose the filter by how you size the set, how many hash functions you need, and whether false positives are acceptable for the caller.

| Need | Use | Notes |
| --- | --- | --- |
| Simple string membership with default hashes | `NewDefaultBloomFilter`, `NewDefaultFilter` | Good for small allow/deny prechecks where occasional false positives are acceptable. |
| Explicit hash and machine sizing | `NewFuncFilterWithOptions`, `WithMaxValue`, `WithMachineNum`, `WithHashFunc` | Keeps hashing policy visible and testable. |
| Compose several function filters | `NewBitMapBloomFilter`, `NewBitMapBloomFilterWithFilters` | Use when multiple hash functions should update one bitmap. |
| Size by capacity, expected elements, and hash count | `NewBitSetBloomFilter`, `NewBitSetBloomFilterWithOptions` | Prefer when you want false-positive behavior tied to sizing inputs. |
| Load membership data from text | `InitFromReader`, `InitFromFileWithOptions` | Inject `WithOpenFile` in tests and for nonstandard file sources. |
| Reuse hash helpers | `BloomFNVHash`, `BloomBKDRHash`, `JavaDefaultHash`, filter constructors | Keep hash selection consistent across filters and tests. |

## Bloom filter correctness checklist

- Confirm that false positives are acceptable. Bloom filters can say “maybe present”; they cannot prove absence errors once sized incorrectly.
- Size the filter from expected cardinality and desired false-positive rate instead of copying arbitrary capacity values.
- Keep the same hash functions and normalization rules for `Add` and `Contains`; mismatches silently corrupt results.
- Validate constructor inputs with the `E` variants when parameters come from configuration or users.
- Do not use a Bloom filter as the only authorization, uniqueness, or payment correctness check; back it with an exact store when correctness matters.
- Initialize from readers/files before serving traffic, or clearly handle partial initialization errors.

## When not to use vblf

- Use a `map[string]struct{}` or database index when exact membership is required and memory/storage cost is acceptable.
- Use a counting Bloom filter or another data structure when deletes must be supported safely.
- Use a probabilistic sketch library when you need cardinality, frequency, or heavy-hitter estimates rather than membership checks.
- Avoid package-level defaults for tenant-specific or request-specific filters; construct independent filters with explicit sizing.
- Avoid custom hash functions unless you can test distribution with representative input data.

## Related packages

- Use `vhash` when you need standalone non-cryptographic hash helpers outside a Bloom filter.
- Use `vcache` when membership checks should be paired with bounded in-memory value storage.
- Use `vset` when exact membership, deletion, and iteration are required instead of probabilistic checks.

## Benchmarks and trade-offs

Benchmark candidate filters with realistic key counts, key shapes, and false-positive expectations:

```bash
go test -bench=. -benchmem -run=^$ ./internal/bloomfilter ./vblf
```

Larger bitsets reduce false positives but consume more memory. More hash functions can improve accuracy up to the sizing optimum, then add CPU without meaningful benefit. Function filters are convenient for common strings; bitset filters make sizing choices more explicit.

File and reader initialization costs are usually startup costs. Keep them out of request paths unless the input is small and bounded.

## FAQ

### Can a Bloom filter return false negatives?

Not if it is initialized and queried with the same hashing and normalization rules. False negatives usually indicate a bug: different hashes, partial initialization, data corruption, or querying before all keys were added.

### Can I remove an item from a vblf filter?

No. The exposed filters are insert-and-query structures. Use an exact set, rebuild the filter, or choose a counting Bloom filter implementation when deletes are required.

### Which constructor should I start with?

Start with `NewBitSetBloomFilterWithOptions` when you know expected element counts, or `NewDefaultBloomFilter` for simple prechecks. Switch to explicit hash/filter composition when tests show the default false-positive behavior is not acceptable.

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
