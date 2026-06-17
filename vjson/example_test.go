package vjson_test

import (
	"errors"
	"fmt"

	"github.com/imajinyun/go-knifer"
	"github.com/imajinyun/go-knifer/vjson"
)

func ExampleToStr() {
	s, _ := vjson.ToStr(map[string]any{"name": "go"})
	fmt.Println(s)
	// Output: {"name":"go"}
}

func ExampleIsJSON() {
	fmt.Println(vjson.IsJSON(`{"a":1}`))
	fmt.Println(vjson.IsJSON(`not json`))
	// Output:
	// true
	// false
}

func ExampleGetByPath() {
	root, _ := vjson.Parse(`{"user":{"name":"go"}}`)
	fmt.Println(vjson.GetByPath(root, "user.name"))
	// Output: go
}

func ExampleParseObj_error() {
	_, err := vjson.ParseObj(`[1,2,3]`)
	fmt.Println(errors.Is(err, knifer.ErrCodeInvalidInput))
	// Output: true
}

func ExamplePutByPath() {
	root := vjson.NewObject()
	_ = vjson.PutByPath(root, "user.name", "go-knifer")
	fmt.Println(vjson.GetByPath(root, "user.name"))
	// Output: go-knifer
}

func ExampleToBean() {
	type user struct {
		Name string `json:"name"`
	}
	var u user
	_ = vjson.ToBean(`{"name":"go-knifer"}`, &u)
	fmt.Println(u.Name)
	// Output: go-knifer
}

func ExampleXMLToJSON() {
	obj, _ := vjson.XMLToJSON(`<user><name>go-knifer</name></user>`)
	fmt.Println(vjson.GetByPath(obj, "user.name"))
	// Output: go-knifer
}
