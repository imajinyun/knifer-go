package vdb_test

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vdb"
)

func ExampleSelect() {
	b := vdb.Select("id", "name").From("users").Where(vdb.Gt("age", 18))
	sql, args, _ := b.SQL()
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// SELECT id, name FROM users WHERE age > ?
	// [18]
}

func ExampleNewEntity() {
	e := vdb.NewEntity("users")
	e.Values["name"] = "Alice"
	e.Values["age"] = 30
	b := vdb.Insert(e)
	sql, args, _ := b.SQL()
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// INSERT INTO users (age, name) VALUES (?, ?)
	// [30 Alice]
}

func ExampleBuildConditions() {
	sql, args, err := vdb.BuildConditions(
		vdb.Like("name", vdb.BuildLikeValue("go", "prefix")),
		vdb.In("role", "admin", "owner"),
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// name LIKE ? AND role IN (?, ?)
	// [go% admin owner]
}

func ExampleDelete() {
	b := vdb.Delete("users").Where(vdb.IsNull("deleted_at"))
	sql, args, _ := b.SQL()
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// DELETE FROM users WHERE deleted_at IS NULL
	// []
}

func ExampleBuildLikeValue() {
	fmt.Println(vdb.BuildLikeValue("go", "contains"))
	// Output: %go%
}

func ExampleNewPage() {
	sql, args, _ := vdb.NewBuilder(vdb.WithDialect(vdb.DialectMySQL)).
		Select("id", "created_at").
		From("orders").
		Page(vdb.NewPage(2, 10, vdb.Desc("created_at"))).
		SQL()
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// SELECT `id`, `created_at` FROM `orders` ORDER BY `created_at` DESC LIMIT 10 OFFSET 10
	// []
}

func ExampleIsSafeIdentifier() {
	fmt.Println(vdb.IsSafeIdentifier("orders.created_at"))
	fmt.Println(vdb.IsSafeIdentifier("orders; drop table orders"))
	// Output:
	// true
	// false
}

func ExampleParseNamed() {
	named, err := vdb.ParseNamed(
		"SELECT * FROM users WHERE id = :id::int AND status = :status",
		map[string]any{"id": 7, "status": "active"},
		vdb.DialectPostgres,
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(named.SQL)
	fmt.Println(named.Params)
	fmt.Println(named.Names)
	// Output:
	// SELECT * FROM users WHERE id = $1::int AND status = $2
	// [7 active]
	// [id status]
}

func ExampleUpdate() {
	entity := vdb.EntityFromMap("users", map[string]any{"active": false})
	sql, args, _ := vdb.Update(entity).Where(vdb.Eq("id", 7)).SQL()
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// UPDATE users SET active = ? WHERE id = ?
	// [false 7]
}

func ExampleOrGroup() {
	sql, args, _ := vdb.BuildConditions(
		vdb.AndGroup(vdb.Eq("tenant_id", 42)),
		vdb.OrGroup(vdb.Eq("status", "active"), vdb.Eq("status", "pending")),
	)
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// (tenant_id = ?) OR (status = ? AND status = ?)
	// [42 active pending]
}
