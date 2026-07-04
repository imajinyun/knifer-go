// Package xml is the internal implementation of the vxml facade.
// External callers must depend on github.com/imajinyun/knifer-go/vxml.
package xml

import (
	"bytes"
	"encoding/json"
	stdxml "encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const (
	NBSP           = "&nbsp;"
	AMP            = "&amp;"
	QUOTE          = "&quot;"
	APOS           = "&apos;"
	LT             = "&lt;"
	GT             = "&gt;"
	InvalidRegex   = "[\\x00-\\x08\\x0b-\\x0c\\x0e-\\x1f]"
	CommentRegex   = "(?s)<!--.+?-->"
	IndentDefault  = 2
	ContentKey     = "content"
	DefaultCharset = "UTF-8"
)

var (
	invalidRe = regexp.MustCompile(InvalidRegex)
	commentRe = regexp.MustCompile(CommentRegex)
)

type cleanConfig struct {
	invalidRe *regexp.Regexp
	commentRe *regexp.Regexp
}

// CleanOption customizes XML cleaning helpers per call.
type CleanOption func(*cleanConfig)

// WithInvalidRegexp sets the regexp used by CleanInvalidWithOptions.
func WithInvalidRegexp(re *regexp.Regexp) CleanOption {
	return func(c *cleanConfig) {
		if re != nil {
			c.invalidRe = re
		}
	}
}

// WithCommentRegexp sets the regexp used by CleanCommentWithOptions.
func WithCommentRegexp(re *regexp.Regexp) CleanOption {
	return func(c *cleanConfig) {
		if re != nil {
			c.commentRe = re
		}
	}
}

func applyCleanOptions(opts []CleanOption) cleanConfig {
	cfg := cleanConfig{invalidRe: invalidRe, commentRe: commentRe}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.invalidRe == nil {
		cfg.invalidRe = invalidRe
	}
	if cfg.commentRe == nil {
		cfg.commentRe = commentRe
	}
	return cfg
}

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

// Document is a lightweight XML document tree.
type Document struct {
	Root *Element
}

// Element is a lightweight XML element node.
type Element struct {
	Name     stdxml.Name
	Attr     []stdxml.Attr
	Text     string
	Children []*Element
	Parent   *Element
}

// TokenHandler consumes streaming XML tokens.
type TokenHandler func(stdxml.Token) error

// NamespaceCache stores prefix-to-namespace URI mappings discovered from a document.
type NamespaceCache struct {
	Default string
	Prefix  map[string]string
	URI     map[string]string
}

// ---------------------------------------------------------------------------
// Parse options
// ---------------------------------------------------------------------------

// ParseOption customizes XML parsing per call.
type ParseOption func(*parseConfig)

type parseConfig struct {
	namespaceAware bool
	strict         bool
	charsetReader  func(charset string, input io.Reader) (io.Reader, error)
	entity         map[string]string
	maxBytes       int64
	openFile       func(string) (io.ReadCloser, error)
	decoderFactory func(io.Reader) *stdxml.Decoder
	parseInt       func(string, int, int) (int64, error)
	parseFloat     func(string, int) (float64, error)
}

func defaultParseConfig() parseConfig {
	return parseConfig{namespaceAware: true, strict: true}
}

func applyParse(opts []ParseOption) parseConfig {
	cfg := defaultParseConfig()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.openFile == nil {
		cfg.openFile = defaultOpenFile
	}
	if cfg.decoderFactory == nil {
		cfg.decoderFactory = stdxml.NewDecoder
	}
	if cfg.parseInt == nil {
		cfg.parseInt = strconv.ParseInt
	}
	if cfg.parseFloat == nil {
		cfg.parseFloat = strconv.ParseFloat
	}
	return cfg
}

func defaultOpenFile(path string) (io.ReadCloser, error) {
	// #nosec G304 -- XML file helpers intentionally read the caller-provided path.
	return os.Open(path)
}

// WithNamespaceAware controls whether parsed element names keep namespace URIs.
func WithNamespaceAware(b bool) ParseOption {
	return func(c *parseConfig) { c.namespaceAware = b }
}

// WithStrict controls XML decoder strict mode.
func WithStrict(b bool) ParseOption {
	return func(c *parseConfig) { c.strict = b }
}

// WithCharsetReader sets the charset reader used by the XML decoder.
func WithCharsetReader(reader func(charset string, input io.Reader) (io.Reader, error)) ParseOption {
	return func(c *parseConfig) {
		if reader != nil {
			c.charsetReader = reader
		}
	}
}

// WithEntity sets custom XML decoder entity replacements.
func WithEntity(entity map[string]string) ParseOption {
	return func(c *parseConfig) { c.entity = entity }
}

// WithMaxBytes bounds XML input read from readers and files. Non-positive values mean unlimited.
func WithMaxBytes(maxBytes int64) ParseOption {
	return func(c *parseConfig) { c.maxBytes = maxBytes }
}

// WithOpenFile sets the file opener used by XML file read helpers.
func WithOpenFile(openFile func(string) (io.ReadCloser, error)) ParseOption {
	return func(c *parseConfig) {
		if openFile != nil {
			c.openFile = openFile
		}
	}
}

// WithDecoderFactory sets the XML decoder factory used by DOM and SAX readers.
func WithDecoderFactory(factory func(io.Reader) *stdxml.Decoder) ParseOption {
	return func(c *parseConfig) {
		if factory != nil {
			c.decoderFactory = factory
		}
	}
}

// WithScalarIntParser sets the integer parser used by XML-to-map scalar conversion.
func WithScalarIntParser(parse func(string, int, int) (int64, error)) ParseOption {
	return func(c *parseConfig) {
		if parse != nil {
			c.parseInt = parse
		}
	}
}

// WithScalarFloatParser sets the float parser used by XML-to-map scalar conversion.
func WithScalarFloatParser(parse func(string, int) (float64, error)) ParseOption {
	return func(c *parseConfig) {
		if parse != nil {
			c.parseFloat = parse
		}
	}
}

// ---------------------------------------------------------------------------
// Write options
// ---------------------------------------------------------------------------

// WriteOption customizes XML serialization per call.
type WriteOption func(*writeConfig)

type writeConfig struct {
	charset            string
	indent             int // 0 means no indentation
	omitXMLDeclaration bool
	ignoreNullFields   bool
	rootName           string
	namespace          string
	filePerm           fs.FileMode
	dirPerm            fs.FileMode
	overwrite          bool
	createParents      bool
	mkdirAll           func(string, fs.FileMode) error
	openFile           func(string, int, fs.FileMode) (io.WriteCloser, error)
}

// BeanOption customizes XML map-to-bean conversion helpers per call.
type BeanOption func(*beanConfig)

type beanConfig struct {
	marshal   func(any) ([]byte, error)
	unmarshal func([]byte, any) error
}

func applyBean(opts []BeanOption) beanConfig {
	cfg := beanConfig{marshal: json.Marshal, unmarshal: json.Unmarshal}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.marshal == nil {
		cfg.marshal = json.Marshal
	}
	if cfg.unmarshal == nil {
		cfg.unmarshal = json.Unmarshal
	}
	return cfg
}

// WithBeanMarshalFunc sets the marshal provider used by XML bean conversion helpers.
func WithBeanMarshalFunc(marshal func(any) ([]byte, error)) BeanOption {
	return func(c *beanConfig) {
		if marshal != nil {
			c.marshal = marshal
		}
	}
}

// WithBeanUnmarshalFunc sets the unmarshal provider used by XML bean conversion helpers.
func WithBeanUnmarshalFunc(unmarshal func([]byte, any) error) BeanOption {
	return func(c *beanConfig) {
		if unmarshal != nil {
			c.unmarshal = unmarshal
		}
	}
}

// TransformOption customizes XML transform helpers per call.
type TransformOption func(*transformConfig)

type transformConfig struct {
	parse []ParseOption
	write []WriteOption
}

// WithTransformParseOptions sets parser options used by TransformWithOptions.
func WithTransformParseOptions(opts ...ParseOption) TransformOption {
	return func(c *transformConfig) { c.parse = append(c.parse, opts...) }
}

// WithTransformWriteOptions sets writer options used by TransformWithOptions.
func WithTransformWriteOptions(opts ...WriteOption) TransformOption {
	return func(c *transformConfig) { c.write = append(c.write, opts...) }
}

// FormatOption customizes XML formatting per call.
type FormatOption func(*formatConfig)

type formatConfig struct {
	parse []ParseOption
	write []WriteOption
}

// WithFormatParseOptions sets parser options used by FormatWithOptions.
func WithFormatParseOptions(opts ...ParseOption) FormatOption {
	return func(c *formatConfig) { c.parse = append(c.parse, opts...) }
}

// WithFormatWriteOptions sets writer options used by FormatWithOptions.
func WithFormatWriteOptions(opts ...WriteOption) FormatOption {
	return func(c *formatConfig) { c.write = append(c.write, opts...) }
}

func defaultWriteConfig() writeConfig {
	return writeConfig{charset: DefaultCharset, filePerm: 0o644, dirPerm: 0o750, overwrite: true, createParents: true}
}

func applyWrite(opts []WriteOption) writeConfig {
	cfg := defaultWriteConfig()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.charset == "" {
		cfg.charset = DefaultCharset
	}
	if cfg.mkdirAll == nil {
		cfg.mkdirAll = os.MkdirAll
	}
	if cfg.openFile == nil {
		cfg.openFile = defaultOpenWriteFile
	}
	return cfg
}

func defaultOpenWriteFile(path string, flag int, perm fs.FileMode) (io.WriteCloser, error) {
	// #nosec G304 -- XML file helpers intentionally write to the caller-provided destination path.
	return os.OpenFile(path, flag, perm)
}

// WithCharset sets the XML declaration charset.
func WithCharset(s string) WriteOption { return func(c *writeConfig) { c.charset = s } }

// WithIndent sets the indentation width in spaces (0 disables pretty printing).
func WithIndent(n int) WriteOption { return func(c *writeConfig) { c.indent = n } }

// WithPretty enables pretty printing with the default indentation.
func WithPretty() WriteOption { return func(c *writeConfig) { c.indent = IndentDefault } }

// WithOmitDeclaration controls whether the <?xml ... ?> prolog is emitted.
func WithOmitDeclaration(b bool) WriteOption {
	return func(c *writeConfig) { c.omitXMLDeclaration = b }
}

// WithIgnoreNullFields skips struct fields whose value is a typed nil.
func WithIgnoreNullFields(b bool) WriteOption { return func(c *writeConfig) { c.ignoreNullFields = b } }

// WithRootName overrides the root element name for MarshalMap / MarshalBean.
func WithRootName(s string) WriteOption { return func(c *writeConfig) { c.rootName = s } }

// WithNamespace sets the xmlns attribute on the synthesized root element.
func WithNamespace(s string) WriteOption { return func(c *writeConfig) { c.namespace = s } }

// WithFilePerm sets the file permission used by WriteFile.
func WithFilePerm(perm fs.FileMode) WriteOption { return func(c *writeConfig) { c.filePerm = perm } }

// WithDirPerm sets the parent-directory permission used by WriteFile.
func WithDirPerm(perm fs.FileMode) WriteOption { return func(c *writeConfig) { c.dirPerm = perm } }

// WithOverwrite controls whether WriteFile may replace an existing file.
func WithOverwrite(overwrite bool) WriteOption {
	return func(c *writeConfig) { c.overwrite = overwrite }
}

// WithCreateParents controls whether WriteFile creates parent directories.
func WithCreateParents(create bool) WriteOption {
	return func(c *writeConfig) { c.createParents = create }
}

// WithMkdirAll sets the directory creator used by WriteFile.
func WithMkdirAll(mkdirAll func(string, fs.FileMode) error) WriteOption {
	return func(c *writeConfig) {
		if mkdirAll != nil {
			c.mkdirAll = mkdirAll
		}
	}
}

// WithOpenWriteFile sets the file opener used by WriteFile.
func WithOpenWriteFile(openFile func(string, int, fs.FileMode) (io.WriteCloser, error)) WriteOption {
	return func(c *writeConfig) {
		if openFile != nil {
			c.openFile = openFile
		}
	}
}

// ---------------------------------------------------------------------------
// Reading
// ---------------------------------------------------------------------------

// ReadXML parses XML content directly, or treats the input as a file path when
// it does not start with '<'.
func ReadXML(pathOrContent string, opts ...ParseOption) (*Document, error) {
	if strings.HasPrefix(strings.TrimSpace(pathOrContent), "<") {
		return ParseXML(pathOrContent, opts...)
	}
	return ReadXMLFile(pathOrContent, opts...)
}

// ReadXMLFile parses an XML file.
func ReadXMLFile(path string, opts ...ParseOption) (*Document, error) {
	cfg := applyParse(opts)
	f, err := cfg.openFile(path) // #nosec G304 -- SDK file helper intentionally reads the caller-provided XML path.
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	return readXMLReader(f, cfg)
}

// ReadXMLBytes parses XML bytes.
func ReadXMLBytes(data []byte, opts ...ParseOption) (*Document, error) {
	return ReadXMLReader(bytes.NewReader(data), opts...)
}

// ReadXMLReader parses XML from reader.
func ReadXMLReader(r io.Reader, opts ...ParseOption) (*Document, error) {
	return readXMLReader(r, applyParse(opts))
}

// ParseXML parses an XML string.
func ParseXML(xmlStr string, opts ...ParseOption) (*Document, error) {
	return ReadXMLReader(strings.NewReader(xmlStr), opts...)
}

func readXMLReader(r io.Reader, cfg parseConfig) (*Document, error) {
	if cfg.maxBytes > 0 {
		r = io.LimitReader(r, cfg.maxBytes)
	}
	dec := cfg.decoderFactory(r)
	if dec == nil {
		return nil, invalidInputf("vxml: decoder factory returned nil")
	}
	dec.Strict = cfg.strict
	dec.CharsetReader = cfg.charsetReader
	dec.Entity = cfg.entity
	var (
		stack []*Element
		root  *Element
	)
	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, wrapInvalidInput("vxml: decode token", err)
		}
		switch t := tok.(type) {
		case stdxml.StartElement:
			name := t.Name
			if !cfg.namespaceAware {
				name.Space = ""
			}
			ele := &Element{Name: name, Attr: slices.Clone(t.Attr)}
			if len(stack) > 0 {
				parent := stack[len(stack)-1]
				ele.Parent = parent
				parent.Children = append(parent.Children, ele)
			} else if root == nil {
				root = ele
			}
			stack = append(stack, ele)
		case stdxml.EndElement:
			if len(stack) == 0 {
				return nil, invalidInputf("vxml: unexpected closing tag: %s", t.Name.Local)
			}
			stack = stack[:len(stack)-1]
		case stdxml.CharData:
			if len(stack) > 0 {
				stack[len(stack)-1].Text += string(t)
			}
		}
	}
	if root == nil {
		return nil, invalidInputf("vxml: root element not found")
	}
	return &Document{Root: root}, nil
}

// ReadBySAX streams XML tokens from reader to handler.
func ReadBySAX(r io.Reader, handler TokenHandler) error {
	return ReadBySAXWithOptions(r, handler)
}

// ReadBySAXWithOptions streams XML tokens from reader to handler with custom parse options.
func ReadBySAXWithOptions(r io.Reader, handler TokenHandler, opts ...ParseOption) error {
	return readBySAXWithConfig(r, handler, applyParse(opts))
}

func stripTokenNamespace(tok stdxml.Token) stdxml.Token {
	switch t := tok.(type) {
	case stdxml.StartElement:
		t.Name.Space = ""
		for i := range t.Attr {
			t.Attr[i].Name.Space = ""
		}
		return t
	case stdxml.EndElement:
		t.Name.Space = ""
		return t
	default:
		return tok
	}
}

// ReadBySAXFile streams XML tokens from file.
func ReadBySAXFile(path string, handler TokenHandler) (err error) {
	return ReadBySAXFileWithOptions(path, handler)
}

// ReadBySAXFileWithOptions streams XML tokens from file with custom parse options.
func ReadBySAXFileWithOptions(path string, handler TokenHandler, opts ...ParseOption) (err error) {
	cfg := applyParse(opts)
	// #nosec G304 -- SDK file helper intentionally opens the caller-provided XML path.
	f, err := cfg.openFile(path)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := f.Close(); err == nil {
			err = closeErr
		}
	}()
	return readBySAXWithConfig(f, handler, cfg)
}

func readBySAXWithConfig(r io.Reader, handler TokenHandler, cfg parseConfig) error {
	if handler == nil {
		return nil
	}
	if cfg.maxBytes > 0 {
		r = io.LimitReader(r, cfg.maxBytes)
	}
	dec := cfg.decoderFactory(r)
	if dec == nil {
		return invalidInputf("vxml: decoder factory returned nil")
	}
	dec.Strict = cfg.strict
	dec.CharsetReader = cfg.charsetReader
	dec.Entity = cfg.entity
	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return wrapInvalidInput("vxml: sax decode", err)
		}
		if !cfg.namespaceAware {
			tok = stripTokenNamespace(tok)
		}
		if err := handler(tok); err != nil {
			return err
		}
	}
}

// ---------------------------------------------------------------------------
// Writing
// ---------------------------------------------------------------------------

// WriteTo serializes a document or element to writer.
func WriteTo(w io.Writer, v any, opts ...WriteOption) error {
	if w == nil {
		return invalidInputf("vxml: nil writer")
	}
	return writeWithConfig(w, v, applyWrite(opts))
}

// MarshalString serializes a document or element to string.
func MarshalString(v any, opts ...WriteOption) (string, error) {
	var buf bytes.Buffer
	if err := WriteTo(&buf, v, opts...); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// WriteFile writes a document or element to path.
func WriteFile(path string, v any, opts ...WriteOption) (err error) {
	cfg := applyWrite(opts)
	if cfg.createParents {
		if err := cfg.mkdirAll(filepath.Dir(path), cfg.dirPerm); err != nil {
			return err
		}
	}
	flag := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if !cfg.overwrite {
		flag |= os.O_EXCL
	}
	// #nosec G304 -- SDK file helper intentionally creates the caller-provided XML path.
	f, err := cfg.openFile(path, flag, cfg.filePerm)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := f.Close(); err == nil {
			err = closeErr
		}
	}()
	return writeWithConfig(f, v, cfg)
}

// MarshalMap serializes map data to an XML string.
func MarshalMap(data map[string]any, opts ...WriteOption) (string, error) {
	cfg := applyWrite(opts)
	root := cfg.rootName
	if root == "" {
		root = "xml"
	}
	doc := CreateXMLWithRootNS(root, cfg.namespace)
	Append(doc.Root, data)
	return MarshalString(doc, opts...)
}

// MarshalBean serializes a struct or map-like value to an XML string.
func MarshalBean(bean any, opts ...WriteOption) (string, error) {
	cfg := applyWrite(opts)
	name := cfg.rootName
	if name == "" {
		name = typeName(bean)
	}
	if name == "" {
		name = "xml"
	}
	m := structToMap(bean, false, cfg.ignoreNullFields)
	doc := CreateXMLWithRootNS(name, cfg.namespace)
	Append(doc.Root, m)
	return MarshalString(doc, opts...)
}

// TransformWith copies XML from source to result with per-call options.
func TransformWith(source io.Reader, result io.Writer, opts ...WriteOption) error {
	return TransformWithOptions(source, result, WithTransformWriteOptions(opts...))
}

// TransformWithOptions copies XML from source to result with parser and writer options.
func TransformWithOptions(source io.Reader, result io.Writer, opts ...TransformOption) error {
	cfg := transformConfig{}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	doc, err := ReadXMLReader(source, cfg.parse...)
	if err != nil {
		return err
	}
	return WriteTo(result, doc, cfg.write...)
}

// Format pretty prints XML content.
func Format(xmlStr string) (string, error) {
	return FormatWithOptions(xmlStr)
}

// FormatWithOptions pretty prints XML content with parser and writer options.
func FormatWithOptions(xmlStr string, opts ...FormatOption) (string, error) {
	cfg := formatConfig{write: []WriteOption{WithPretty()}}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	doc, err := ParseXML(xmlStr, cfg.parse...)
	if err != nil {
		return "", err
	}
	return MarshalString(doc, cfg.write...)
}

// ---------------------------------------------------------------------------
// Element construction & traversal
// ---------------------------------------------------------------------------

// CreateXML creates an empty XML document.
func CreateXML() *Document { return &Document{} }

// CreateXMLWithRoot creates an XML document with root element.
func CreateXMLWithRoot(rootElementName string) *Document {
	return CreateXMLWithRootNS(rootElementName, "")
}

// CreateXMLWithRootNS creates an XML document with root element and namespace URI.
func CreateXMLWithRootNS(rootElementName, namespace string) *Document {
	root := &Element{Name: stdxml.Name{Local: rootElementName}}
	if namespace != "" {
		root.Attr = append(root.Attr, stdxml.Attr{Name: stdxml.Name{Local: "xmlns"}, Value: namespace})
	}
	return &Document{Root: root}
}

// GetRootElement returns the document root element.
func GetRootElement(doc *Document) *Element {
	if doc == nil {
		return nil
	}
	return doc.Root
}

// GetOwnerDocument returns the document that owns node.
func GetOwnerDocument(node *Element) *Document {
	if node == nil {
		return nil
	}
	for node.Parent != nil {
		node = node.Parent
	}
	return &Document{Root: node}
}

// CleanInvalid removes XML 1.0 invalid control characters.
func CleanInvalid(xmlContent string) string { return CleanInvalidWithOptions(xmlContent) }

// CleanInvalidWithOptions removes XML 1.0 invalid control characters with options.
func CleanInvalidWithOptions(xmlContent string, opts ...CleanOption) string {
	return applyCleanOptions(opts).invalidRe.ReplaceAllString(xmlContent, "")
}

// CleanComment removes XML comments.
func CleanComment(xmlContent string) string { return CleanCommentWithOptions(xmlContent) }

// CleanCommentWithOptions removes XML comments with options.
func CleanCommentWithOptions(xmlContent string, opts ...CleanOption) string {
	return applyCleanOptions(opts).commentRe.ReplaceAllString(xmlContent, "")
}

// GetElements returns child elements with tag name. Empty tagName returns all direct children.
func GetElements(element *Element, tagName string) []*Element {
	if element == nil {
		return nil
	}
	out := make([]*Element, 0, len(element.Children))
	for _, child := range element.Children {
		if tagName == "" || child.Name.Local == tagName {
			out = append(out, child)
		}
	}
	return out
}

// GetElement returns the first child element with tag name.
func GetElement(element *Element, tagName string) *Element {
	for _, child := range GetElements(element, tagName) {
		return child
	}
	return nil
}

// ElementText returns child text or defaultValue when missing.
func ElementText(element *Element, tagName string, defaultValue ...string) string {
	if child := GetElement(element, tagName); child != nil {
		return strings.TrimSpace(child.Text)
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return ""
}

// TransElements returns the input list without nil elements.
func TransElements(nodes []*Element) []*Element {
	out := make([]*Element, 0, len(nodes))
	for _, node := range nodes {
		if node != nil {
			out = append(out, node)
		}
	}
	return out
}

// IsElement reports whether node is not nil.
func IsElement(node *Element) bool { return node != nil }

// AppendChild appends and returns a child element.
func AppendChild(node *Element, tagName string, namespace ...string) *Element {
	if node == nil {
		return nil
	}
	ns := ""
	if len(namespace) > 0 {
		ns = namespace[0]
	}
	child := &Element{Name: stdxml.Name{Local: tagName}, Parent: node}
	if ns != "" {
		child.Attr = append(child.Attr, stdxml.Attr{Name: stdxml.Name{Local: "xmlns"}, Value: ns})
	}
	node.Children = append(node.Children, child)
	return child
}

// AppendText appends text to an element.
func AppendText(node *Element, text any) *Element {
	if node != nil && text != nil {
		node.Text += fmt.Sprint(text)
	}
	return node
}

// Append appends map, slice, struct, or scalar data to node.
func Append(node *Element, data any) {
	if node == nil || data == nil {
		return
	}
	appendValue(node, data)
}

// ---------------------------------------------------------------------------
// XPath (simple expression subset)
// ---------------------------------------------------------------------------

// GetElementByXPath returns the first element matched by a simple expression.
func GetElementByXPath(expression string, source any) *Element {
	return GetNodeByXPath(expression, source)
}

// GetNodeListByXPath returns all elements matched by a simple expression.
func GetNodeListByXPath(expression string, source any) []*Element {
	root := elementOf(source)
	if root == nil {
		return nil
	}
	return findByPath(root, expression)
}

// GetNodeByXPath returns the first node matched by a simple expression.
func GetNodeByXPath(expression string, source any) *Element {
	nodes := GetNodeListByXPath(expression, source)
	if len(nodes) == 0 {
		return nil
	}
	return nodes[0]
}

// GetByXPath returns matched text, element, or list based on returnType:
// "string"/"text", "node" (default), or "nodes"/"nodelist"/"list".
func GetByXPath(expression string, source any, returnType string) any {
	switch strings.ToLower(returnType) {
	case "string", "text":
		if node := GetNodeByXPath(expression, source); node != nil {
			return strings.TrimSpace(node.Text)
		}
		return ""
	case "nodes", "nodelist", "list":
		return GetNodeListByXPath(expression, source)
	default:
		return GetNodeByXPath(expression, source)
	}
}

// ---------------------------------------------------------------------------
// Escape / unescape
// ---------------------------------------------------------------------------

// Escape escapes XML text.
func Escape(s string) string {
	var buf bytes.Buffer
	if err := stdxml.EscapeText(&buf, []byte(s)); err != nil {
		return s
	}
	return buf.String()
}

// Unescape unescapes XML/HTML entities.
func Unescape(s string) string { return html.UnescapeString(s) }

// ---------------------------------------------------------------------------
// XML <-> Map / Bean
// ---------------------------------------------------------------------------

// XMLToMap parses XML into a nested map. Repeated sibling tags become []any.
func XMLToMap(xmlStr string) (map[string]any, error) {
	return XMLToMapWithOptions(xmlStr)
}

// XMLToMapWithOptions parses XML into a nested map with parser options.
func XMLToMapWithOptions(xmlStr string, opts ...ParseOption) (map[string]any, error) {
	cfg := applyParse(opts)
	doc, err := ParseXML(xmlStr, opts...)
	if err != nil {
		return nil, err
	}
	result := map[string]any{}
	if doc.Root != nil {
		addMapValue(result, doc.Root.Name.Local, elementToValue(doc.Root, cfg))
	}
	return result, nil
}

// XMLNodeToMap converts an element into a nested map value.
func XMLNodeToMap(node *Element) map[string]any {
	return XMLNodeToMapWithOptions(node)
}

// XMLNodeToMapWithOptions converts an element into a nested map value using custom scalar parsers.
func XMLNodeToMapWithOptions(node *Element, opts ...ParseOption) map[string]any {
	cfg := applyParse(opts)
	result := map[string]any{}
	if node != nil {
		addMapValue(result, node.Name.Local, elementToValue(node, cfg))
	}
	return result
}

// XMLToBean parses XML and decodes the generated map into dst.
func XMLToBean(xmlStr string, dst any) error {
	return XMLToBeanWithOptions(xmlStr, dst)
}

// XMLToBeanWithOptions parses XML and decodes the generated map into dst with parser options.
func XMLToBeanWithOptions(xmlStr string, dst any, opts ...ParseOption) error {
	m, err := XMLToMapWithOptions(xmlStr, opts...)
	if err != nil {
		return err
	}
	return mapToBeanWithOptions(m, dst)
}

// XMLNodeToBean converts an element tree to a map and decodes it into dst.
func XMLNodeToBean(node *Element, dst any) error { return XMLNodeToBeanWithOptions(node, dst) }

// XMLNodeToBeanWithOptions converts an element tree to a map and decodes it into dst with bean options.
func XMLNodeToBeanWithOptions(node *Element, dst any, opts ...BeanOption) error {
	return mapToBeanWithOptions(XMLNodeToMap(node), dst, opts...)
}

// XMLNodeToBeanWithParseOptions converts an element tree to a map and decodes it into dst with parser and bean options.
func XMLNodeToBeanWithParseOptions(node *Element, dst any, parseOpts []ParseOption, beanOpts ...BeanOption) error {
	return mapToBeanWithOptions(XMLNodeToMapWithOptions(node, parseOpts...), dst, beanOpts...)
}

// XMLToMapInto parses XML and merges values into result.
func XMLToMapInto(xmlStr string, result map[string]any) (map[string]any, error) {
	return XMLToMapIntoWithOptions(xmlStr, result)
}

// XMLToMapIntoWithOptions parses XML and merges values into result with parser options.
func XMLToMapIntoWithOptions(xmlStr string, result map[string]any, opts ...ParseOption) (map[string]any, error) {
	m, err := XMLToMapWithOptions(xmlStr, opts...)
	if err != nil {
		return result, err
	}
	if result == nil {
		result = map[string]any{}
	}
	for k, v := range m {
		result[k] = v
	}
	return result, nil
}

// XMLNodeToMapInto converts an element to map and merges values into result.
func XMLNodeToMapInto(node *Element, result map[string]any) map[string]any {
	return XMLNodeToMapIntoWithOptions(node, result)
}

// XMLNodeToMapIntoWithOptions converts an element to map and merges values into result with parser options.
func XMLNodeToMapIntoWithOptions(node *Element, result map[string]any, opts ...ParseOption) map[string]any {
	if result == nil {
		result = map[string]any{}
	}
	for k, v := range XMLNodeToMapWithOptions(node, opts...) {
		result[k] = v
	}
	return result
}

func mapToBeanWithOptions(m map[string]any, dst any, opts ...BeanOption) error {
	cfg := applyBean(opts)
	data, err := cfg.marshal(m)
	if err != nil {
		return wrapInternal("vxml: encode intermediate", err)
	}
	if err := cfg.unmarshal(data, dst); err != nil {
		return wrapInvalidInput("vxml: decode into dst", err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Namespace cache
// ---------------------------------------------------------------------------

// NewNamespaceCache collects namespace declarations from doc.
func NewNamespaceCache(doc *Document) *NamespaceCache {
	cache := &NamespaceCache{Prefix: map[string]string{}, URI: map[string]string{}}
	walk(GetRootElement(doc), func(ele *Element) {
		for _, attr := range ele.Attr {
			if attr.Name.Local == "xmlns" && attr.Name.Space == "" {
				cache.Default = attr.Value
				cache.Prefix["DEFAULT"] = attr.Value
				cache.URI[attr.Value] = "DEFAULT"
				continue
			}
			if attr.Name.Space == "xmlns" {
				cache.Prefix[attr.Name.Local] = attr.Value
				cache.URI[attr.Value] = attr.Name.Local
			}
		}
	})
	return cache
}

// NamespaceURI returns namespace URI for prefix.
func (c *NamespaceCache) NamespaceURI(prefix string) string {
	if c == nil {
		return ""
	}
	if prefix == "" {
		return c.Default
	}
	return c.Prefix[prefix]
}

// PrefixOf returns one prefix for namespace URI.
func (c *NamespaceCache) PrefixOf(uri string) string {
	if c == nil {
		return ""
	}
	return c.URI[uri]
}

// ---------------------------------------------------------------------------
// Internals
// ---------------------------------------------------------------------------

func writeWithConfig(w io.Writer, v any, cfg writeConfig) error {
	if !cfg.omitXMLDeclaration {
		if _, err := fmt.Fprintf(w, `<?xml version="1.0" encoding="%s"?>`, cfg.charset); err != nil {
			return err
		}
		if cfg.indent > 0 {
			if _, err := io.WriteString(w, "\n"); err != nil {
				return err
			}
		}
	}
	ele := elementOf(v)
	if ele == nil {
		return invalidInputf("vxml: unsupported node")
	}
	return writeElement(w, ele, cfg.indent, 0)
}

func elementOf(v any) *Element {
	switch x := v.(type) {
	case *Document:
		if x == nil {
			return nil
		}
		return x.Root
	case *Element:
		return x
	default:
		return nil
	}
}

func writeElement(w io.Writer, ele *Element, indentSize, level int) error {
	if ele == nil {
		return nil
	}
	if indentSize > 0 {
		if _, err := io.WriteString(w, strings.Repeat(" ", indentSize*level)); err != nil {
			return err
		}
	}
	if _, err := io.WriteString(w, "<"); err != nil {
		return err
	}
	if err := writeName(w, ele.Name); err != nil {
		return err
	}
	for _, attr := range ele.Attr {
		if _, err := io.WriteString(w, " "); err != nil {
			return err
		}
		if err := writeName(w, attr.Name); err != nil {
			return err
		}
		if _, err := io.WriteString(w, `="`); err != nil {
			return err
		}
		if _, err := io.WriteString(w, Escape(attr.Value)); err != nil {
			return err
		}
		if _, err := io.WriteString(w, `"`); err != nil {
			return err
		}
	}
	if len(ele.Children) == 0 && strings.TrimSpace(ele.Text) == "" {
		_, err := io.WriteString(w, "/>")
		return err
	}
	if _, err := io.WriteString(w, ">"); err != nil {
		return err
	}
	if text := strings.TrimSpace(ele.Text); text != "" {
		if _, err := io.WriteString(w, Escape(text)); err != nil {
			return err
		}
	}
	if len(ele.Children) > 0 {
		if indentSize > 0 {
			if _, err := io.WriteString(w, "\n"); err != nil {
				return err
			}
		}
		for i, child := range ele.Children {
			if err := writeElement(w, child, indentSize, level+1); err != nil {
				return err
			}
			if indentSize > 0 && i < len(ele.Children)-1 {
				if _, err := io.WriteString(w, "\n"); err != nil {
					return err
				}
			}
		}
		if indentSize > 0 {
			if _, err := io.WriteString(w, "\n"); err != nil {
				return err
			}
			if _, err := io.WriteString(w, strings.Repeat(" ", indentSize*level)); err != nil {
				return err
			}
		}
	}
	if _, err := io.WriteString(w, "</"); err != nil {
		return err
	}
	if err := writeName(w, ele.Name); err != nil {
		return err
	}
	_, err := io.WriteString(w, ">")
	return err
}

func writeName(w io.Writer, name stdxml.Name) error {
	_, err := io.WriteString(w, name.Local)
	return err
}

func elementToValue(ele *Element, cfg parseConfig) any {
	obj := map[string]any{}
	for _, attr := range ele.Attr {
		obj[attr.Name.Local] = parseScalar(attr.Value, cfg)
	}
	for _, child := range ele.Children {
		addMapValue(obj, child.Name.Local, elementToValue(child, cfg))
	}
	text := strings.TrimSpace(ele.Text)
	if len(obj) == 0 {
		if text == "" {
			return ""
		}
		return parseScalar(text, cfg)
	}
	if text != "" {
		obj[ContentKey] = parseScalar(text, cfg)
	}
	return obj
}

func addMapValue(m map[string]any, key string, value any) {
	if old, ok := m[key]; ok {
		if arr, ok := old.([]any); ok {
			m[key] = append(arr, value)
		} else {
			m[key] = []any{old, value}
		}
		return
	}
	m[key] = value
}

func parseScalar(s string, cfg parseConfig) any {
	s = strings.TrimSpace(s)
	switch strings.ToLower(s) {
	case "true":
		return true
	case "false":
		return false
	case "null":
		return nil
	}
	if cfg.parseInt == nil {
		cfg.parseInt = strconv.ParseInt
	}
	if cfg.parseFloat == nil {
		cfg.parseFloat = strconv.ParseFloat
	}
	if i, err := cfg.parseInt(s, 10, 64); err == nil {
		return i
	}
	if f, err := cfg.parseFloat(s, 64); err == nil {
		return f
	}
	return s
}

func appendValue(node *Element, data any) {
	if m, ok := normalizeMap(data); ok {
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		slices.Sort(keys)
		for _, key := range keys {
			appendNamedValue(node, key, m[key])
		}
		return
	}
	rv := reflect.ValueOf(data)
	if rv.IsValid() && (rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array) {
		for i := 0; i < rv.Len(); i++ {
			appendNamedValue(node, "element", rv.Index(i).Interface())
		}
		return
	}
	AppendText(node, data)
}

func appendNamedValue(parent *Element, key string, value any) {
	rv := reflect.ValueOf(value)
	if rv.IsValid() && (rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array) {
		for i := 0; i < rv.Len(); i++ {
			appendNamedValue(parent, key, rv.Index(i).Interface())
		}
		return
	}
	child := AppendChild(parent, key)
	if value == nil {
		return
	}
	if m, ok := normalizeMap(value); ok {
		appendValue(child, m)
		return
	}
	if isStruct(value) {
		appendValue(child, structToMap(value, true, false))
		return
	}
	AppendText(child, value)
}

func normalizeMap(data any) (map[string]any, bool) {
	if data == nil {
		return nil, false
	}
	switch m := data.(type) {
	case map[string]any:
		return m, true
	case map[string]string:
		out := make(map[string]any, len(m))
		for k, v := range m {
			out[k] = v
		}
		return out, true
	case map[any]any:
		out := make(map[string]any, len(m))
		for k, v := range m {
			out[fmt.Sprint(k)] = v
		}
		return out, true
	}
	rv := reflect.ValueOf(data)
	if rv.Kind() == reflect.Map {
		out := map[string]any{}
		iter := rv.MapRange()
		for iter.Next() {
			out[fmt.Sprint(iter.Key().Interface())] = iter.Value().Interface()
		}
		return out, true
	}
	return nil, false
}

func isStruct(data any) bool {
	if data == nil {
		return false
	}
	rv := reflect.ValueOf(data)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	return rv.IsValid() && rv.Kind() == reflect.Struct
}

func structToMap(data any, honorXMLName, ignoreNull bool) map[string]any {
	out := map[string]any{}
	rv := reflect.ValueOf(data)
	if !rv.IsValid() {
		return out
	}
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return out
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return out
	}
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		if field.PkgPath != "" {
			continue
		}
		name := field.Name
		if tag := field.Tag.Get("xml"); tag != "" {
			name = strings.Split(tag, ",")[0]
			if name == "-" {
				continue
			}
		}
		if honorXMLName && name == "XMLName" {
			continue
		}
		if name == "" || name == "XMLName" {
			continue
		}
		value := rv.Field(i).Interface()
		if ignoreNull && isNilValue(value) {
			continue
		}
		out[name] = value
	}
	return out
}

func isNilValue(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}

func typeName(data any) string {
	if data == nil {
		return ""
	}
	t := reflect.TypeOf(data)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() == reflect.Struct {
		return strings.ToLower(t.Name())
	}
	return ""
}

func findByPath(root *Element, expr string) []*Element {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil
	}
	if strings.HasPrefix(expr, "//") {
		name := strings.TrimPrefix(expr, "//")
		var out []*Element
		walk(root, func(ele *Element) {
			if ele.Name.Local == name {
				out = append(out, ele)
			}
		})
		return out
	}
	parts := strings.Split(strings.Trim(expr, "/"), "/")
	if len(parts) == 0 {
		return nil
	}
	current := []*Element{root}
	if parts[0] == root.Name.Local {
		parts = parts[1:]
	}
	for _, part := range parts {
		if part == "" {
			continue
		}
		next := make([]*Element, 0)
		for _, ele := range current {
			for _, child := range ele.Children {
				if child.Name.Local == part {
					next = append(next, child)
				}
			}
		}
		current = next
	}
	return current
}

func walk(ele *Element, fn func(*Element)) {
	if ele == nil {
		return
	}
	fn(ele)
	for _, child := range ele.Children {
		walk(child, fn)
	}
}
