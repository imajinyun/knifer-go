package xml

import (
	stdxml "encoding/xml"
	"errors"
	"io"
	"io/fs"
	"os"
	"reflect"
	"strings"
	"testing"
)

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

func TestWriteFileReturnsCloseError(t *testing.T) {
	closeErr := errors.New("xml close failed")
	err := WriteFile("/virtual/out.xml", CreateXMLWithRoot("root"), WithOmitDeclaration(true),
		WithMkdirAll(func(string, fs.FileMode) error { return nil }),
		WithOpenWriteFile(func(string, int, fs.FileMode) (io.WriteCloser, error) {
			return closeErrorWriteCloser{Writer: io.Discard, err: closeErr}, nil
		}),
	)
	if !errors.Is(err, closeErr) {
		t.Fatalf("WriteFile close error = %v, want close cause", err)
	}
}
