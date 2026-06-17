package vxml_test

import (
	"fmt"
	"strings"

	"github.com/imajinyun/go-knifer/vxml"
)

func ExampleXMLToMap() {
	m, _ := vxml.XMLToMap(`<user><name>go-knifer</name></user>`)
	user := m["user"].(map[string]any)
	fmt.Println(user["name"])
	// Output: go-knifer
}

func ExampleFormatWithOptions() {
	formatted, _ := vxml.FormatWithOptions(`<root><name>go-knifer</name></root>`, vxml.WithFormatWriteOptions(vxml.WithOmitDeclaration(true)))
	fmt.Println(strings.Contains(formatted, "\n  <name>"))
	// Output: true
}
