# vid Quickstart

`vid` provides UUID, ObjectId, NanoId, and Snowflake ID generators, with support for injecting random sources, clocks, and Snowflake worker/datacenter configuration.

## Generate UUIDs

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vid"
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

	"github.com/imajinyun/go-knifer/vid"
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

	"github.com/imajinyun/go-knifer/vid"
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

	"github.com/imajinyun/go-knifer/vid"
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
