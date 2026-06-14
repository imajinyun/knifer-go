package vxml

import (
	"strings"
	"testing"
)

type facadeBean struct {
	Name  string  `xml:"name" json:"name"`
	Age   int     `xml:"age" json:"age"`
	Empty *string `xml:"empty" json:"empty"`
}

func TestFacadeMapBeanXPathAndNamespace(t *testing.T) {
	xmlStr, err := MarshalBean(facadeBean{Name: "bob", Age: 20}, WithRootName("user"), WithIgnoreNullFields(true), WithOmitDeclaration(true))
	if err != nil || strings.Contains(xmlStr, "empty") || !strings.Contains(xmlStr, `<name>bob</name>`) {
		t.Fatalf("MarshalBean facade = %q, %v", xmlStr, err)
	}
	var decoded struct {
		Root struct {
			Name string `json:"name"`
		} `json:"root"`
	}
	if err := XMLToBean(`<root><name>alice</name></root>`, &decoded); err != nil || decoded.Root.Name != "alice" {
		t.Fatalf("XMLToBean facade decoded=%+v err=%v", decoded, err)
	}
	doc, err := ParseXML(`<root xmlns="urn:default" xmlns:p="urn:p"><p:a>1</p:a><p:a>2</p:a></root>`)
	if err != nil {
		t.Fatal(err)
	}
	if got := GetByXPath("//a", doc, "nodes"); len(got.([]*Element)) != 2 {
		t.Fatalf("GetByXPath nodes facade = %#v", got)
	}
	if got := GetNodeByXPath("/root/a", doc); got == nil || strings.TrimSpace(got.Text) != "1" {
		t.Fatalf("GetNodeByXPath facade = %#v", got)
	}
	if got := GetElementByXPath("/root/a", doc); got == nil || strings.TrimSpace(got.Text) != "1" {
		t.Fatalf("GetElementByXPath facade = %#v", got)
	}
	m := XMLNodeToMap(doc.Root)
	if m["root"] == nil {
		t.Fatalf("XMLNodeToMap facade = %#v", m)
	}
	merged := XMLNodeToMapInto(doc.Root, map[string]any{"old": true})
	if merged["old"] != true || merged["root"] == nil {
		t.Fatalf("XMLNodeToMapInto facade = %#v", merged)
	}
	var decodedNode struct {
		Root struct {
			A []any `json:"a"`
		} `json:"root"`
	}
	if err := XMLNodeToBean(doc.Root, &decodedNode); err != nil || len(decodedNode.Root.A) != 2 {
		t.Fatalf("XMLNodeToBean facade decoded=%+v err=%v", decodedNode, err)
	}
	cache := NewNamespaceCache(doc)
	if cache.NamespaceURI("") != "urn:default" || cache.NamespaceURI("p") != "urn:p" || cache.PrefixOf("urn:p") != "p" {
		t.Fatalf("namespace cache facade = %#v", cache)
	}
	if (*NamespaceCache)(nil).NamespaceURI("p") != "" || (*NamespaceCache)(nil).PrefixOf("urn:p") != "" {
		t.Fatal("nil namespace cache should return empty values")
	}
}
