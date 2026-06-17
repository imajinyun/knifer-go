package vxml

import (
	stdxml "encoding/xml"
	"errors"
	"io"
	"regexp"
	"strconv"
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

func TestFacadeXMLMapBeanAndXPathOptionWrappers(t *testing.T) {
	m, err := XMLToMapWithOptions(`<root><n>42</n><f>1.5</f></root>`,
		WithScalarIntParser(func(s string, base, bitSize int) (int64, error) {
			if s == "42" {
				return 7, nil
			}
			return strconv.ParseInt(s, base, bitSize)
		}),
		WithScalarFloatParser(func(string, int) (float64, error) { return 2.5, nil }),
	)
	if err != nil {
		t.Fatalf("XMLToMapWithOptions: %v", err)
	}
	root := m["root"].(map[string]any)
	if root["n"] != int64(7) || root["f"] != 2.5 {
		t.Fatalf("custom scalar parsers not applied: %#v", root)
	}

	doc, err := ParseXML(`<root><n>42</n><f>1.5</f><item>a</item><item>b</item></root>`)
	if err != nil {
		t.Fatal(err)
	}
	nodeMap := XMLNodeToMapWithOptions(doc.Root, WithScalarIntParser(func(string, int, int) (int64, error) { return 9, nil }))
	if nodeMap["root"].(map[string]any)["n"] != int64(9) {
		t.Fatalf("XMLNodeToMapWithOptions = %#v", nodeMap)
	}
	merged, err := XMLToMapIntoWithOptions(`<root><n>1</n></root>`, nil, WithScalarIntParser(func(string, int, int) (int64, error) { return 5, nil }))
	if err != nil || merged["root"].(map[string]any)["n"] != int64(5) {
		t.Fatalf("XMLToMapIntoWithOptions = %#v err=%v", merged, err)
	}
	nodeMerged := XMLNodeToMapIntoWithOptions(doc.Root, map[string]any{"old": true}, WithScalarFloatParser(func(string, int) (float64, error) { return 8.5, nil }))
	if nodeMerged["old"] != true || nodeMerged["root"].(map[string]any)["f"] != 8.5 {
		t.Fatalf("XMLNodeToMapIntoWithOptions = %#v", nodeMerged)
	}

	var decoded struct {
		Root struct {
			N int64 `json:"n"`
		} `json:"root"`
	}
	if err := XMLToBeanWithOptions(`<root><n>3</n></root>`, &decoded, WithScalarIntParser(func(string, int, int) (int64, error) { return 11, nil })); err != nil {
		t.Fatalf("XMLToBeanWithOptions: %v", err)
	}
	if decoded.Root.N != 11 {
		t.Fatalf("XMLToBeanWithOptions decoded=%+v", decoded)
	}

	beanCalled := false
	var custom struct{ Name string }
	if err := XMLNodeToBeanWithOptions(doc.Root, &custom, WithBeanUnmarshalFunc(func(_ []byte, dst any) error {
		beanCalled = true
		dst.(*struct{ Name string }).Name = "custom"
		return nil
	})); err != nil || !beanCalled || custom.Name != "custom" {
		t.Fatalf("XMLNodeToBeanWithOptions custom=%+v called=%v err=%v", custom, beanCalled, err)
	}

	marshalErr := errors.New("marshal failed")
	if err := XMLNodeToBeanWithParseOptions(doc.Root, &custom, nil, WithBeanMarshalFunc(func(any) ([]byte, error) { return nil, marshalErr })); !errors.Is(err, marshalErr) {
		t.Fatalf("XMLNodeToBeanWithParseOptions marshal err = %v", err)
	}

	nodes := GetNodeListByXPath("/root/item", doc)
	if len(nodes) != 2 || strings.TrimSpace(nodes[1].Text) != "b" {
		t.Fatalf("GetNodeListByXPath = %#v", nodes)
	}

	if got := CleanInvalidWithOptions("a!b", WithInvalidRegexp(regexp.MustCompile("!"))); got != "ab" {
		t.Fatalf("CleanInvalid WithInvalidRegexp = %q", got)
	}
	if got := CleanCommentWithOptions("a<!--x-->b", WithCommentRegexp(regexp.MustCompile("<!--x-->"))); got != "ab" {
		t.Fatalf("CleanComment WithCommentRegexp = %q", got)
	}
}

func TestFacadeXMLReaderOptions(t *testing.T) {
	charsetCalled := false
	doc, err := ParseXML(`<?xml version="1.0" encoding="x-test"?><root>&custom;</root>`,
		WithStrict(false),
		WithCharsetReader(func(charset string, input io.Reader) (io.Reader, error) {
			charsetCalled = charset == "x-test"
			return input, nil
		}),
		WithEntity(map[string]string{"custom": "value"}),
	)
	if err != nil {
		t.Fatalf("ParseXML with charset/entity: %v", err)
	}
	if !charsetCalled || strings.TrimSpace(doc.Root.Text) != "value" {
		t.Fatalf("charsetCalled=%v text=%q", charsetCalled, doc.Root.Text)
	}

	decoderCalled := false
	doc, err = ReadXMLReader(strings.NewReader(`<ignored/>`), WithDecoderFactory(func(io.Reader) *stdxml.Decoder {
		decoderCalled = true
		return stdxml.NewDecoder(strings.NewReader(`<provided/>`))
	}))
	if err != nil {
		t.Fatalf("ReadXMLReader WithDecoderFactory: %v", err)
	}
	if !decoderCalled || doc.Root.Name.Local != "provided" {
		t.Fatalf("decoderCalled=%v root=%s", decoderCalled, doc.Root.Name.Local)
	}
	if _, err := ReadXMLReader(strings.NewReader(`<root><a>1</a></root>`), WithMaxBytes(8)); err == nil {
		t.Fatal("WithMaxBytes should bound reader input")
	}
}
