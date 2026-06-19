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

func ExampleEscape() {
	fmt.Println(vxml.Escape("<name>go-knifer</name>"))
	// Output: &lt;name&gt;go-knifer&lt;/name&gt;
}

func ExampleMarshalMap() {
	xmlStr, err := vxml.MarshalMap(
		map[string]any{"name": "bob"},
		vxml.WithRootName("user"),
		vxml.WithOmitDeclaration(true),
	)

	fmt.Println(xmlStr)
	fmt.Println(err)
	// Output:
	// <user><name>bob</name></user>
	// <nil>
}

func ExampleCleanComment() {
	fmt.Println(vxml.CleanComment("a<!-- hidden -->b"))
	// Output: ab
}
