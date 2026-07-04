package xml

import (
	stdxml "encoding/xml"
	"io"
	"io/fs"
	"strings"
	"testing"
)

func TestWithCharsetReader(t *testing.T) {
	cfg := parseConfig{}
	fn := func(charset string, input io.Reader) (io.Reader, error) { return input, nil }
	WithCharsetReader(fn)(&cfg)
	if cfg.charsetReader == nil {
		t.Fatal("WithCharsetReader did not set charsetReader")
	}
}

func TestWithEntity(t *testing.T) {
	cfg := parseConfig{}
	entity := map[string]string{"foo": "bar"}
	WithEntity(entity)(&cfg)
	if cfg.entity == nil || cfg.entity["foo"] != "bar" {
		t.Fatal("WithEntity did not set entity")
	}
}

func TestWithBeanMarshalFunc(t *testing.T) {
	cfg := beanConfig{}
	fn := func(v any) ([]byte, error) { return nil, nil }
	WithBeanMarshalFunc(fn)(&cfg)
	if cfg.marshal == nil {
		t.Fatal("WithBeanMarshalFunc did not set marshal")
	}
}

func TestNilProviderOptionsDoNotOverwriteConfiguredProviders(t *testing.T) {
	parseCfg := parseConfig{}
	charsetReader := func(_ string, input io.Reader) (io.Reader, error) { return input, nil }
	openFile := func(string) (io.ReadCloser, error) { return io.NopCloser(strings.NewReader("")), nil }
	decoderFactory := stdxml.NewDecoder
	parseInt := func(string, int, int) (int64, error) { return 1, nil }
	parseFloat := func(string, int) (float64, error) { return 1, nil }
	WithCharsetReader(charsetReader)(&parseCfg)
	WithCharsetReader(nil)(&parseCfg)
	WithOpenFile(openFile)(&parseCfg)
	WithOpenFile(nil)(&parseCfg)
	WithDecoderFactory(decoderFactory)(&parseCfg)
	WithDecoderFactory(nil)(&parseCfg)
	WithScalarIntParser(parseInt)(&parseCfg)
	WithScalarIntParser(nil)(&parseCfg)
	WithScalarFloatParser(parseFloat)(&parseCfg)
	WithScalarFloatParser(nil)(&parseCfg)
	if parseCfg.charsetReader == nil || parseCfg.openFile == nil || parseCfg.decoderFactory == nil || parseCfg.parseInt == nil || parseCfg.parseFloat == nil {
		t.Fatalf("nil parse provider option overwrote configured provider: %#v", parseCfg)
	}

	beanCfg := beanConfig{}
	marshal := func(any) ([]byte, error) { return []byte("{}"), nil }
	unmarshal := func([]byte, any) error { return nil }
	WithBeanMarshalFunc(marshal)(&beanCfg)
	WithBeanMarshalFunc(nil)(&beanCfg)
	WithBeanUnmarshalFunc(unmarshal)(&beanCfg)
	WithBeanUnmarshalFunc(nil)(&beanCfg)
	if beanCfg.marshal == nil || beanCfg.unmarshal == nil {
		t.Fatalf("nil bean provider option overwrote configured provider: %#v", beanCfg)
	}

	writeCfg := writeConfig{}
	mkdirAll := func(string, fs.FileMode) error { return nil }
	openWriteFile := func(string, int, fs.FileMode) (io.WriteCloser, error) {
		return nopWriteCloser{}, nil
	}
	WithMkdirAll(mkdirAll)(&writeCfg)
	WithMkdirAll(nil)(&writeCfg)
	WithOpenWriteFile(openWriteFile)(&writeCfg)
	WithOpenWriteFile(nil)(&writeCfg)
	if writeCfg.mkdirAll == nil || writeCfg.openFile == nil {
		t.Fatalf("nil write provider option overwrote configured provider: %#v", writeCfg)
	}
}

func TestWithNamespace(t *testing.T) {
	cfg := writeConfig{}
	WithNamespace("urn:test")(&cfg)
	if cfg.namespace != "urn:test" {
		t.Fatalf("WithNamespace = %q, want urn:test", cfg.namespace)
	}
}

func TestWithCreateParentsXML(t *testing.T) {
	cfg := writeConfig{}
	WithCreateParents(false)(&cfg)
	if cfg.createParents {
		t.Fatal("WithCreateParents(false) did not set createParents")
	}
}

func TestXMLNodeToBeanWithParseOptions(t *testing.T) {
	doc, err := ParseXML(`<root><name>test</name></root>`)
	if err != nil {
		t.Fatal(err)
	}
	var dst any
	err = XMLNodeToBeanWithParseOptions(doc.Root, &dst, nil)
	if err != nil {
		t.Fatalf("XMLNodeToBeanWithParseOptions error = %v", err)
	}
}

func TestWithCreateParentsOnWriteConfig(t *testing.T) {
	cfg := writeConfig{}
	WithCreateParents(true)(&cfg)
	if !cfg.createParents {
		t.Fatal("WithCreateParents(true) did not set createParents")
	}
}

func TestWithNamespaceOnWrite(t *testing.T) {
	_, err := MarshalString(CreateXMLWithRoot("root"), WithNamespace("urn:test"))
	if err != nil {
		t.Fatalf("MarshalString with namespace error = %v", err)
	}
}

func TestWithCharsetReaderIntegration(t *testing.T) {
	reader := func(charset string, input io.Reader) (io.Reader, error) {
		return input, nil
	}
	doc, err := ParseXML(`<root/>`, WithCharsetReader(reader))
	if err != nil {
		t.Fatalf("ParseXML with charset reader error = %v", err)
	}
	if doc == nil || doc.Root == nil {
		t.Fatal("ParseXML with charset reader returned nil doc")
	}
}

func TestWithFilePermAndDirPerm(t *testing.T) {
	cfg := writeConfig{}
	WithFilePerm(fs.FileMode(0o644))(&cfg)
	WithDirPerm(fs.FileMode(0o755))(&cfg)
	if cfg.filePerm != 0o644 || cfg.dirPerm != 0o755 {
		t.Fatalf("perms: file=%o dir=%o", cfg.filePerm, cfg.dirPerm)
	}
}

func TestMarshalStringWithWriteOptions(t *testing.T) {
	xml, err := MarshalString(CreateXMLWithRoot("root"), WithPretty(), WithOmitDeclaration(true))
	if err != nil {
		t.Fatalf("MarshalString error = %v", err)
	}
	if !strings.Contains(xml, "<root/>") {
		t.Fatalf("xml does not contain root: %s", xml)
	}
}
