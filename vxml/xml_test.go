package vxml

import (
	stdxml "encoding/xml"
	"errors"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	knifer "github.com/imajinyun/go-knifer"
)

type facadeBean struct {
	Name  string  `xml:"name" json:"name"`
	Age   int     `xml:"age" json:"age"`
	Empty *string `xml:"empty" json:"empty"`
}

func TestFacadeXMLErrorContract(t *testing.T) {
	_, err := ParseXML(`<root><unclosed></root>`)
	if err == nil {
		t.Fatal("ParseXML() error = nil, want invalid input")
	}
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("errors.Is(err, ErrCodeInvalidInput) = false: %v", err)
	}
	code, ok := knifer.CodeOf(err)
	if !ok || code != knifer.ErrCodeInvalidInput {
		t.Fatalf("CodeOf(err) = %q, %v; want invalid input", code, ok)
	}
	var xmlErr *Error
	if !errors.As(err, &xmlErr) {
		t.Fatalf("errors.As(err, *vxml.Error) = false: %v", err)
	}
}

func TestFacadeXMLUtilities(t *testing.T) {
	doc, err := ParseXML(`<root><name>alice</name></root>`)
	if err != nil {
		t.Fatalf("ParseXML failed: %v", err)
	}
	if ElementText(GetRootElement(doc), "name") != "alice" {
		t.Fatal("ElementText facade failed")
	}
	AppendChild(GetRootElement(doc), "age")
	AppendText(GetElement(GetRootElement(doc), "age"), 30)
	out, err := MarshalString(doc, WithOmitDeclaration(true))
	if err != nil || !strings.Contains(out, `<age>30</age>`) {
		t.Fatalf("MarshalString facade = %q, %v", out, err)
	}
	m, err := XMLToMap(out)
	if err != nil || m["root"] == nil {
		t.Fatalf("XMLToMap facade = %#v, %v", m, err)
	}
	back, err := MarshalMap(map[string]any{"name": "bob"}, WithRootName("user"), WithOmitDeclaration(true))
	if err != nil || back != `<user><name>bob</name></user>` {
		t.Fatalf("MarshalMap facade = %q, %v", back, err)
	}
	if Escape("<x>") != "&lt;x&gt;" || Unescape("&lt;x&gt;") != "<x>" {
		t.Fatal("escape facade failed")
	}
	if XMLName("local") != (stdxml.Name{Local: "local"}) {
		t.Fatal("XMLName facade failed")
	}
}

func TestFacadeParseAndReadOptions(t *testing.T) {
	doc, err := ParseXML(`<root xmlns:p="urn:p"><p:a>1</p:a></root>`, WithNamespaceAware(false))
	if err != nil {
		t.Fatalf("ParseXML facade failed: %v", err)
	}
	child := GetElement(GetRootElement(doc), "a")
	if child == nil || child.Name.Space != "" {
		t.Fatalf("namespace option facade not applied: %#v", child)
	}

	doc, err = ReadXMLBytes([]byte(`<root><b>2</b></root>`))
	if err != nil || ElementText(doc.Root, "b") != "2" {
		t.Fatalf("ReadXMLBytes facade doc=%#v err=%v", doc, err)
	}
	doc, err = ReadXMLReader(strings.NewReader(`<root><c>3</c></root>`))
	if err != nil || ElementText(doc.Root, "c") != "3" {
		t.Fatalf("ReadXMLReader facade doc=%#v err=%v", doc, err)
	}
	tmp := t.TempDir() + "/in.xml"
	if err := os.WriteFile(tmp, []byte(`<root><d>4</d></root>`), 0o600); err != nil {
		t.Fatal(err)
	}
	doc, err = ReadXMLFile(tmp)
	if err != nil || ElementText(doc.Root, "d") != "4" {
		t.Fatalf("ReadXMLFile facade doc=%#v err=%v", doc, err)
	}
	doc, err = ReadXML(tmp)
	if err != nil || ElementText(doc.Root, "d") != "4" {
		t.Fatalf("ReadXML path facade doc=%#v err=%v", doc, err)
	}
}

func TestFacadeSAXTransformWriteAndFormat(t *testing.T) {
	var starts []string
	if err := ReadBySAX(strings.NewReader(`<root><a>1</a></root>`), func(tok stdxml.Token) error {
		if start, ok := tok.(stdxml.StartElement); ok {
			starts = append(starts, start.Name.Local)
		}
		return nil
	}); err != nil || !reflect.DeepEqual(starts, []string{"root", "a"}) {
		t.Fatalf("ReadBySAX facade starts=%v err=%v", starts, err)
	}
	var nsStarts []stdxml.Name
	if err := ReadBySAXWithOptions(strings.NewReader(`<root xmlns:p="urn:p"><p:a>1</p:a></root>`), func(tok stdxml.Token) error {
		if start, ok := tok.(stdxml.StartElement); ok {
			nsStarts = append(nsStarts, start.Name)
		}
		return nil
	}, WithNamespaceAware(false)); err != nil {
		t.Fatalf("ReadBySAXWithOptions facade: %v", err)
	}
	if !reflect.DeepEqual(nsStarts, []stdxml.Name{{Local: "root"}, {Local: "a"}}) {
		t.Fatalf("ReadBySAXWithOptions facade names=%#v", nsStarts)
	}
	tmp := t.TempDir() + "/sax.xml"
	if err := os.WriteFile(tmp, []byte(`<root><a>1</a></root>`), 0o600); err != nil {
		t.Fatal(err)
	}
	starts = nil
	if err := ReadBySAXFile(tmp, func(tok stdxml.Token) error {
		if start, ok := tok.(stdxml.StartElement); ok {
			starts = append(starts, start.Name.Local)
		}
		return nil
	}); err != nil || !reflect.DeepEqual(starts, []string{"root", "a"}) {
		t.Fatalf("ReadBySAXFile facade starts=%v err=%v", starts, err)
	}
	starts = nil
	if err := ReadBySAXFileWithOptions(tmp, func(tok stdxml.Token) error {
		if start, ok := tok.(stdxml.StartElement); ok {
			starts = append(starts, start.Name.Local)
		}
		return nil
	}, WithStrict(true)); err != nil || !reflect.DeepEqual(starts, []string{"root", "a"}) {
		t.Fatalf("ReadBySAXFileWithOptions facade starts=%v err=%v", starts, err)
	}

	var out strings.Builder
	if err := TransformWith(strings.NewReader(`<root><a>1</a></root>`), &out, WithOmitDeclaration(true)); err != nil || out.String() != `<root><a>1</a></root>` {
		t.Fatalf("TransformWith facade = %q, %v", out.String(), err)
	}
	formatted, err := Format(`<root><a>1</a></root>`)
	if err != nil || !strings.Contains(formatted, "\n  <a>") {
		t.Fatalf("Format facade = %q, %v", formatted, err)
	}
	writePath := t.TempDir() + "/out.xml"
	if err := WriteFile(writePath, CreateXMLWithRoot("root"), WithOmitDeclaration(true)); err != nil {
		t.Fatalf("WriteFile facade failed: %v", err)
	}
	data, err := os.ReadFile(writePath)
	if err != nil || string(data) != `<root/>` {
		t.Fatalf("WriteFile facade content=%q err=%v", data, err)
	}
	if err := WriteFile(writePath, CreateXMLWithRoot("root"), WithOverwrite(false)); err == nil {
		t.Fatal("WriteFile should reject overwrite when disabled")
	}
	missingParent := t.TempDir() + "/missing/out.xml"
	if err := WriteFile(missingParent, CreateXMLWithRoot("root"), WithCreateParents(false)); err == nil {
		t.Fatal("WriteFile should reject missing parent when parent creation is disabled")
	}
	if err := WriteTo(io.Discard, "unsupported"); err == nil {
		t.Fatal("WriteTo should reject unsupported values")
	}
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

func TestFacadeElementAppendAndGuards(t *testing.T) {
	if CreateXML().Root != nil || IsElement(nil) || GetRootElement(nil) != nil || GetOwnerDocument(nil) != nil {
		t.Fatal("guard helpers failed")
	}
	doc := CreateXMLWithRoot("root")
	Append(doc.Root, map[string]any{"items": []int{1, 2}, "nested": map[string]any{"ok": true}})
	out, err := MarshalString(doc, WithOmitDeclaration(true))
	if err != nil || !strings.Contains(out, `<items>1</items><items>2</items>`) || !strings.Contains(out, `<ok>true</ok>`) {
		t.Fatalf("Append map/slice facade = %q, %v", out, err)
	}
	sliceDoc := CreateXMLWithRoot("root")
	Append(sliceDoc.Root, []string{"a", "b"})
	sliceOut, err := MarshalString(sliceDoc, WithOmitDeclaration(true))
	if err != nil || sliceOut != `<root><element>a</element><element>b</element></root>` {
		t.Fatalf("Append slice facade = %q, %v", sliceOut, err)
	}
	if AppendChild(nil, "x") != nil || AppendText(nil, "x") != nil {
		t.Fatal("nil append helpers should return nil")
	}
	if got := AppendText(CreateXMLWithRoot("r").Root, nil); got == nil || got.Text != "" {
		t.Fatalf("AppendText nil facade = %#v", got)
	}
	if child := AppendChild(doc.Root, "withNS", "urn:child"); child == nil || child.Attr[0].Value != "urn:child" {
		t.Fatalf("AppendChild namespace facade = %#v", child)
	}
	if got := ElementText(doc.Root, "missing", "default"); got != "default" {
		t.Fatalf("ElementText default facade = %q", got)
	}
	if got := TransElements([]*Element{nil, doc.Root}); len(got) != 1 || got[0] != doc.Root {
		t.Fatalf("TransElements facade = %#v", got)
	}
}
