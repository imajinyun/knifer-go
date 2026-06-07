package vxml

import (
	stdxml "encoding/xml"
	"io"
	"io/fs"

	xmlimpl "github.com/imajinyun/go-knifer/internal/xml"
)

// WithNamespaceAware controls whether parsed element names keep namespace URIs.
func WithNamespaceAware(b bool) ParseOption { return xmlimpl.WithNamespaceAware(b) }

// WithStrict controls XML decoder strict mode.
func WithStrict(b bool) ParseOption { return xmlimpl.WithStrict(b) }

// WithCharsetReader sets the charset reader used by the XML decoder.
func WithCharsetReader(reader func(charset string, input io.Reader) (io.Reader, error)) ParseOption {
	return xmlimpl.WithCharsetReader(reader)
}

// WithEntity sets custom XML decoder entity replacements.
func WithEntity(entity map[string]string) ParseOption { return xmlimpl.WithEntity(entity) }

// WithMaxBytes bounds XML input read from readers and files. Non-positive values mean unlimited.
func WithMaxBytes(maxBytes int64) ParseOption { return xmlimpl.WithMaxBytes(maxBytes) }

// WithOpenFile sets the file opener used by XML file read helpers.
func WithOpenFile(openFile func(string) (io.ReadCloser, error)) ParseOption {
	return xmlimpl.WithOpenFile(openFile)
}

// WithDecoderFactory sets the XML decoder factory used by DOM and SAX readers.
func WithDecoderFactory(factory func(io.Reader) *stdxml.Decoder) ParseOption {
	return xmlimpl.WithDecoderFactory(factory)
}

// WithScalarIntParser sets the integer parser used by XML-to-map scalar conversion.
func WithScalarIntParser(parse func(string, int, int) (int64, error)) ParseOption {
	return xmlimpl.WithScalarIntParser(parse)
}

// WithScalarFloatParser sets the float parser used by XML-to-map scalar conversion.
func WithScalarFloatParser(parse func(string, int) (float64, error)) ParseOption {
	return xmlimpl.WithScalarFloatParser(parse)
}

// WithCharset sets the XML declaration charset.
func WithCharset(s string) WriteOption { return xmlimpl.WithCharset(s) }

// WithIndent sets the indentation width in spaces (0 disables pretty printing).
func WithIndent(n int) WriteOption { return xmlimpl.WithIndent(n) }

// WithPretty enables pretty printing with the default indentation.
func WithPretty() WriteOption { return xmlimpl.WithPretty() }

// WithOmitDeclaration controls whether the <?xml ... ?> prolog is emitted.
func WithOmitDeclaration(b bool) WriteOption { return xmlimpl.WithOmitDeclaration(b) }

// WithIgnoreNullFields skips struct fields whose value is a typed nil.
func WithIgnoreNullFields(b bool) WriteOption { return xmlimpl.WithIgnoreNullFields(b) }

// WithRootName overrides the synthesized root element name for MarshalMap / MarshalBean.
func WithRootName(s string) WriteOption { return xmlimpl.WithRootName(s) }

// WithNamespace sets the xmlns attribute on the synthesized root element.
func WithNamespace(s string) WriteOption { return xmlimpl.WithNamespace(s) }

// WithFilePerm sets the file permission used by WriteFile.
func WithFilePerm(perm fs.FileMode) WriteOption { return xmlimpl.WithFilePerm(perm) }

// WithDirPerm sets the parent-directory permission used by WriteFile.
func WithDirPerm(perm fs.FileMode) WriteOption { return xmlimpl.WithDirPerm(perm) }

// WithOverwrite controls whether WriteFile may replace an existing file.
func WithOverwrite(overwrite bool) WriteOption { return xmlimpl.WithOverwrite(overwrite) }

// WithCreateParents controls whether WriteFile creates parent directories.
func WithCreateParents(create bool) WriteOption { return xmlimpl.WithCreateParents(create) }

// WithMkdirAll sets the directory creator used by WriteFile.
func WithMkdirAll(mkdirAll func(string, fs.FileMode) error) WriteOption {
	return xmlimpl.WithMkdirAll(mkdirAll)
}

// WithOpenWriteFile sets the file opener used by WriteFile.
func WithOpenWriteFile(openFile func(string, int, fs.FileMode) (io.WriteCloser, error)) WriteOption {
	return xmlimpl.WithOpenWriteFile(openFile)
}

// WithBeanMarshalFunc sets the marshal provider used by XML bean conversion helpers.
func WithBeanMarshalFunc(marshal func(any) ([]byte, error)) BeanOption {
	return xmlimpl.WithBeanMarshalFunc(marshal)
}

// WithBeanUnmarshalFunc sets the unmarshal provider used by XML bean conversion helpers.
func WithBeanUnmarshalFunc(unmarshal func([]byte, any) error) BeanOption {
	return xmlimpl.WithBeanUnmarshalFunc(unmarshal)
}

// WithTransformParseOptions sets parser options used by TransformWithOptions.
func WithTransformParseOptions(opts ...ParseOption) TransformOption {
	return xmlimpl.WithTransformParseOptions(opts...)
}

// WithTransformWriteOptions sets writer options used by TransformWithOptions.
func WithTransformWriteOptions(opts ...WriteOption) TransformOption {
	return xmlimpl.WithTransformWriteOptions(opts...)
}

// WithFormatParseOptions sets parser options used by FormatWithOptions.
func WithFormatParseOptions(opts ...ParseOption) FormatOption {
	return xmlimpl.WithFormatParseOptions(opts...)
}

// WithFormatWriteOptions sets writer options used by FormatWithOptions.
func WithFormatWriteOptions(opts ...WriteOption) FormatOption {
	return xmlimpl.WithFormatWriteOptions(opts...)
}
