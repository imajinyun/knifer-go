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

func TestFacadeSAXReadOptions(t *testing.T) {
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
}

func TestFacadeXMLReadFileProviderOptions(t *testing.T) {
	openedPath := ""
	doc, err := ReadXMLFile("virtual.xml", WithOpenFile(func(path string) (io.ReadCloser, error) {
		openedPath = path
		return io.NopCloser(strings.NewReader(`<root><from>facade</from></root>`)), nil
	}))
	if err != nil || ElementText(doc.Root, "from") != "facade" || openedPath != "virtual.xml" {
		t.Fatalf("ReadXMLFile provider doc=%#v path=%q err=%v", doc, openedPath, err)
	}

	var starts []string
	openedPath = ""
	if err := ReadBySAXFileWithOptions("sax.xml", func(tok stdxml.Token) error {
		if start, ok := tok.(stdxml.StartElement); ok {
			starts = append(starts, start.Name.Local)
		}
		return nil
	}, WithOpenFile(func(path string) (io.ReadCloser, error) {
		openedPath = path
		return io.NopCloser(strings.NewReader(`<root><a/></root>`)), nil
	})); err != nil || !reflect.DeepEqual(starts, []string{"root", "a"}) || openedPath != "sax.xml" {
		t.Fatalf("ReadBySAXFileWithOptions provider starts=%v path=%q err=%v", starts, openedPath, err)
	}
}
