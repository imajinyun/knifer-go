package vxml

import (
	"testing"
)

func TestFacadeCreateXMLWithRootNS(t *testing.T) {
	doc := CreateXMLWithRootNS("root", "urn:example")
	if doc == nil {
		t.Fatal("CreateXMLWithRootNS returned nil")
	}
	root := GetRootElement(doc)
	if root == nil {
		t.Fatal("root element is nil")
	}
}

func TestFacadeCleanInvalidAndComment(t *testing.T) {
	if got := CleanInvalid("hello\x00world"); got != "helloworld" {
		t.Fatalf("CleanInvalid = %q", got)
	}
	if got := CleanComment("a<!-- comment -->b"); got != "ab" {
		t.Fatalf("CleanComment = %q", got)
	}
}

func TestFacadeGetElements(t *testing.T) {
	doc, err := ReadXML("<root><child>value</child><other>other</other></root>")
	if err != nil {
		t.Fatalf("ReadXML: %v", err)
	}
	root := GetRootElement(doc)
	children := GetElements(root, "")
	if len(children) != 2 {
		t.Fatalf("GetElements(root, '') = %d, want 2", len(children))
	}
	children = GetElements(root, "child")
	if len(children) != 1 {
		t.Fatalf("GetElements(root, 'child') = %d, want 1", len(children))
	}
}

func TestFacadeXMLToMapInto(t *testing.T) {
	result := map[string]any{}
	got, err := XMLToMapInto("<a><b>1</b></a>", result)
	if err != nil {
		t.Fatalf("XMLToMapInto: %v", err)
	}
	if got["a"] == nil {
		t.Fatal("XMLToMapInto result missing key 'a'")
	}
}

func TestFacadeWriteOptionConstructors(t *testing.T) {
	_ = WithCharset("UTF-8")
	_ = WithPretty()
}

func TestFacadeCleanOptionConstructors(t *testing.T) {
	_ = WithInvalidRegexp(nil)
	_ = WithCommentRegexp(nil)
}

func TestFacadeStringSetters(t *testing.T) {
	input := `<root><item id="1">hello</item></root>`
	decoded := make(map[string]any)
	_, err := XMLToMapInto(input, decoded)
	if err != nil {
		t.Fatalf("XMLToMapInto: %v", err)
	}

	out, err := XMLToMap(input)
	if err != nil || out == nil {
		t.Fatalf("XMLToMap = %v, %v", out, err)
	}
	_ = XMLNodeToMap(GetRootElement(CreateXMLWithRoot("x")))

	var dst struct {
		XMLName struct{} `xml:"root"`
		Item    string   `xml:"item"`
	}
	if err := XMLToBean(input, &dst); err != nil {
		t.Fatalf("XMLToBean: %v", err)
	}
	_ = XMLNodeToBean(GetRootElement(CreateXMLWithRoot("x")), &dst)
}

func TestFacadeCreateXML(t *testing.T) {
	doc := CreateXML()
	if doc == nil {
		t.Fatal("CreateXML returned nil")
	}
}

func TestFacadeGetOwnerDocument(t *testing.T) {
	doc := CreateXMLWithRoot("root")
	root := GetRootElement(doc)
	owner := GetOwnerDocument(root)
	if owner == nil {
		t.Fatal("GetOwnerDocument returned nil")
	}
}

func TestFacadeCleanWithOptions(t *testing.T) {
	if got := CleanInvalidWithOptions("hello\x00world"); got != "helloworld" {
		t.Fatalf("CleanInvalidWithOptions = %q", got)
	}
	if got := CleanCommentWithOptions("a<!-- c -->b"); got != "ab" {
		t.Fatalf("CleanCommentWithOptions = %q", got)
	}
}

func TestFacadeReadXML(t *testing.T) {
	doc, err := ReadXML("<a><b>1</b></a>")
	if err != nil || doc == nil {
		t.Fatalf("ReadXML = %v, %v", doc, err)
	}
}

func TestFacadeMapRoundtrip(t *testing.T) {
	m, err := XMLToMap("<root><x>y</x></root>")
	if err != nil || m == nil {
		t.Fatalf("XMLToMap = %v, %v", m, err)
	}
	b, err := MarshalMap(m)
	if err != nil || len(b) == 0 {
		t.Fatalf("MarshalMap = %q, %v", b, err)
	}
}
