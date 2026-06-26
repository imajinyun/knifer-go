package vxml_test

import (
	"bytes"
	stdxml "encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/imajinyun/knifer-go/vxml"
)

type exampleXMLUser struct {
	Name string `xml:"name" json:"name"`
	Age  int    `xml:"age" json:"age"`
}

func ExampleXMLToMap() {
	m, _ := vxml.XMLToMap(`<user><name>knifer-go</name></user>`)
	user := m["user"].(map[string]any)
	fmt.Println(user["name"])
	// Output: knifer-go
}

func ExampleXMLToMapInto() {
	result := map[string]any{"source": "cache"}
	m, err := vxml.XMLToMapInto(`<user><name>knifer-go</name></user>`, result)
	if err != nil {
		fmt.Println(err)
		return
	}
	user := m["user"].(map[string]any)
	fmt.Println(m["source"], user["name"])
	// Output: cache knifer-go
}

func ExampleXMLToBean() {
	var decoded struct {
		User exampleXMLUser `json:"user"`
	}
	if err := vxml.XMLToBean(`<user><name>knifer-go</name><age>3</age></user>`, &decoded); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(decoded.User.Name, decoded.User.Age)
	// Output: knifer-go 3
}

func ExampleParseXML() {
	doc, err := vxml.ParseXML(`<root><name>knifer-go</name></root>`)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(vxml.GetRootElement(doc).Name.Local)
	// Output: root
}

func ExampleReadXMLBytes() {
	doc, err := vxml.ReadXMLBytes([]byte(`<root><name>knifer-go</name></root>`))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(vxml.ElementText(doc.Root, "name"))
	// Output: knifer-go
}

func ExampleReadXMLReader() {
	doc, err := vxml.ReadXMLReader(strings.NewReader(`<root><name>knifer-go</name></root>`))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(vxml.ElementText(doc.Root, "name"))
	// Output: knifer-go
}

func ExampleReadXMLFile() {
	dir, err := os.MkdirTemp("", "knifer-go-vxml-example-")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "input.xml")
	if err := os.WriteFile(path, []byte(`<root><name>knifer-go</name></root>`), 0o600); err != nil {
		fmt.Println(err)
		return
	}
	doc, err := vxml.ReadXMLFile(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(vxml.ElementText(doc.Root, "name"))
	// Output: knifer-go
}

func ExampleReadBySAX() {
	names := make([]string, 0)
	if err := vxml.ReadBySAX(strings.NewReader(`<root><item>one</item></root>`), func(tok stdxml.Token) error {
		if start, ok := tok.(stdxml.StartElement); ok {
			names = append(names, start.Name.Local)
		}
		return nil
	}); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(names)
	// Output: [root item]
}

func ExampleCreateXMLWithRoot() {
	doc := vxml.CreateXMLWithRoot("root")
	text := vxml.AppendChild(doc.Root, "name")
	vxml.AppendText(text, "knifer-go")

	xmlStr, err := vxml.MarshalString(doc, vxml.WithOmitDeclaration(true))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(xmlStr)
	// Output: <root><name>knifer-go</name></root>
}

func ExampleCreateXMLWithRootNS() {
	doc := vxml.CreateXMLWithRootNS("feed", "urn:feeds")
	xmlStr, err := vxml.MarshalString(doc, vxml.WithOmitDeclaration(true))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(xmlStr)
	// Output: <feed xmlns="urn:feeds"/>
}

func ExampleAppend() {
	doc := vxml.CreateXMLWithRoot("user")
	vxml.Append(doc.Root, map[string]any{"name": "knifer-go", "age": 3})

	xmlStr, err := vxml.MarshalString(doc, vxml.WithOmitDeclaration(true))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(xmlStr)
	// Output: <user><age>3</age><name>knifer-go</name></user>
}

func ExampleGetElements() {
	doc, _ := vxml.ParseXML(`<root><item>one</item><item>two</item><meta>ok</meta></root>`)
	items := vxml.GetElements(doc.Root, "item")
	fmt.Println(len(items), strings.TrimSpace(items[1].Text))
	// Output: 2 two
}

func ExampleElementText() {
	doc, _ := vxml.ParseXML(`<root><name>knifer-go</name></root>`)
	fmt.Println(vxml.ElementText(doc.Root, "name"))
	fmt.Println(vxml.ElementText(doc.Root, "missing", "default"))
	// Output:
	// knifer-go
	// default
}

func ExampleGetByXPath() {
	doc, _ := vxml.ParseXML(`<root><item>one</item><item>two</item></root>`)
	fmt.Println(vxml.GetByXPath("/root/item", doc, "text"))
	fmt.Println(len(vxml.GetByXPath("/root/item", doc, "nodes").([]*vxml.Element)))
	// Output:
	// one
	// 2
}

func ExampleFormatWithOptions() {
	formatted, _ := vxml.FormatWithOptions(`<root><name>knifer-go</name></root>`, vxml.WithFormatWriteOptions(vxml.WithOmitDeclaration(true)))
	fmt.Println(strings.Contains(formatted, "\n  <name>"))
	// Output: true
}

func ExampleMarshalString() {
	doc := vxml.CreateXMLWithRoot("root")
	xmlStr, err := vxml.MarshalString(doc, vxml.WithOmitDeclaration(true))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(xmlStr)
	// Output: <root/>
}

func ExampleWriteTo() {
	doc := vxml.CreateXMLWithRoot("root")
	var out bytes.Buffer
	if err := vxml.WriteTo(&out, doc, vxml.WithOmitDeclaration(true)); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(out.String())
	// Output: <root/>
}

func ExampleMarshalBean() {
	xmlStr, err := vxml.MarshalBean(
		exampleXMLUser{Name: "knifer-go", Age: 3},
		vxml.WithRootName("user"),
		vxml.WithOmitDeclaration(true),
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(xmlStr)
	// Output: <user><age>3</age><name>knifer-go</name></user>
}

func ExampleTransformWith() {
	var out bytes.Buffer
	if err := vxml.TransformWith(
		strings.NewReader(`<root><name>knifer-go</name></root>`),
		&out,
		vxml.WithOmitDeclaration(true),
	); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(out.String())
	// Output: <root><name>knifer-go</name></root>
}

func ExampleEscape() {
	fmt.Println(vxml.Escape("<name>knifer-go</name>"))
	// Output: &lt;name&gt;knifer-go&lt;/name&gt;
}

func ExampleUnescape() {
	fmt.Println(vxml.Unescape("&lt;name&gt;knifer-go&lt;/name&gt;"))
	// Output: <name>knifer-go</name>
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

func ExampleCleanInvalid() {
	fmt.Println(vxml.CleanInvalid("a\x00b"))
	// Output: ab
}
