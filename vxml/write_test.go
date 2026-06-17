package vxml

import (
	"io"
	"io/fs"
	"os"
	"strings"
	"testing"
)

func TestFacadeSAXTransformWriteAndFormat(t *testing.T) {
	var out strings.Builder
	if err := TransformWith(strings.NewReader(`<root><a>1</a></root>`), &out, WithOmitDeclaration(true)); err != nil || out.String() != `<root><a>1</a></root>` {
		t.Fatalf("TransformWith facade = %q, %v", out.String(), err)
	}
	out.Reset()
	if err := TransformWithOptions(
		strings.NewReader(`<root><a>1</a></root>`),
		&out,
		WithTransformParseOptions(WithNamespaceAware(false)),
		WithTransformWriteOptions(WithOmitDeclaration(true), WithIndent(2)),
	); err != nil || !strings.Contains(out.String(), "\n  <a>") {
		t.Fatalf("TransformWithOptions facade = %q, %v", out.String(), err)
	}
	formatted, err := Format(`<root><a>1</a></root>`)
	if err != nil || !strings.Contains(formatted, "\n  <a>") {
		t.Fatalf("Format facade = %q, %v", formatted, err)
	}
	formatted, err = FormatWithOptions(
		`<root xmlns:p="urn:p"><p:a>1</p:a></root>`,
		WithFormatParseOptions(WithNamespaceAware(false)),
		WithFormatWriteOptions(WithOmitDeclaration(true), WithIndent(4)),
	)
	if err != nil || !strings.Contains(formatted, "\n    <a>") || strings.Contains(formatted, "<?xml") {
		t.Fatalf("FormatWithOptions facade = %q, %v", formatted, err)
	}
	namespaced, err := MarshalMap(map[string]any{"name": "go"}, WithRootName("user"), WithNamespace("urn:test"), WithOmitDeclaration(true))
	if err != nil || !strings.Contains(namespaced, `xmlns="urn:test"`) {
		t.Fatalf("MarshalMap WithNamespace = %q, %v", namespaced, err)
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

func TestFacadeXMLWriteFileProviderOptions(t *testing.T) {
	var mkdirPath string
	var mkdirPerm fs.FileMode
	var openPath string
	var openFlag int
	var openPerm fs.FileMode
	var written strings.Builder
	err := WriteFile("/virtual/out.xml", CreateXMLWithRoot("root"), WithOmitDeclaration(true),
		WithMkdirAll(func(path string, perm fs.FileMode) error {
			mkdirPath, mkdirPerm = path, perm
			return nil
		}),
		WithOpenWriteFile(func(path string, flag int, perm fs.FileMode) (io.WriteCloser, error) {
			openPath, openFlag, openPerm = path, flag, perm
			return nopWriteCloser{Writer: &written}, nil
		}),
		WithDirPerm(0o700), WithFilePerm(0o600),
	)
	if err != nil {
		t.Fatalf("WriteFile provider: %v", err)
	}
	if mkdirPath != "/virtual" || mkdirPerm != 0o700 || openPath != "/virtual/out.xml" || openPerm != 0o600 || openFlag&os.O_CREATE == 0 || written.String() != `<root/>` {
		t.Fatalf("providers mkdir=%q/%v open=%q flag=%#x perm=%v content=%q", mkdirPath, mkdirPerm, openPath, openFlag, openPerm, written.String())
	}
}

type nopWriteCloser struct{ io.Writer }

func (w nopWriteCloser) Close() error { return nil }
