# vid Quickstart

`vid` provides UUID, ObjectId, NanoId, and Snowflake ID generators, with support for injecting random sources, clocks, and Snowflake worker/datacenter configuration.

## Golden path APIs

The first-choice API set for this facade is kept in sync with `ai-context.json` and the generated tools catalog.

- `ConfigureDefaultSnowflake`
- `GetSnowflakeNextID`
- `FastSimpleUUID`
- `ConfigureDefaultFallbackRandomSourceProvider`
- `CreateSnowflake`

## Which helper should I use?

Choose ID helpers by the property you need: randomness, compactness, sortability, or distributed sequence generation.

| Need | Use | Notes |
| --- | --- | --- |
| General random identifiers | `RandomUUID`, `SimpleUUID` | Good default for opaque IDs that do not need ordering. |
| Fast non-cryptographic UUID-style IDs | `FastUUID`, `FastSimpleUUID` | Use only when unpredictability is not a security requirement. |
| MongoDB-style time-bearing IDs | `ObjectId` | Useful when interoperability or approximate creation time is desired. |
| Compact random textual IDs | `NanoId`, `NanoIdN`, `NanoIdWithOptions` | Tune length and alphabet for the collision budget and downstream format. |
| Distributed sortable numeric IDs | `CreateSnowflake`, `GetSnowflakeNextID`, `GetSnowflakeNextIDStr` | Configure worker/datacenter ids so generators do not collide. |
| Deterministic tests | random source, clock, and Snowflake options | Inject deterministic dependencies only in tests and examples. |

## ID safety checklist

- Do not treat ordinary IDs as authentication secrets. Use `vrand.SecureBytes` for bearer tokens, reset tokens, API keys, and session secrets.
- Choose ID length and alphabet based on collision risk, storage limits, and whether users will type or copy the value.
- Avoid fast pseudo-random ID helpers when IDs must be unpredictable to attackers.
- Configure Snowflake worker and datacenter IDs uniquely per generator instance to avoid collisions.
- Remember that time-bearing IDs can reveal creation time or ordering; avoid them where metadata disclosure matters.
- Keep deterministic random sources and clocks limited to tests so production IDs remain unique and appropriately unpredictable.

## When not to use vid

- Use `vrand.SecureBytes`, a token package, or an authentication service for bearer secrets, reset links, API keys, and session tokens.
- Use database sequences or transactionally allocated IDs when the database must be the source of ordering and uniqueness.
- Use application-specific key formats when IDs must embed tenant, type, checksum, or migration metadata.
- Avoid time-bearing IDs when creation-time or ordering metadata would disclose sensitive business information.
- Avoid package-level Snowflake defaults in multi-tenant libraries; construct and pass explicit generators instead.

## Related packages

- Use `vrand` when identifiers must be backed by secure random bytes or deterministic pseudo-random test sources.
- Use `vcodec` when IDs need Base64, hex, or URL-safe encoding for transport.
- Use `vcrypto` when identifiers are tied to signing, MACs, or cryptographic verification.

## Benchmarks and trade-offs

Benchmark ID generation under representative concurrency and collision requirements:

```bash
go test -bench=. -benchmem -run=^$ ./internal/id ./vid
```

Random UUID and NanoID helpers are easy to distribute, but collision probability depends on entropy, alphabet, and length. Snowflake IDs are compact and sortable, but require unique worker/datacenter configuration and careful clock behavior.

Fast pseudo-random helpers can be useful for non-security identifiers in hot paths, but unpredictability is not their contract. Prefer secure randomness for any ID that grants access or hides resources.

## FAQ

### Can I use vid IDs as login or reset tokens?

No. IDs identify records; tokens grant access. Use `vrand.SecureBytes` or a dedicated token workflow for credentials and bearer secrets.

### When should I use Snowflake IDs?

Use Snowflake IDs when you need compact sortable numeric IDs generated across multiple workers. Ensure worker/datacenter configuration is unique across the deployment.

### Are ObjectIds or Snowflake IDs private?

No. They can reveal ordering or time-derived metadata. Use random opaque IDs when creation-time disclosure is undesirable.

### Which vid helpers return errors?

Most convenience ID helpers return strings and use compatibility fallback behavior. Use option-based constructors and validate configuration errors at setup boundaries when worker IDs, alphabets, clocks, or random readers come from configuration.

## Generate UUIDs

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vid"
)

func main() {
	fmt.Println(vid.RandomUUID())
	fmt.Println(vid.SimpleUUID())
	fmt.Println(vid.FastUUID())
	fmt.Println(vid.FastSimpleUUID())
}
```

## Generate ObjectIds and NanoIds

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vid"
)

func main() {
	fmt.Println(vid.ObjectId())
	fmt.Println(vid.NanoId())
	fmt.Println(vid.NanoIdN(12))
	fmt.Println(vid.NanoIdWithOptions(
		vid.WithNanoIDAlphabet("abcdef012345"),
		vid.WithNanoIDLength(16),
	))
}
```

## Use the Snowflake generator

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vid"
)

func main() {
	sf := vid.CreateSnowflake(1, 1)
	fmt.Println(sf.WorkerID(), sf.DatacenterID())
	fmt.Println(sf.NextID())
	fmt.Println(sf.NextIDStr())
}
```

## Configure the default Snowflake generator

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vid"
)

func main() {
	vid.ConfigureDefaultSnowflake(
		vid.WithSnowflakeWorkerID(2),
		vid.WithSnowflakeDatacenterID(3),
		vid.WithSnowflakeCache(false),
	)

	fmt.Println(vid.GetSnowflakeNextID())
	fmt.Println(vid.GetSnowflakeNextIDStr())
}
```
