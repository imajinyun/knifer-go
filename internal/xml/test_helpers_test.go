package xml

import (
	"errors"
	"io"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
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

type nopWriteCloser struct{ io.Writer }

func (w nopWriteCloser) Close() error { return nil }

func gotOK(got map[string]any, key string, want any) bool {
	return got != nil && got[key] == want
}
