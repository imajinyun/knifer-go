package vxml

import (
	stdxml "encoding/xml"
	"strings"
	"testing"
)

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
