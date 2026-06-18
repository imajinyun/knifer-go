package vbean_test

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vbean"
)

func ExampleToMap() {
	type User struct {
		Name string
		Age  int
	}

	u := User{Name: "Alice", Age: 30}
	m, _ := vbean.ToMap(u)
	fmt.Println(m["Name"], m["Age"])
	// Output: Alice 30
}

func ExampleToStruct() {
	m := map[string]any{"name": "Bob", "age": 25}

	type User struct {
		Name string
		Age  int
	}

	var u User
	_ = vbean.ToStruct(m, &u)
	fmt.Printf("%s is %d", u.Name, u.Age)
	// Output: Bob is 25
}
