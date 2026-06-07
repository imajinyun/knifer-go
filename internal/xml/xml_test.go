package xml

import (
	stdxml "encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"

	knifer "github.com/imajinyun/go-knifer"
)

type sampleBean struct {
	Name  string  `xml:"name" json:"name"`
	Age   int     `xml:"age" json:"age"`
	Empty *string `xml:"empty" json:"empty"`
}

type failingWriter struct {
	err error
}

func (w failingWriter) Write(_ []byte) (int, error) { return 0, w.err }

func TestReadCreateWriteAndFormat(t *testing.T) {
	doc, err := ParseXML(`<root><name id="1">alice</name><tags>a</tags><tags>b</tags></root>`)
	if err != nil {
		t.Fatalf("ParseXML failed: %v", err)
	}
	root := GetRootElement(doc)
	if root == nil || root.Name.Local != "root" {
		t.Fatalf("root mismatch: %#v", root)
	}
	if got := ElementText(root, "name"); got != "alice" {
		t.Fatalf("ElementText = %q", got)
	}
	if got := GetElements(root, ""); len(got) != 3 {
		t.Fatalf("GetElements all children length = %d", len(got))
	}
	if got := GetElements(root, "tags"); len(got) != 2 {
		t.Fatalf("GetElements tags length = %d", len(got))
	}
	if GetOwnerDocument(GetElement(root, "name")).Root != root {
		t.Fatal("GetOwnerDocument should walk to root")
	}
	plain, err := MarshalString(doc, WithOmitDeclaration(true))
	if err != nil || plain != `<root><name id="1">alice</name><tags>a</tags><tags>b</tags></root>` {
		t.Fatalf("MarshalString = %q, %v", plain, err)
	}
	formatted, err := Format(plain)
	if err != nil || !strings.Contains(formatted, `<?xml version="1.0" encoding="UTF-8"?>`) || !strings.Contains(formatted, "\n  <name") {
		t.Fatalf("Format = %q, %v", formatted, err)
	}
	created := CreateXMLWithRootNS("user", "urn:test")
	AppendChild(created.Root, "name")
	AppendText(GetElement(created.Root, "name"), "bob")
	createdStr, err := MarshalString(created, WithOmitDeclaration(true))
	if err != nil || !strings.Contains(createdStr, `xmlns="urn:test"`) || !strings.Contains(createdStr, `<name>bob</name>`) {
		t.Fatalf("CreateXMLWithRootNS serialized = %q, %v", createdStr, err)
	}
}

func TestReadVariantsSAXXPathAndFile(t *testing.T) {
	if CleanInvalid("a\x00b\x08c") != "abc" {
		t.Fatal("CleanInvalid failed")
	}
	if CleanComment("<a><!-- hidden --><b/></a>") != "<a><b/></a>" {
		t.Fatal("CleanComment failed")
	}
	if got := CleanInvalidWithOptions("aXb", WithInvalidRegexp(regexp.MustCompile(`X`))); got != "ab" {
		t.Fatalf("CleanInvalidWithOptions = %q", got)
	}
	if got := CleanCommentWithOptions("<a>[hidden]<b/></a>", WithCommentRegexp(regexp.MustCompile(`\[[^]]+\]`))); got != "<a><b/></a>" {
		t.Fatalf("CleanCommentWithOptions = %q", got)
	}
	if Escape(`<a&"'>`) != "&lt;a&amp;&#34;&#39;&gt;" || Unescape("&lt;a&amp;&gt;") != "<a&>" {
		t.Fatal("escape/unescape failed")
	}

	fromBytes, err := ReadXMLBytes([]byte(`<root><a>1</a></root>`))
	if err != nil || ElementText(fromBytes.Root, "a") != "1" {
		t.Fatalf("ReadXMLBytes doc=%#v err=%v", fromBytes, err)
	}
	fromReader, err := ReadXMLReader(strings.NewReader(`<root><b>2</b></root>`))
	if err != nil || ElementText(fromReader.Root, "b") != "2" {
		t.Fatalf("ReadXMLReader doc=%#v err=%v", fromReader, err)
	}
	_, err = ParseXML(`<root><unclosed></root>`)
	assertXMLInvalidInput(t, err)
	if _, err := ReadXML(""); err == nil {
		t.Fatal("ReadXML empty path should return error")
	}

	var starts []string
	if err := ReadBySAX(strings.NewReader(`<root><a>1</a></root>`), func(tok stdxml.Token) error {
		if start, ok := tok.(stdxml.StartElement); ok {
			starts = append(starts, start.Name.Local)
		}
		return nil
	}); err != nil || !reflect.DeepEqual(starts, []string{"root", "a"}) {
		t.Fatalf("ReadBySAX starts=%v err=%v", starts, err)
	}
	if err := ReadBySAX(strings.NewReader(`<root>`), nil); err != nil {
		t.Fatalf("ReadBySAX nil handler should ignore input: %v", err)
	}
	var nsStarts []stdxml.Name
	if err := ReadBySAXWithOptions(strings.NewReader(`<root xmlns:p="urn:p"><p:a>1</p:a></root>`), func(tok stdxml.Token) error {
		if start, ok := tok.(stdxml.StartElement); ok {
			nsStarts = append(nsStarts, start.Name)
		}
		return nil
	}, WithNamespaceAware(false)); err != nil {
		t.Fatalf("ReadBySAXWithOptions namespace-aware false: %v", err)
	}
	if !reflect.DeepEqual(nsStarts, []stdxml.Name{{Local: "root"}, {Local: "a"}}) {
		t.Fatalf("ReadBySAXWithOptions names = %#v", nsStarts)
	}
	assertXMLInvalidInput(t, ReadBySAXWithOptions(strings.NewReader(`<root><a/></root>`), func(stdxml.Token) error { return nil }, WithMaxBytes(6)))
	handlerErr := errors.New("handler stop")
	if err := ReadBySAX(strings.NewReader(`<root/>`), func(stdxml.Token) error { return handlerErr }); !errors.Is(err, handlerErr) {
		t.Fatalf("ReadBySAX handler err = %v", err)
	}
	assertXMLInvalidInput(t, ReadBySAX(strings.NewReader(`<root>`), func(stdxml.Token) error { return nil }))

	tmp := t.TempDir() + "/x.xml"
	if err := os.WriteFile(tmp, []byte(`<root><a>1</a><a>2</a></root>`), 0o600); err != nil {
		t.Fatal(err)
	}
	doc, err := ReadXML(tmp)
	if err != nil {
		t.Fatalf("ReadXML file failed: %v", err)
	}
	if got := GetByXPath("/root/a", doc, "string"); got != "1" {
		t.Fatalf("GetByXPath string = %v", got)
	}
	if got := GetByXPath("/root/a", doc, "nodes"); len(got.([]*Element)) != 2 {
		t.Fatalf("GetByXPath nodes = %#v", got)
	}
	if got := GetElementByXPath("/root/a", doc); got == nil || strings.TrimSpace(got.Text) != "1" {
		t.Fatalf("GetElementByXPath = %#v", got)
	}
	if got := GetNodeByXPath("/root/missing", doc); got != nil {
		t.Fatalf("missing XPath should be nil: %#v", got)
	}
	if got := GetNodeListByXPath("//a", doc); len(got) != 2 {
		t.Fatalf("GetNodeListByXPath = %d", len(got))
	}
	var saxFileStarts []string
	if err := ReadBySAXFile(tmp, func(tok stdxml.Token) error {
		if start, ok := tok.(stdxml.StartElement); ok {
			saxFileStarts = append(saxFileStarts, start.Name.Local)
		}
		return nil
	}); err != nil || !reflect.DeepEqual(saxFileStarts, []string{"root", "a", "a"}) {
		t.Fatalf("ReadBySAXFile starts=%v err=%v", saxFileStarts, err)
	}
	saxFileStarts = nil
	if err := ReadBySAXFileWithOptions(tmp, func(tok stdxml.Token) error {
		if start, ok := tok.(stdxml.StartElement); ok {
			saxFileStarts = append(saxFileStarts, start.Name.Local)
		}
		return nil
	}, WithStrict(true)); err != nil || !reflect.DeepEqual(saxFileStarts, []string{"root", "a", "a"}) {
		t.Fatalf("ReadBySAXFileWithOptions starts=%v err=%v", saxFileStarts, err)
	}

	var out strings.Builder
	if err := TransformWith(strings.NewReader(`<root><a>1</a></root>`), &out, WithOmitDeclaration(true)); err != nil || out.String() != `<root><a>1</a></root>` {
		t.Fatalf("TransformWith = %q, %v", out.String(), err)
	}
	out.Reset()
	if err := TransformWithOptions(strings.NewReader(`<root xmlns:p="urn:p"><p:a>1</p:a></root>`), &out,
		WithTransformParseOptions(WithNamespaceAware(false)),
		WithTransformWriteOptions(WithOmitDeclaration(true)),
	); err != nil || strings.Contains(out.String(), `xmlns`) || !strings.Contains(out.String(), `<a>1</a>`) {
		t.Fatalf("TransformWithOptions = %q, %v", out.String(), err)
	}
}

func TestXMLErrorContract(t *testing.T) {
	_, err := ParseXML("")
	assertXMLInvalidInput(t, err)

	assertXMLInvalidInput(t, WriteTo(nil, CreateXMLWithRoot("root")))
	assertXMLInvalidInput(t, WriteTo(&strings.Builder{}, "unsupported"))

	var dst struct {
		Root struct {
			Value int `json:"value"`
		} `json:"root"`
	}
	assertXMLInvalidInput(t, XMLToBean(`<root><value>not-int</value></root>`, &dst))
}

func assertXMLInvalidInput(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatalf("err = nil, want %s", knifer.ErrCodeInvalidInput)
	}
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("errors.Is(%v, %s) = false", err, knifer.ErrCodeInvalidInput)
	}
	got, ok := knifer.CodeOf(err)
	if !ok || got != knifer.ErrCodeInvalidInput {
		t.Fatalf("CodeOf(%v) = %q, %v; want %q, true", err, got, ok, knifer.ErrCodeInvalidInput)
	}
}

func TestMapBeanAndNamespaceConversions(t *testing.T) {
	m, err := XMLToMap(`<root enabled="true"><name>alice</name><age>30</age><score>3.5</score><none>null</none><tags>a</tags><tags>b</tags></root>`)
	if err != nil {
		t.Fatalf("XMLToMap failed: %v", err)
	}
	root := m["root"].(map[string]any)
	if root["enabled"] != true || root["name"] != "alice" || root["age"] != int64(30) || root["score"] != 3.5 || root["none"] != nil {
		t.Fatalf("XMLToMap root = %#v", root)
	}
	if tags, ok := root["tags"].([]any); !ok || len(tags) != 2 {
		t.Fatalf("XMLToMap tags = %#v", root["tags"])
	}
	merged, err := XMLToMapInto(`<x><a>1</a></x>`, map[string]any{"old": true})
	if err != nil || merged["old"] != true || merged["x"] == nil {
		t.Fatalf("XMLToMapInto = %#v, %v", merged, err)
	}
	stripped, err := XMLToMapWithOptions(`<root xmlns:p="urn:p"><p:a>1</p:a></root>`, WithNamespaceAware(false))
	if err != nil || stripped["root"].(map[string]any)["a"] != int64(1) {
		t.Fatalf("XMLToMapWithOptions = %#v, %v", stripped, err)
	}
	limited, err := XMLToMapIntoWithOptions(`<root><a>1</a></root>`, nil, WithMaxBytes(6))
	if err == nil || limited != nil {
		t.Fatalf("XMLToMapIntoWithOptions should fail with max bytes: %#v, %v", limited, err)
	}
	if got := XMLNodeToMapInto(nil, nil); len(got) != 0 {
		t.Fatalf("XMLNodeToMapInto nil = %#v", got)
	}

	xmlStr, err := MarshalMap(map[string]any{"name": "alice", "tags": []string{"a", "b"}}, WithRootName("user"), WithOmitDeclaration(true))
	if err != nil || xmlStr != `<user><name>alice</name><tags>a</tags><tags>b</tags></user>` {
		t.Fatalf("MarshalMap = %q, %v", xmlStr, err)
	}
	beanStr, err := MarshalBean(sampleBean{Name: "bob", Age: 20}, WithRootName("sample"), WithIgnoreNullFields(true), WithOmitDeclaration(true))
	if err != nil || strings.Contains(beanStr, "empty") || !strings.Contains(beanStr, `<name>bob</name>`) || !strings.Contains(beanStr, `<age>20</age>`) {
		t.Fatalf("MarshalBean = %q, %v", beanStr, err)
	}
	defaultRoot, err := MarshalBean(sampleBean{Name: "typed"}, WithOmitDeclaration(true), WithIgnoreNullFields(true))
	if err != nil || !strings.HasPrefix(defaultRoot, `<samplebean>`) {
		t.Fatalf("MarshalBean default root = %q, %v", defaultRoot, err)
	}

	var decoded struct {
		Root struct {
			Name string `json:"name"`
			Age  int64  `json:"age"`
		} `json:"root"`
	}
	if err := XMLToBean(`<root><name>alice</name><age>30</age></root>`, &decoded); err != nil || decoded.Root.Name != "alice" || decoded.Root.Age != 30 {
		t.Fatalf("XMLToBean decoded=%+v err=%v", decoded, err)
	}
	var decodedNode struct {
		Root struct {
			Name string `json:"name"`
		} `json:"root"`
	}
	doc, err := ParseXML(`<root><name>node</name></root>`)
	if err != nil {
		t.Fatal(err)
	}
	if err := XMLNodeToBean(doc.Root, &decodedNode); err != nil || decodedNode.Root.Name != "node" {
		t.Fatalf("XMLNodeToBean decoded=%+v err=%v", decodedNode, err)
	}
	customCalled := false
	var custom struct{ Name string }
	if err := XMLNodeToBeanWithOptions(doc.Root, &custom, WithBeanUnmarshalFunc(func(_ []byte, dst any) error {
		customCalled = true
		dst.(*struct{ Name string }).Name = "custom"
		return nil
	})); err != nil || !customCalled || custom.Name != "custom" {
		t.Fatalf("XMLNodeToBeanWithOptions custom=%+v called=%v err=%v", custom, customCalled, err)
	}

	nsDoc, err := ParseXML(`<root xmlns="urn:default" xmlns:p="urn:p"><p:a>1</p:a></root>`)
	if err != nil {
		t.Fatal(err)
	}
	cache := NewNamespaceCache(nsDoc)
	if cache.NamespaceURI("") != "urn:default" || cache.NamespaceURI("DEFAULT") != "urn:default" || cache.NamespaceURI("p") != "urn:p" || cache.PrefixOf("urn:p") != "p" {
		t.Fatalf("namespace cache = %#v", cache)
	}
	if (*NamespaceCache)(nil).NamespaceURI("p") != "" || (*NamespaceCache)(nil).PrefixOf("urn:p") != "" {
		t.Fatal("nil namespace cache should return empty values")
	}
}

func TestFormatWithOptions(t *testing.T) {
	formatted, err := FormatWithOptions(`<root xmlns:p="urn:p"><p:a>1</p:a></root>`,
		WithFormatParseOptions(WithNamespaceAware(false)),
		WithFormatWriteOptions(WithOmitDeclaration(true), WithIndent(4)),
	)
	if err != nil || strings.Contains(formatted, `xmlns`) || !strings.Contains(formatted, "\n    <a>") {
		t.Fatalf("FormatWithOptions = %q, %v", formatted, err)
	}
}

func TestPerCallOptionsAndWriteErrors(t *testing.T) {
	doc, err := ParseXML(`<root xmlns:p="urn:p"><p:a>1</p:a></root>`, WithNamespaceAware(false))
	if err != nil {
		t.Fatalf("ParseXML with option failed: %v", err)
	}
	child := GetElement(doc.Root, "a")
	if child == nil || child.Name.Space != "" {
		t.Fatalf("per-call namespace option not applied: %#v", child)
	}

	doc, err = ParseXML(`<root xmlns:p="urn:p"><p:a>1</p:a></root>`)
	if err != nil {
		t.Fatalf("ParseXML failed: %v", err)
	}
	child = GetElement(doc.Root, "a")
	if child == nil || child.Name.Space != "urn:p" {
		t.Fatalf("default namespace awareness should remain enabled: %#v", child)
	}

	pretty, err := MarshalString(CreateXMLWithRoot("root"), WithCharset("GBK"), WithPretty())
	if err != nil || !strings.HasPrefix(pretty, `<?xml version="1.0" encoding="GBK"?>`) || !strings.Contains(pretty, "\n<root/>") {
		t.Fatalf("MarshalString options = %q, %v", pretty, err)
	}
	if err := WriteTo(nil, doc); err == nil || !strings.Contains(err.Error(), "nil writer") {
		t.Fatalf("WriteTo nil writer err = %v", err)
	}
	if err := WriteTo(io.Discard, "unsupported"); err == nil || !strings.Contains(err.Error(), "unsupported node") {
		t.Fatalf("WriteTo unsupported err = %v", err)
	}
	writeErr := errors.New("write failed")
	if err := WriteTo(failingWriter{err: writeErr}, doc); !errors.Is(err, writeErr) {
		t.Fatalf("WriteTo writer err = %v", err)
	}

	tmp := t.TempDir() + "/out.xml"
	if err := WriteFile(tmp, doc, WithOmitDeclaration(true)); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	if err := WriteFile(tmp, doc, WithOmitDeclaration(true), WithOverwrite(false)); err == nil {
		t.Fatalf("WriteFile should reject overwrite when disabled")
	}
	written, err := os.ReadFile(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(written), `<a>1</a>`) {
		t.Fatalf("WriteFile content = %q", written)
	}
	if _, err := ReadXMLReader(strings.NewReader(`<root><a>1</a></root>`), WithMaxBytes(8)); err == nil {
		t.Fatalf("ReadXMLReader should fail when max bytes truncates input")
	}
}

func TestXMLFileProviderOptions(t *testing.T) {
	openedRead := ""
	doc, err := ReadXMLFile("virtual.xml", WithOpenFile(func(path string) (io.ReadCloser, error) {
		openedRead = path
		return io.NopCloser(strings.NewReader(`<root><from>provider</from></root>`)), nil
	}))
	if err != nil || ElementText(doc.Root, "from") != "provider" || openedRead != "virtual.xml" {
		t.Fatalf("ReadXMLFile provider doc=%#v path=%q err=%v", doc, openedRead, err)
	}

	var saxStarts []string
	openedRead = ""
	if err := ReadBySAXFileWithOptions("sax.xml", func(tok stdxml.Token) error {
		if start, ok := tok.(stdxml.StartElement); ok {
			saxStarts = append(saxStarts, start.Name.Local)
		}
		return nil
	}, WithOpenFile(func(path string) (io.ReadCloser, error) {
		openedRead = path
		return io.NopCloser(strings.NewReader(`<root><a/></root>`)), nil
	})); err != nil || !reflect.DeepEqual(saxStarts, []string{"root", "a"}) || openedRead != "sax.xml" {
		t.Fatalf("ReadBySAXFileWithOptions provider starts=%v path=%q err=%v", saxStarts, openedRead, err)
	}

	var mkdirPath string
	var mkdirPerm fs.FileMode
	var openPath string
	var openFlag int
	var openPerm fs.FileMode
	var written strings.Builder
	closer := nopWriteCloser{Writer: &written}
	err = WriteFile("/virtual/out.xml", CreateXMLWithRoot("root"), WithOmitDeclaration(true),
		WithMkdirAll(func(path string, perm fs.FileMode) error {
			mkdirPath, mkdirPerm = path, perm
			return nil
		}),
		WithOpenWriteFile(func(path string, flag int, perm fs.FileMode) (io.WriteCloser, error) {
			openPath, openFlag, openPerm = path, flag, perm
			return closer, nil
		}),
		WithDirPerm(0o700), WithFilePerm(0o600),
	)
	if err != nil {
		t.Fatalf("WriteFile provider: %v", err)
	}
	if mkdirPath != "/virtual" || mkdirPerm != 0o700 || openPath != "/virtual/out.xml" || openPerm != 0o600 || openFlag&os.O_CREATE == 0 || written.String() != `<root/>` {
		t.Fatalf("WriteFile providers mkdir=%q/%v open=%q flag=%#x perm=%v content=%q", mkdirPath, mkdirPerm, openPath, openFlag, openPerm, written.String())
	}
}

type nopWriteCloser struct{ io.Writer }

func (w nopWriteCloser) Close() error { return nil }

func TestAppendAndGuards(t *testing.T) {
	if CreateXML().Root != nil || IsElement(nil) || GetRootElement(nil) != nil || GetOwnerDocument(nil) != nil {
		t.Fatal("guard helpers failed")
	}
	doc := CreateXMLWithRoot("root")
	Append(doc.Root, map[string]any{"items": []int{1, 2}, "nested": map[string]any{"ok": true}})
	got, err := MarshalString(doc, WithOmitDeclaration(true))
	if err != nil || !strings.Contains(got, `<items>1</items><items>2</items>`) || !strings.Contains(got, `<ok>true</ok>`) {
		t.Fatalf("Append map/slice serialized = %q, %v", got, err)
	}
	sliceDoc := CreateXMLWithRoot("root")
	Append(sliceDoc.Root, []string{"a", "b"})
	sliceStr, err := MarshalString(sliceDoc, WithOmitDeclaration(true))
	if err != nil || sliceStr != `<root><element>a</element><element>b</element></root>` {
		t.Fatalf("Append slice serialized = %q, %v", sliceStr, err)
	}
	nestedStructDoc := CreateXMLWithRoot("root")
	Append(nestedStructDoc.Root, map[string]any{"user": sampleBean{Name: "struct", Age: 3}})
	nestedStructStr, err := MarshalString(nestedStructDoc, WithOmitDeclaration(true))
	if err != nil || !strings.Contains(nestedStructStr, `<name>struct</name>`) || !strings.Contains(nestedStructStr, `<age>3</age>`) {
		t.Fatalf("Append nested struct serialized = %q, %v", nestedStructStr, err)
	}
	structDoc := CreateXMLWithRoot("root")
	Append(structDoc.Root, sampleBean{Name: "struct", Age: 3})
	structStr, err := MarshalString(structDoc, WithOmitDeclaration(true))
	if err != nil || structStr != `<root>{struct 3 &lt;nil&gt;}</root>` {
		t.Fatalf("Append root scalar struct serialized = %q, %v", structStr, err)
	}
	if AppendChild(nil, "x") != nil || AppendText(nil, "x") != nil {
		t.Fatal("nil append helpers should return nil")
	}
	if got := AppendText(CreateXMLWithRoot("r").Root, nil); got == nil || got.Text != "" {
		t.Fatalf("AppendText nil = %#v", got)
	}
	if child := AppendChild(doc.Root, "withNS", "urn:child"); child == nil || child.Attr[0].Value != "urn:child" {
		t.Fatalf("AppendChild namespace = %#v", child)
	}
	if got := ElementText(doc.Root, "missing", "default"); got != "default" {
		t.Fatalf("ElementText default = %q", got)
	}
	if got := ElementText(doc.Root, "missing"); got != "" {
		t.Fatalf("ElementText missing = %q", got)
	}
	if got := TransElements([]*Element{nil, doc.Root}); len(got) != 1 || got[0] != doc.Root {
		t.Fatalf("TransElements = %#v", got)
	}
}

func TestInternalHelpers(t *testing.T) {
	if got := typeName(&sampleBean{}); got != "samplebean" {
		t.Fatalf("typeName pointer = %q", got)
	}
	if got := typeName(map[string]any{}); got != "" {
		t.Fatalf("typeName map = %q", got)
	}
	if got, ok := normalizeMap(map[string]string{"a": "b"}); !ok || !gotOK(got, "a", "b") {
		t.Fatalf("normalizeMap string map = %#v", got)
	}
	if got, ok := normalizeMap(map[any]any{"a": 1}); !ok || !gotOK(got, "a", 1) {
		t.Fatalf("normalizeMap any map = %#v", got)
	}
	if got, ok := normalizeMap(map[int]string{1: "one"}); !ok || !gotOK(got, "1", "one") {
		t.Fatalf("normalizeMap typed map = %#v", got)
	}
	if m, ok := normalizeMap(nil); ok || m != nil {
		t.Fatalf("normalizeMap nil = %#v %v", m, ok)
	}
	if !isNilValue((*string)(nil)) || isNilValue("") {
		t.Fatal("isNilValue mismatch")
	}
	if !isStruct(sampleBean{}) || !isStruct(&sampleBean{}) || isStruct(nil) || isStruct(map[string]any{}) {
		t.Fatal("isStruct mismatch")
	}
	if got := parseScalar("false"); got != false {
		t.Fatalf("parseScalar false = %#v", got)
	}
	if got := parseScalar("plain"); got != "plain" {
		t.Fatalf("parseScalar string = %#v", got)
	}
	if got := fmt.Sprint(structToMap(struct {
		XMLName stdxml.Name `xml:"root"`
	}{}, true, false)); strings.Contains(got, "XMLName") {
		t.Fatalf("structToMap should skip XMLName: %s", got)
	}
}

func gotOK(got map[string]any, key string, want any) bool {
	return got != nil && got[key] == want
}
