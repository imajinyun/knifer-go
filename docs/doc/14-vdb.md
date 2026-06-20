# vdb Quickstart

`vdb` provides lightweight SQL builders, condition composition, pagination, entity mapping, and `database/sql` session wrappers, with emphasis on parameterized SQL and dialect placeholders.

## Build SELECT SQL

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vdb"
)

func main() {
	sqlText, args, err := vdb.NewBuilder(
		vdb.WithDialect(vdb.DialectPostgres),
		vdb.WithWrapper(vdb.WrapperForDialect(vdb.DialectPostgres)),
	).
		Select("id", "name").
		From("users").
		Where(vdb.Eq("name", "alice"), vdb.Gte("age", 18)).
		OrderBy(vdb.Desc("id")).
		SQL()
	if err != nil {
		panic(err)
	}
	fmt.Println(sqlText)
	fmt.Println(args)
}
```

## Compose condition fragments

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vdb"
)

func main() {
	where, args, err := vdb.BuildConditions(
		vdb.AndGroup(
			vdb.Eq("status", "active"),
			vdb.OrWith(vdb.Like("name", vdb.BuildLikeValue("bob", "contains"))),
		),
		vdb.In("role", "admin", "owner"),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(where)
	fmt.Println(args)
}
```

## Generate write SQL with Entity

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vdb"
)

func main() {
	user := vdb.NewEntity("users").Set("name", "alice").Set("age", 30)

	insertSQL, insertArgs, err := vdb.Insert(user).SQL()
	if err != nil {
		panic(err)
	}
	fmt.Println(insertSQL, insertArgs)

	updateSQL, updateArgs, err := vdb.Update(user).Where(vdb.Eq("id", 1)).SQL()
	if err != nil {
		panic(err)
	}
	fmt.Println(updateSQL, updateArgs)
}
```

## Pagination and dialect wrapping

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vdb"
)

func main() {
	page := vdb.NewPage(2, 10, vdb.Desc("created_at"))
	sqlText, args, err := vdb.NewBuilder(vdb.WithDialect(vdb.DialectMySQL)).
		Select("id", "created_at").
		From("orders").
		Page(page).
		SQL()
	if err != nil {
		panic(err)
	}
	fmt.Println(sqlText)
	fmt.Println(args)
}
```

`Page` carries optional `ORDER BY` fields so pagination can stay stable without a separate `OrderBy(...)` call when the page contract owns the sort.

## Validate identifiers before using raw SQL escape hatches

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vdb"
)

func main() {
	fmt.Println(vdb.IsSafeIdentifier("orders.created_at"))
	fmt.Println(vdb.IsSafeIdentifier("orders; drop table orders"))
}
```

Use `IsSafeIdentifier` only for guardrails around trusted identifier inputs. Values should still go through placeholders instead of string concatenation.
