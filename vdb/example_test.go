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
