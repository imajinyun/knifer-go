# vdb Quickstart

`vdb` provides lightweight SQL builders, condition composition, pagination, entity mapping, and `database/sql` session wrappers, with emphasis on parameterized SQL and dialect placeholders.

## Which helper should I use?

| Need | Use | Notes |
| --- | --- | --- |
| Build SELECT/INSERT/UPDATE/DELETE SQL | `Select`, `Insert`, `Update`, `Delete`, `NewBuilder` | Builders return SQL plus args so callers can execute parameterized statements. |
| Compose WHERE fragments | `Eq`, `Ne`, `Gt`, `Gte`, `Lt`, `Lte`, `Like`, `In`, `Between`, `AndGroup`, `OrGroup`, `BuildConditions` | Values become placeholders; identifiers still need to be trusted or validated. |
| Generate LIKE values | `BuildLikeValue` | Escape user search terms at the application boundary if wildcards should not be user-controlled. |
| Build write payloads | `NewEntity`, `Insert`, `Update`, `ConditionsFromEntity` | Entity table and column names are identifiers, not data values. |
| Apply pagination | `NewPage`, `NewPageResult`, `Page`, `Asc`, `Desc` | Always include a stable order when paging mutable tables. |
| Use dialect placeholders and wrappers | `WithDialect`, `WithWrapper`, `WrapperForDialect`, dialect constants | Match placeholder style to the target driver. |
| Parse named parameters | `ParseNamed`, `ExecNamed` | Named args are converted to dialect-specific placeholders. |
| Open or wrap database/sql handles | `Open`, `Use`, `WithMaxOpenConns`, `WithMaxIdleConns`, `WithConnMaxLifetime`, `WithConnMaxIdleTime` | Configure pools before production use. |
| Execute through sessions | `Exec`, `Session` methods, transaction helpers on `DB` | Pass `context.Context` through every database call. |
| Validate identifier escape hatches | `IsSafeIdentifier`, `Raw` | `Raw` is for trusted SQL fragments only. |

## SQL safety checklist

- Never concatenate user input into SQL. Put data values in builder conditions, named args, or `Exec` args so they become placeholders.
- Treat table names, column names, sort fields, and raw fragments as identifiers. Validate them with an allow-list; `IsSafeIdentifier` is only a guardrail, not authorization.
- Use `Raw` only for trusted static fragments. Prefer builder APIs for dynamic conditions and values.
- Always propagate caller contexts into `Exec`, query, and transaction operations so cancellations and deadlines reach the driver.
- Configure connection pools with explicit max open, max idle, lifetime, and idle-time values before production use.
- Handle `sql.ErrNoRows` or domain-level “not found” cases separately from infrastructure errors.
- Close database handles you open, and keep transaction boundaries small and explicit.
- Use stable `ORDER BY` fields for pagination to avoid missing or duplicated rows when data changes between pages.

## When not to use vdb

- Use `database/sql`, `sqlx`, or `pgx` directly when you need advanced scanning, COPY/bulk APIs, driver-specific features, or fine-grained transaction isolation options.
- Use a reviewed migration tool for schema creation and changes. `vdb` builders are application-query helpers, not migration generators.
- Use explicit handwritten SQL when the query is complex enough that a fluent builder hides optimizer-relevant structure.

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

## Related packages

- Use `vconf` when database connection settings come from layered configuration.
- Use `vjson` when query results or fixtures need JSON encoding and inspection.
- Use `verr` and `vlog` when database errors need wrapping and structured diagnostics.

## Benchmarks and trade-offs

Run the focused SQL builder benchmarks when changing query construction behavior:

```bash
go test -bench=. -benchmem -run=^$ ./internal/db ./vdb
```

The benchmark suite covers paged-order query generation. Builder overhead is usually small compared with database round trips, but it still allocates strings and argument slices. For hot paths, measure the exact query shape and keep raw SQL fragments static and reviewed.

## FAQ

### Are builder-generated queries safe from SQL injection?

Values passed through conditions and args are parameterized. Identifiers such as table names, column names, and sort fields are not data values; validate them with an allow-list before passing them to builders.

### Can I pass user-selected sort columns to `OrderBy`?

Only after mapping user choices to known column constants. Do not pass arbitrary request strings as identifiers.

### Should I use `Raw` for dynamic SQL?

Use `Raw` only for trusted static fragments. If dynamic data is needed, keep the SQL fragment fixed and pass values as args.

### Does `vdb` manage migrations?

No. Use external migration tooling and human-reviewed migration SQL for schema changes.
